// Package playback plays clips through the system audio device.
package playback

import (
	"errors"
	"sync"
	"time"

	"github.com/ebitengine/oto/v3"
)

const (
	deviceRate = 48000
	frameBytes = 8 // stereo float32
)

var (
	deviceOnce sync.Once
	deviceCtx  *oto.Context
	deviceErr  error
)

// device opens the audio output once; oto allows a single context per
// process.
func device() (*oto.Context, error) {
	deviceOnce.Do(func() {
		var ready chan struct{}
		deviceCtx, ready, deviceErr = oto.NewContext(&oto.NewContextOptions{
			SampleRate:      deviceRate,
			ChannelCount:    2,
			Format:          oto.FormatFloat32LE,
			ApplicationName: "Sonora",
		})
		if deviceErr != nil {
			return
		}
		select {
		case <-ready:
		case <-time.After(5 * time.Second):
			deviceErr = errors.New("audio device did not respond")
		}
	})
	return deviceCtx, deviceErr
}
