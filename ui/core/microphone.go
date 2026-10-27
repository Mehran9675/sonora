//go:build !android

package core

import "context"

// RequestMicrophone asks for access to the microphone where the
// platform needs an explicit grant (Android); elsewhere it's a no-op.
func (c *Ctx) RequestMicrophone(context.Context) error { return nil }
