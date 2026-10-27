package core

import (
	"sync/atomic"

	"github.com/mehran9675/sonora/internal/manager"
)

// away is set while the app is in the background (minimized, or behind
// another app on a phone).
var away atomic.Bool

// SetAway records whether the app is in the background; the app's
// OnBackground and OnForeground hooks call it, from any goroutine.
func SetAway(v bool) { away.Store(v) }

// StayAwake keeps the screen on and the app running for long work -
// generating, transcribing, downloading a model - until the returned
// function is called. Calls nest. Call both on the UI goroutine.
//
// On Android, running on behind other apps needs a foreground service
// the APK doesn't declare yet, so there only the screen is kept on.
func (c *Ctx) StayAwake(reason string) (done func()) {
	if c.awake == 0 {
		c.App.KeepScreenOn(true)
		if release, err := c.App.KeepAlive(reason); err == nil {
			c.release = release
		}
	}
	c.awake++
	finished := false
	return func() {
		if finished {
			return
		}
		finished = true
		if c.awake--; c.awake == 0 {
			c.App.KeepScreenOn(false)
			if c.release != nil {
				c.release()
				c.release = nil
			}
		}
	}
}

// NotifyIfAway shows a system notification when the app is in the
// background, so the user learns that long work finished.
func (c *Ctx) NotifyIfAway(title, body string) {
	if away.Load() {
		go c.App.Notify(title, body)
	}
}

// Install installs an engine, keeping the device awake through its
// downloads; done (optional) runs afterwards on the UI goroutine.
func (c *Ctx) Install(rt *manager.Runtime, done func(error)) {
	awake := c.StayAwake("Installing " + rt.Manifest().Name)
	c.Manager.Install(rt, func(err error) {
		awake()
		if err == nil {
			c.NotifyIfAway(rt.Manifest().Name+" is installed", "Ready to use.")
		}
		if done != nil {
			done(err)
		}
	})
}
