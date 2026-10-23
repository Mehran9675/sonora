package core

import (
	"errors"
	"strings"

	"github.com/mehran9675/giode"
)

// Info shows a short notification.
func (c *Ctx) Info(msg string) { c.Toasts.Show(giode.ToastInfo, msg) }

// Success shows a confirmation.
func (c *Ctx) Success(msg string) { c.Toasts.Show(giode.ToastSuccess, msg) }

// Error shows err's first line; cancelled dialogs are ignored.
func (c *Ctx) Error(err error) {
	if err == nil || errors.Is(err, giode.ErrCanceled) {
		return
	}
	msg := err.Error()
	if i := strings.IndexByte(msg, '\n'); i > 0 {
		msg = msg[:i]
	}
	if len(msg) > 160 {
		msg = msg[:160] + "…"
	}
	c.Toasts.Show(giode.ToastError, msg)
}

// Background runs fn off the UI goroutine and then done (if any) on it.
func (c *Ctx) Background(fn func() error, done func(error)) {
	go func() {
		err := fn()
		c.App.Post(func() {
			if done != nil {
				done(err)
			} else {
				c.Error(err)
			}
		})
	}()
}
