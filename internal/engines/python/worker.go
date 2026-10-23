package python

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os/exec"
	"strings"
	"sync"
	"time"

	"github.com/mehran9675/sonora/internal/engines"
	"github.com/mehran9675/sonora/internal/paths"
	"github.com/mehran9675/sonora/internal/pyenv"
)

// worker is a running worker.py process hosting one adapter.
type worker struct {
	cmd   *exec.Cmd
	stdin io.WriteCloser

	mu      sync.Mutex
	nextID  int
	pending map[int]chan message
	log     engines.Logger
	dead    chan struct{}
	exitErr error
	stderr  []string
}

// startWorker launches the worker and waits for its handshake.
func startWorker(ctx context.Context, python, script, adapter, engineDir string, env map[string]string, log engines.Logger) (*worker, error) {
	cmd := exec.Command(python, "-u", script, "--adapter", adapter, "--engine-dir", engineDir)
	cmd.Env = pyenv.Env(env)
	paths.HideWindow(cmd)
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, err
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return nil, err
	}
	w := &worker{cmd: cmd, stdin: stdin, pending: map[int]chan message{}, log: log, dead: make(chan struct{})}
	if err := cmd.Start(); err != nil {
		return nil, err
	}
	ready := make(chan error, 1)
	go pyenv.ScanLines(stderr, w.onStderr)
	go w.readLoop(stdout, ready)
	go func() {
		w.exitErr = cmd.Wait()
		close(w.dead)
	}()
	select {
	case err := <-ready:
		if err != nil {
			w.close()
			return nil, err
		}
		return w, nil
	case <-w.dead:
		return nil, w.failure("the engine stopped while starting")
	case <-ctx.Done():
		w.close()
		return nil, ctx.Err()
	}
}

func (w *worker) onStderr(line string) {
	w.mu.Lock()
	w.stderr = append(w.stderr, line)
	if len(w.stderr) > 12 {
		w.stderr = w.stderr[1:]
	}
	log := w.log
	w.mu.Unlock()
	log(line)
}

func (w *worker) setLog(log engines.Logger) {
	w.mu.Lock()
	w.log = log
	w.mu.Unlock()
}

func (w *worker) readLoop(r io.Reader, ready chan<- error) {
	sc := bufio.NewScanner(r)
	sc.Buffer(make([]byte, 64*1024), 16*1024*1024)
	handshake := false
	for sc.Scan() {
		var m message
		if json.Unmarshal(sc.Bytes(), &m) != nil {
			continue
		}
		switch {
		case m.Event == "ready" && !handshake:
			handshake = true
			ready <- nil
		case m.Event == "fatal" && !handshake:
			handshake = true
			ready <- errors.New(m.Message)
		case m.Event == "log" || m.Event == "progress":
			w.mu.Lock()
			log := w.log
			w.mu.Unlock()
			if m.Event == "progress" {
				log(fmt.Sprintf("%s %d%%", m.Message, int(m.Value*100)))
			} else {
				log(m.Message)
			}
		case m.ID != 0:
			w.mu.Lock()
			ch := w.pending[m.ID]
			delete(w.pending, m.ID)
			w.mu.Unlock()
			if ch != nil {
				ch <- m
			}
		}
	}
}

// call sends a command and waits for its reply. Cancelling ctx kills
// the worker, since a model mid-generation can't be interrupted.
func (w *worker) call(ctx context.Context, cmd string, opts map[string]any, req any) (json.RawMessage, error) {
	w.mu.Lock()
	w.nextID++
	id := w.nextID
	ch := make(chan message, 1)
	w.pending[id] = ch
	w.mu.Unlock()
	line, _ := json.Marshal(request{ID: id, Cmd: cmd, Options: opts, Request: req})
	if _, err := w.stdin.Write(append(line, '\n')); err != nil {
		return nil, w.failure("the engine is not running")
	}
	select {
	case m := <-ch:
		if !m.OK {
			return nil, errors.New(m.Error)
		}
		return m.Result, nil
	case <-w.dead:
		return nil, w.failure("the engine stopped unexpectedly")
	case <-ctx.Done():
		w.close()
		return nil, ctx.Err()
	}
}

func (w *worker) alive() bool {
	select {
	case <-w.dead:
		return false
	default:
		return true
	}
}

// failure builds an error from the last stderr lines, which usually
// hold the Python traceback.
func (w *worker) failure(what string) error {
	w.mu.Lock()
	defer w.mu.Unlock()
	if len(w.stderr) == 0 {
		return errors.New(what)
	}
	return fmt.Errorf("%s:\n%s", what, strings.Join(w.stderr, "\n"))
}

func (w *worker) close() {
	if !w.alive() {
		return
	}
	w.stdin.Write([]byte(`{"cmd":"quit"}` + "\n"))
	w.stdin.Close()
	select {
	case <-w.dead:
	case <-time.After(2 * time.Second):
		w.cmd.Process.Kill()
	}
}
