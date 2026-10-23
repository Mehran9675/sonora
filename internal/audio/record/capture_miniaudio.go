//go:build !windows && !android && !ios

package record

import (
	"fmt"
	"runtime"
	"sync/atomic"
	"time"

	"github.com/gen2brain/malgo"
)

// miniaudio captures through CoreAudio on macOS and PulseAudio, ALSA
// or JACK on Linux.
type miniaudio struct {
	ctx *malgo.AllocatedContext
	dev *malgo.Device
}

// firstData is how long a backend may take to deliver audio before the
// next one is tried.
const firstData = 800 * time.Millisecond

// backends lists what to try, in order. A backend can open fine and
// still deliver nothing (seen with some PulseAudio setups), so each is
// checked for actual audio.
func backends() [][]malgo.Backend {
	if runtime.GOOS == "linux" {
		return [][]malgo.Backend{{malgo.BackendPulseaudio}, {malgo.BackendAlsa}, {malgo.BackendJack}}
	}
	return [][]malgo.Backend{nil} // the platform default
}

func open(push func([]byte)) (capture, error) {
	var lastErr error
	for _, b := range backends() {
		m, err := openWith(b, push)
		if err == nil {
			return m, nil
		}
		lastErr = err
	}
	return nil, fmt.Errorf("%w (%v)", ErrNoInput, lastErr)
}

func openWith(b []malgo.Backend, push func([]byte)) (*miniaudio, error) {
	ctx, err := malgo.InitContext(b, malgo.ContextConfig{}, nil)
	if err != nil {
		return nil, err
	}
	cfg := malgo.DefaultDeviceConfig(malgo.Capture)
	cfg.Capture.Format = malgo.FormatS16
	cfg.Capture.Channels = 1
	cfg.SampleRate = SampleRate
	var got atomic.Bool
	dev, err := malgo.InitDevice(ctx.Context, cfg, malgo.DeviceCallbacks{
		Data: func(_, in []byte, _ uint32) {
			got.Store(true)
			buf := make([]byte, len(in)) // in is only valid during the callback
			copy(buf, in)
			push(buf)
		},
	})
	if err != nil {
		ctx.Uninit()
		ctx.Free()
		return nil, err
	}
	m := &miniaudio{ctx: ctx, dev: dev}
	if err := dev.Start(); err != nil {
		m.stop()
		return nil, err
	}
	for deadline := time.Now().Add(firstData); !got.Load(); time.Sleep(20 * time.Millisecond) {
		if time.Now().After(deadline) {
			m.stop()
			return nil, fmt.Errorf("the %v input delivered no audio", b)
		}
	}
	return m, nil
}

func (m *miniaudio) stop() {
	m.dev.Stop()
	m.dev.Uninit()
	m.ctx.Uninit()
	m.ctx.Free()
}
