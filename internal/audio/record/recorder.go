// Package record captures audio from the default microphone, for
// recording a voice to clone.
package record

import (
	"errors"
	"math"
	"sync"
	"time"

	"github.com/mehran9675/sonora/internal/audio"
)

// SampleRate is the capture rate; 44.1 kHz suits every cloning model.
const SampleRate = 44100

// levelWindow is how many recent loudness values Levels keeps.
const levelWindow = 48

// ErrNoInput means no microphone could be opened.
var ErrNoInput = errors.New("no microphone found, or access to it was denied")

// capture is the platform's input device.
type capture interface {
	stop()
}

// Recorder records mono 16-bit audio. Methods are safe from any
// goroutine.
type Recorder struct {
	mu      sync.Mutex
	samples []float32
	levels  []float32
	started time.Time
	dev     capture
}

// Start opens the default microphone and begins recording.
func Start() (*Recorder, error) {
	r := &Recorder{started: time.Now()}
	dev, err := open(r.push)
	if err != nil {
		return nil, err
	}
	r.dev = dev
	return r, nil
}

// push receives captured 16-bit little-endian mono PCM from the
// platform device.
func (r *Recorder) push(pcm []byte) {
	n := len(pcm) / 2
	if n == 0 {
		return
	}
	chunk := make([]float32, n)
	var sum float64
	for i := range chunk {
		v := float32(int16(uint16(pcm[2*i])|uint16(pcm[2*i+1])<<8)) / 32768
		chunk[i] = v
		sum += float64(v) * float64(v)
	}
	// Map RMS to a 0..1 meter on a decibel scale (-54 dB .. 0 dB).
	level := float32(0)
	if rms := math.Sqrt(sum / float64(n)); rms > 0 {
		level = float32(max(0, min(1, (20*math.Log10(rms)+54)/54)))
	}
	r.mu.Lock()
	r.samples = append(r.samples, chunk...)
	r.levels = append(r.levels, level)
	if len(r.levels) > levelWindow {
		r.levels = r.levels[len(r.levels)-levelWindow:]
	}
	r.mu.Unlock()
}

// Elapsed returns how long the recorder has been running.
func (r *Recorder) Elapsed() time.Duration { return time.Since(r.started) }

// Levels returns recent input loudness values in [0, 1], oldest first.
func (r *Recorder) Levels() []float32 {
	r.mu.Lock()
	defer r.mu.Unlock()
	return append([]float32(nil), r.levels...)
}

// Stop ends the recording and returns it with leading and trailing
// silence trimmed.
func (r *Recorder) Stop() *audio.Clip {
	r.dev.stop()
	r.mu.Lock()
	defer r.mu.Unlock()
	return &audio.Clip{Samples: trimSilence(r.samples, SampleRate), SampleRate: SampleRate}
}

// trimSilence drops quiet stretches at both ends, keeping a short pad
// so words aren't clipped.
func trimSilence(s []float32, rate int) []float32 {
	const threshold = 0.02
	pad := rate / 10
	first, last := -1, -1
	for i, v := range s {
		if v > threshold || v < -threshold {
			if first < 0 {
				first = i
			}
			last = i
		}
	}
	if first < 0 {
		return s
	}
	return s[max(0, first-pad):min(len(s), last+pad)]
}
