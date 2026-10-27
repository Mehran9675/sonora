// Package record captures audio from the default microphone - a voice
// to clone, or speech to transcribe - straight into an Ogg Opus file,
// so a recording's length is bounded by the disk, not memory.
package record

import (
	"errors"
	"math"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/mehran9675/sonora/internal/audio/opusfile"
)

// SampleRate is the capture rate, Opus's own.
const SampleRate = 48000

// levelWindow is how many recent loudness values Levels keeps.
const levelWindow = 48

// ErrNoInput means no microphone could be opened.
var ErrNoInput = errors.New("no microphone found, or access to it was denied")

// capture is the platform's input device.
type capture interface {
	stop()
}

// Options say where and how to record.
type Options struct {
	Path    string // the .opus file to write; its folder is created
	Bitrate int    // Opus bitrate in bits per second
}

// Recorder records mono audio into an Opus file. Methods are safe from
// any goroutine.
type Recorder struct {
	path string
	w    *opusfile.Writer
	dev  capture
	done chan error // the writer goroutine's result

	mu       sync.Mutex
	wake     *sync.Cond
	queue    [][]float32 // captured blocks waiting to be encoded
	closed   bool
	captured int64 // samples captured
	levels   []float32
}

// Start opens the default microphone and records into o.Path.
func Start(o Options) (*Recorder, error) {
	r, err := newRecorder(o)
	if err != nil {
		return nil, err
	}
	dev, err := open(r.push)
	if err != nil {
		r.abort()
		return nil, err
	}
	r.dev = dev
	return r, nil
}

// newRecorder creates the file and starts encoding what push delivers.
func newRecorder(o Options) (*Recorder, error) {
	if err := os.MkdirAll(filepath.Dir(o.Path), 0o755); err != nil {
		return nil, err
	}
	w, err := opusfile.Create(o.Path, SampleRate, o.Bitrate)
	if err != nil {
		return nil, err
	}
	r := &Recorder{path: o.Path, w: w, done: make(chan error, 1)}
	r.wake = sync.NewCond(&r.mu)
	go r.encode()
	return r, nil
}

// encode writes queued blocks to the file until the queue is closed.
// The microphone callback only queues, so a slow disk can't make it
// drop audio.
func (r *Recorder) encode() {
	var failed error
	for {
		r.mu.Lock()
		for len(r.queue) == 0 && !r.closed {
			r.wake.Wait()
		}
		blocks, closed := r.queue, r.closed
		r.queue = nil
		r.mu.Unlock()
		for _, b := range blocks {
			if failed == nil {
				failed = r.w.Write(b)
			}
		}
		if closed {
			if err := r.w.Close(); failed == nil {
				failed = err
			}
			r.done <- failed
			return
		}
	}
}

// push receives captured 16-bit little-endian mono PCM from the
// platform device.
func (r *Recorder) push(pcm []byte) {
	n := len(pcm) / 2
	if n == 0 {
		return
	}
	block := make([]float32, n)
	var sum float64
	for i := range block {
		v := float32(int16(uint16(pcm[2*i])|uint16(pcm[2*i+1])<<8)) / 32768
		block[i] = v
		sum += float64(v) * float64(v)
	}
	// Map RMS to a 0..1 meter on a decibel scale (-54 dB .. 0 dB).
	level := float32(0)
	if rms := math.Sqrt(sum / float64(n)); rms > 0 {
		level = float32(max(0, min(1, (20*math.Log10(rms)+54)/54)))
	}
	r.mu.Lock()
	if !r.closed {
		r.queue = append(r.queue, block)
		r.captured += int64(n)
		r.wake.Signal()
	}
	r.levels = append(r.levels, level)
	if len(r.levels) > levelWindow {
		r.levels = r.levels[len(r.levels)-levelWindow:]
	}
	r.mu.Unlock()
}

// Elapsed returns how much audio has been recorded.
func (r *Recorder) Elapsed() time.Duration {
	r.mu.Lock()
	defer r.mu.Unlock()
	return time.Duration(r.captured) * time.Second / SampleRate
}

// Levels returns recent input loudness values in [0, 1], oldest first.
func (r *Recorder) Levels() []float32 {
	r.mu.Lock()
	defer r.mu.Unlock()
	return append([]float32(nil), r.levels...)
}

// Stop ends the recording, finishes the file and returns its path and
// length.
func (r *Recorder) Stop() (string, time.Duration, error) {
	if r.dev != nil {
		r.dev.stop()
	}
	length := r.Elapsed()
	r.mu.Lock()
	r.closed = true
	r.wake.Signal()
	r.mu.Unlock()
	if err := <-r.done; err != nil {
		os.Remove(r.path)
		return "", 0, err
	}
	return r.path, length, nil
}

// Cancel ends the recording and deletes the file.
func (r *Recorder) Cancel() {
	if path, _, err := r.Stop(); err == nil {
		os.Remove(path)
	}
}

// abort stops encoding and deletes the file, before a device opened.
func (r *Recorder) abort() {
	r.dev = nil
	r.Cancel()
}
