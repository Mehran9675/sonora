package pyenv

import (
	"context"
	"fmt"
	"io"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/mehran9675/sonora/internal/paths"
)

// Run executes a command, streaming its output lines to log. The error
// carries the last lines of output so failures are self-explanatory.
func Run(ctx context.Context, log Logger, env map[string]string, dir, name string, args ...string) error {
	cmd := exec.CommandContext(ctx, name, args...)
	cmd.Env = Env(env)
	cmd.Dir = dir
	paths.HideWindow(cmd)
	pr, pw := io.Pipe()
	cmd.Stdout, cmd.Stderr = pw, pw
	if err := cmd.Start(); err != nil {
		return err
	}
	var tail []string
	done := make(chan struct{})
	go func() {
		defer close(done)
		ScanLines(pr, func(l string) {
			tail = append(tail, l)
			if len(tail) > 8 {
				tail = tail[1:]
			}
			log(l)
		})
	}()
	err := cmd.Wait()
	pw.Close()
	<-done
	if err != nil {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		return fmt.Errorf("%s\n%s failed: %v\n%s", headline(tail, err), filepath.Base(name), err, strings.Join(tail, "\n"))
	}
	return nil
}

// headline picks the most telling output line - the last one that
// mentions an error - to lead the error message.
func headline(tail []string, err error) string {
	for i := len(tail) - 1; i >= 0; i-- {
		l := strings.ToLower(tail[i])
		if strings.Contains(l, "error") && !strings.HasPrefix(l, "warning") {
			return tail[i]
		}
	}
	return err.Error()
}
