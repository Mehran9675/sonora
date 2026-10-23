// Package audio holds decoded speech audio and the operations the UI
// needs on it.
package audio

import "time"

// Clip is decoded audio: mono float32 samples in [-1, 1].
type Clip struct {
	Samples    []float32
	SampleRate int
}

// Duration returns the clip's length.
func (c *Clip) Duration() time.Duration {
	if c == nil || c.SampleRate == 0 {
		return 0
	}
	return time.Duration(float64(len(c.Samples)) / float64(c.SampleRate) * float64(time.Second))
}
