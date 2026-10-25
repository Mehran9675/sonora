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

func (t *TTS) Generate(string, Params, Progress) ([]float32, int, error) {
	return nil, 0, errUnavailable
}

// Recognizer is a speech recognition model; in this build it can't be
// created.
type Recognizer struct{}

func NewRecognizer(RecognizerConfig) (*Recognizer, error) { return nil, errUnavailable }

func (r *Recognizer) Recognize([]float32, int) (Recognition, error) {
	return Recognition{}, errUnavailable
}
func (r *Recognizer) Close() {}

// VAD is a voice activity detector; in this build it can't be created.
type VAD struct{}

func NewVAD(string, float32) (*VAD, error)                { return nil, errUnavailable }
func (v *VAD) Split([]float32, func(int, []float32) bool) {}
func (v *VAD) Close()                                     {}
