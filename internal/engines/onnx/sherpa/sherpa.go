//go:build sherpa

package sherpa

/*
#cgo CFLAGS: -I${SRCDIR}
#cgo LDFLAGS: -lsherpa-onnx-c-api -lonnxruntime
#cgo android,arm64 LDFLAGS: -L${SRCDIR}/../../../../third_party/sherpa-onnx/android/arm64-v8a
#cgo android,arm LDFLAGS: -L${SRCDIR}/../../../../third_party/sherpa-onnx/android/armeabi-v7a
#cgo android,amd64 LDFLAGS: -L${SRCDIR}/../../../../third_party/sherpa-onnx/android/x86_64
#cgo linux,!android,amd64 LDFLAGS: -L${SRCDIR}/../../../../third_party/sherpa-onnx/linux-x64
#include <stdlib.h>
#include <string.h>
#include "c-api.h"

extern int32_t sonoraProgress(float *samples, int32_t n, float p, void *arg);

static const SherpaOnnxGeneratedAudio *generate(const SherpaOnnxOfflineTts *tts,
		const char *text, int32_t sid, float speed, uintptr_t handle) {
	SherpaOnnxGenerationConfig cfg;
	memset(&cfg, 0, sizeof(cfg));
	cfg.sid = sid;
	cfg.speed = speed;
	cfg.silence_scale = 0.2f;
	return SherpaOnnxOfflineTtsGenerateWithConfig(tts, text, &cfg,
		(SherpaOnnxGeneratedAudioProgressCallbackWithArg)sonoraProgress, (void *)handle);
}
*/
import "C"

import (
	"errors"
	"runtime"
	"runtime/cgo"
	"sync"
	"unsafe"
)

// Available reports whether this build can run on-device models.
const Available = true

// TTS is a loaded model. Generate may be called from any goroutine,
// one call at a time.
type TTS struct {
	mu  sync.Mutex
	ptr *C.SherpaOnnxOfflineTts
}

// New loads the model described by cfg.
func New(cfg Config) (*TTS, error) {
	var strs []*C.char
	str := func(s string) *C.char {
		if s == "" {
			return nil
		}
		c := C.CString(s)
		strs = append(strs, c)
		return c
	}
	defer func() {
		for _, c := range strs {
			C.free(unsafe.Pointer(c))
		}
	}()
	threads := cfg.Threads
	if threads <= 0 {
		threads = max(1, min(4, runtime.NumCPU()))
	}
	var c C.SherpaOnnxOfflineTtsConfig
	c.model.num_threads = C.int32_t(threads)
	c.model.provider = str("cpu")
	c.max_num_sentences = 1
	switch cfg.Family {
	case VITS:
		c.model.vits.model = str(cfg.Model)
		c.model.vits.tokens = str(cfg.Tokens)
		c.model.vits.data_dir = str(cfg.DataDir)
		c.model.vits.noise_scale = 0.667
		c.model.vits.noise_scale_w = 0.8
		c.model.vits.length_scale = 1
	case Kitten:
		c.model.kitten.model = str(cfg.Model)
		c.model.kitten.voices = str(cfg.Voices)
		c.model.kitten.tokens = str(cfg.Tokens)
		c.model.kitten.data_dir = str(cfg.DataDir)
		c.model.kitten.length_scale = 1
	case Kokoro:
		c.model.kokoro.model = str(cfg.Model)
		c.model.kokoro.voices = str(cfg.Voices)
		c.model.kokoro.tokens = str(cfg.Tokens)
		c.model.kokoro.data_dir = str(cfg.DataDir)
		c.model.kokoro.lexicon = str(cfg.Lexicon)
		c.model.kokoro.dict_dir = str(cfg.DictDir)
		c.model.kokoro.length_scale = 1
	default:
		return nil, errors.New("unknown model family " + string(cfg.Family))
	}
	ptr := C.SherpaOnnxCreateOfflineTts(&c)
	if ptr == nil {
		return nil, errors.New("the model could not be loaded; its files may be incomplete")
	}
	return &TTS{ptr: ptr}, nil
}

// SampleRate returns the model's output rate.
func (t *TTS) SampleRate() int { return int(C.SherpaOnnxOfflineTtsSampleRate(t.ptr)) }

// NumSpeakers returns how many voices the model has.
func (t *TTS) NumSpeakers() int { return int(C.SherpaOnnxOfflineTtsNumSpeakers(t.ptr)) }

// ErrStopped is returned when progress asked to stop.
var ErrStopped = errors.New("generation stopped")

// Generate speaks text with speaker sid at speed, reporting progress.
func (t *TTS) Generate(text string, sid int, speed float32, progress Progress) ([]float32, int, error) {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.ptr == nil {
		return nil, 0, errors.New("the model was unloaded")
	}
	ct := C.CString(text)
	defer C.free(unsafe.Pointer(ct))
	state := &genState{progress: progress}
	h := cgo.NewHandle(state)
	defer h.Delete()
	audio := C.generate(t.ptr, ct, C.int32_t(sid), C.float(speed), C.uintptr_t(h))
	if state.stopped {
		if audio != nil {
			C.SherpaOnnxDestroyOfflineTtsGeneratedAudio(audio)
		}
		return nil, 0, ErrStopped
	}
	if audio == nil {
		return nil, 0, errors.New("the model produced no audio")
	}
	defer C.SherpaOnnxDestroyOfflineTtsGeneratedAudio(audio)
	n := int(audio.n)
	out := make([]float32, n)
	if n > 0 {
		copy(out, unsafe.Slice((*float32)(unsafe.Pointer(audio.samples)), n))
	}
	return out, int(audio.sample_rate), nil
}

// Close frees the model. It waits for a running Generate.
func (t *TTS) Close() {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.ptr != nil {
		C.SherpaOnnxDestroyOfflineTts(t.ptr)
		t.ptr = nil
	}
}

type genState struct {
	progress Progress
	stopped  bool
}

//export sonoraProgress
func sonoraProgress(samples *C.float, n C.int32_t, p C.float, arg unsafe.Pointer) C.int32_t {
	s := cgo.Handle(uintptr(arg)).Value().(*genState)
	if s.progress != nil && !s.progress(float32(p)) {
		s.stopped = true
		return 0
	}
	return 1
}
