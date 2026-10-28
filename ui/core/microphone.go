//go:build !android

package core

import (
	"context"

	"github.com/mehran9675/sonora/internal/audio/record"
)

// RequestMicrophone asks for access to the microphone where the
// platform needs an explicit grant (Android); elsewhere it's a no-op.
func (c *Ctx) RequestMicrophone(context.Context) error { return nil }

// Microphones lists the microphones to record from. It can take a
// moment; call it off the UI goroutine.
func Microphones() ([]record.Device, error) { return record.Devices() }
