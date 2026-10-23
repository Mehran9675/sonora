//go:build !sherpa

package sherpa

import "errors"

// Available reports whether this build can run on-device models.
const Available = false

var errUnavailable = errors.New("this build has no on-device engine support")

// ErrStopped is returned when progress asked to stop.
var ErrStopped = errors.New("generation stopped")

// TTS is a loaded model; in this build it can't be created.
type TTS struct{}

func New(Config) (*TTS, error) { return nil, errUnavailable }

func (t *TTS) SampleRate() int  { return 0 }
func (t *TTS) NumSpeakers() int { return 0 }
func (t *TTS) Close()           {}

func (t *TTS) Generate(string, int, float32, Progress) ([]float32, int, error) {
	return nil, 0, errUnavailable
}
