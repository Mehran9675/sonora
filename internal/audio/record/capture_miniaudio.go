//go:build !windows && !ios

package record

// #include <stdlib.h>
import "C"

import (
	"encoding/binary"
	"errors"
	"fmt"
	"runtime"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	"github.com/gen2brain/malgo"
)

// miniaudio captures through CoreAudio on macOS, PulseAudio, ALSA or
// JACK on Linux, and AAudio or OpenSL ES on Android (the app must hold
// the RECORD_AUDIO permission first).
type miniaudio struct {
	ctx *malgo.AllocatedContext
	dev *malgo.Device
}

// firstData is how long a backend may take to deliver audio before the
// next one is tried.
const firstData = 800 * time.Millisecond

// errNotConnected means the chosen microphone isn't there.
var errNotConnected = errors.New("the microphone isn't connected")

// backends lists what to try, in order. A backend can open fine and
// still deliver nothing (seen with some PulseAudio setups), so each is
// checked for actual audio.
func backends() [][]malgo.Backend {
	if runtime.GOOS == "linux" {
		return [][]malgo.Backend{{malgo.BackendPulseaudio}, {malgo.BackendAlsa}, {malgo.BackendJack}}
	}
	return [][]malgo.Backend{nil} // the platform default
}

// Devices lists the microphones from the first backend that has any.
func Devices() ([]Device, error) {
	var lastErr error
	for _, b := range backends() {
		ctx, err := malgo.InitContext(b, malgo.ContextConfig{}, nil)
		if err != nil {
			lastErr = err
			continue
		}
		infos, err := ctx.Devices(malgo.Capture)
		ctx.Uninit()
		ctx.Free()
		if err != nil {
			lastErr = err
			continue
		}
		var out []Device
		seen := map[string]bool{}
		for i := range infos {
			name := infos[i].Name()
			// PulseAudio lists every output's loopback as an input too.
			if name == "" || seen[name] || strings.HasPrefix(name, "Monitor of ") {
				continue
			}
			seen[name] = true
			out = append(out, Device{ID: name, Name: name})
		}
		if len(out) > 0 {
			return out, nil
		}
	}
	return nil, lastErr
}

func open(device string, push func([]byte)) (capture, error) {
	var lastErr error
	if device != "" {
		bs := backends()
		if strings.HasPrefix(device, AndroidPrefix) {
			bs = [][]malgo.Backend{{malgo.BackendAaudio}}
		}
		for _, b := range bs {
			m, err := openWith(b, device, push)
			if err == nil {
				return m, nil
			}
			lastErr = err
		}
	}
	// No choice, or it's gone: the default microphone.
	for _, b := range backends() {
		m, err := openWith(b, "", push)
		if err == nil {
			return m, nil
		}
		lastErr = err
	}
	return nil, fmt.Errorf("%w (%v)", ErrNoInput, lastErr)
}

func openWith(b []malgo.Backend, device string, push func([]byte)) (*miniaudio, error) {
	ctx, err := malgo.InitContext(b, malgo.ContextConfig{}, nil)
	if err != nil {
		return nil, err
	}
	cfg := malgo.DefaultDeviceConfig(malgo.Capture)
	cfg.Capture.Format = malgo.FormatS16
	cfg.Capture.Channels = 1
	cfg.SampleRate = SampleRate
	if device != "" {
		id, err := findDevice(ctx, device)
		if err != nil {
			ctx.Uninit()
			ctx.Free()
			return nil, err
		}
		ptr := id.Pointer() // C memory; the device copies the ID
		defer C.free(ptr)
		cfg.Capture.DeviceID = ptr
	}
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

// findDevice returns the backend's ID for the microphone device names.
func findDevice(ctx *malgo.AllocatedContext, device string) (*malgo.DeviceID, error) {
	if n, ok := strings.CutPrefix(device, AndroidPrefix); ok {
		v, err := strconv.ParseInt(n, 10, 32)
		if err != nil {
			return nil, err
		}
		// ma_device_id.aaudio is an int32 at the start of the union.
		var id malgo.DeviceID
		binary.NativeEndian.PutUint32(id[:4], uint32(int32(v)))
		return &id, nil
	}
	infos, err := ctx.Devices(malgo.Capture)
	if err != nil {
		return nil, err
	}
	for i := range infos {
		if infos[i].Name() == device {
			id := infos[i].ID
			return &id, nil
		}
	}
	return nil, errNotConnected
}

func (m *miniaudio) stop() {
	m.dev.Stop()
	m.dev.Uninit()
	m.ctx.Uninit()
	m.ctx.Free()
}
