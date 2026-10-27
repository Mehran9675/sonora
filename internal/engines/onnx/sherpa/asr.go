//go:build sherpa || android

package sherpa

/*
#include <stdlib.h>
#include "c-api.h"
*/
import "C"

import (
	"errors"
	"runtime"
	"sync"
	"unsafe"
)

// cstrings allocates C strings and frees them together.
type cstrings []*C.char

func (s *cstrings) str(v string) *C.char {
	if v == "" {
		return nil
	}
	c := C.CString(v)
	*s = append(*s, c)
	return c
}

func (s cstrings) free() {
	for _, c := range s {
		C.free(unsafe.Pointer(c))
	}
}

func threads(n int) C.int32_t {
	if n <= 0 {
		n = max(1, min(4, runtime.NumCPU()))
	}
	return C.int32_t(n)
}

// Recognizer is a loaded speech recognition model. Recognize may be
// called from any goroutine, one call at a time.
type Recognizer struct {
	mu  sync.Mutex
	ptr *C.SherpaOnnxOfflineRecognizer
}

// NewRecognizer loads the model described by cfg.
func NewRecognizer(cfg RecognizerConfig) (*Recognizer, error) {
	var s cstrings
	defer s.free()
	var c C.SherpaOnnxOfflineRecognizerConfig
	c.feat_config.sample_rate = 16000
	c.feat_config.feature_dim = 80
	c.decoding_method = s.str("greedy_search")
	m := &c.model_config
	m.tokens = s.str(cfg.Tokens)
	m.num_threads = threads(cfg.Threads)
	m.provider = s.str("cpu")
	switch cfg.Family {
	case Whisper:
		m.whisper.encoder = s.str(cfg.Encoder)
		m.whisper.decoder = s.str(cfg.Decoder)
		m.whisper.language = s.str(cfg.Language)
		m.whisper.task = s.str("transcribe")
		if cfg.Translate {
			m.whisper.task = s.str("translate")
		}
		m.whisper.tail_paddings = -1
	case Transducer:
		m.transducer.encoder = s.str(cfg.Encoder)
		m.transducer.decoder = s.str(cfg.Decoder)
		m.transducer.joiner = s.str(cfg.Joiner)
		m.model_type = s.str(string(Transducer))
	case SenseVoice:
		m.sense_voice.model = s.str(cfg.Model)
		lang := cfg.Language
		if lang == "" {
			lang = "auto"
		}
		m.sense_voice.language = s.str(lang)
		m.sense_voice.use_itn = 1
	default:
		return nil, errors.New("unknown model family " + string(cfg.Family))
	}
	ptr := C.SherpaOnnxCreateOfflineRecognizer(&c)
	if ptr == nil {
		return nil, errors.New("the model could not be loaded; its files may be incomplete")
	}
	return &Recognizer{ptr: ptr}, nil
}

// Recognize transcribes one utterance (Whisper takes at most 30 s).
func (r *Recognizer) Recognize(samples []float32, rate int) (Recognition, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.ptr == nil {
		return Recognition{}, errors.New("the model was unloaded")
	}
	if len(samples) == 0 {
		return Recognition{}, nil
	}
	stream := C.SherpaOnnxCreateOfflineStream(r.ptr)
	if stream == nil {
		return Recognition{}, errors.New("the recognizer could not start")
	}
	defer C.SherpaOnnxDestroyOfflineStream(stream)
	C.SherpaOnnxAcceptWaveformOffline(stream, C.int32_t(rate), (*C.float)(unsafe.Pointer(&samples[0])), C.int32_t(len(samples)))
	C.SherpaOnnxDecodeOfflineStream(r.ptr, stream)
	res := C.SherpaOnnxGetOfflineStreamResult(stream)
	if res == nil {
		return Recognition{}, errors.New("the recognizer returned nothing")
	}
	defer C.SherpaOnnxDestroyOfflineRecognizerResult(res)
	out := Recognition{Text: C.GoString(res.text)}
	if res.lang != nil {
		out.Language = C.GoString(res.lang)
	}
	return out, nil
}

// Close frees the model. It waits for a running Recognize.
func (r *Recognizer) Close() {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.ptr != nil {
		C.SherpaOnnxDestroyOfflineRecognizer(r.ptr)
		r.ptr = nil
	}
}

// VAD is a Silero voice activity detector for 16 kHz audio.
type VAD struct {
	ptr *C.SherpaOnnxVoiceActivityDetector
}

// NewVAD loads silero_vad.onnx from model. Utterances longer than
// maxSpeech seconds are split.
func NewVAD(model string, maxSpeech float32) (*VAD, error) {
	var s cstrings
	defer s.free()
	var c C.SherpaOnnxVadModelConfig
	c.silero_vad.model = s.str(model)
	c.silero_vad.threshold = 0.5
	c.silero_vad.min_silence_duration = 0.4
	c.silero_vad.min_speech_duration = 0.25
	c.silero_vad.window_size = VADWindow
	c.silero_vad.max_speech_duration = C.float(maxSpeech)
	c.sample_rate = 16000
	c.num_threads = 1
	c.provider = s.str("cpu")
	ptr := C.SherpaOnnxCreateVoiceActivityDetector(&c, C.float(maxSpeech+30))
	if ptr == nil {
		return nil, errors.New("the voice activity detector could not be loaded")
	}
	return &VAD{ptr: ptr}, nil
}

// Split feeds 16 kHz samples through the detector and calls fn with
// each stretch of speech and the sample it starts at. fn returning
// false stops early.
func (v *VAD) Split(samples []float32, fn func(start int, speech []float32) bool) {
	C.SherpaOnnxVoiceActivityDetectorReset(v.ptr)
	drain := func() bool {
		for C.SherpaOnnxVoiceActivityDetectorEmpty(v.ptr) == 0 {
			seg := C.SherpaOnnxVoiceActivityDetectorFront(v.ptr)
			speech := make([]float32, int(seg.n))
			copy(speech, unsafe.Slice((*float32)(unsafe.Pointer(seg.samples)), int(seg.n)))
			start := int(seg.start)
			C.SherpaOnnxDestroySpeechSegment(seg)
			C.SherpaOnnxVoiceActivityDetectorPop(v.ptr)
			if !fn(start, speech) {
				return false
			}
		}
		return true
	}
	for i := 0; i+VADWindow <= len(samples); i += VADWindow {
		C.SherpaOnnxVoiceActivityDetectorAcceptWaveform(v.ptr, (*C.float)(unsafe.Pointer(&samples[i])), VADWindow)
		if !drain() {
			return
		}
	}
	C.SherpaOnnxVoiceActivityDetectorFlush(v.ptr)
	drain()
}

// Close frees the detector.
func (v *VAD) Close() {
	if v.ptr != nil {
		C.SherpaOnnxDestroyVoiceActivityDetector(v.ptr)
		v.ptr = nil
	}
}
