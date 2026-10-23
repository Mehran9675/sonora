// Package system speaks with the voices built into the operating
// system: SAPI on Windows, `say` on macOS and eSpeak NG on Linux. It
// needs no downloads, so there is always a working engine.
package system

import (
	"context"
	"os"

	"github.com/mehran9675/sonora/internal/engines"
)

// backend is the per-OS implementation.
type backend interface {
	available() bool
	voices(ctx context.Context) ([]engines.Voice, error)
	speak(ctx context.Context, textFile, voice string, speed float64, out string) error
}

// Engine wraps the platform backend.
type Engine struct {
	m *engines.Manifest
	b backend
}

// New returns the system voice engine.
func New() *Engine {
	b, vendor, hint, voice := platform()
	if b.available() {
		hint = ""
	}
	return &Engine{b: b, m: &engines.Manifest{
		ID:           "system",
		Name:         "System Voices",
		Vendor:       vendor,
		Description:  "The voices that come with your operating system. Instant, offline and nothing to download - great for drafts and accessibility.",
		Tags:         []string{"Built in", "Offline", "Instant"},
		License:      "Part of your OS",
		Size:         "Nothing to download",
		Accent:       []string{"#8b5cf6", "#6366f1"},
		Kind:         engines.KindSystem,
		Speed:        true,
		Hint:         hint,
		DefaultVoice: voice,
	}}
}

func (e *Engine) Manifest() *engines.Manifest                   { return e.m }
func (e *Engine) Installed() bool                               { return e.b.available() }
func (e *Engine) Install(context.Context, engines.Logger) error { return nil }
func (e *Engine) Uninstall() error                              { return nil }
func (e *Engine) Close()                                        {}
func (e *Engine) Loaded() bool                                  { return false }

func (e *Engine) Voices(ctx context.Context, _ engines.Values, _ engines.Logger) ([]engines.Voice, error) {
	return e.b.voices(ctx)
}

func (e *Engine) Synthesize(ctx context.Context, r engines.Request, log engines.Logger) error {
	f, err := os.CreateTemp("", "sonora-*.txt")
	if err != nil {
		return err
	}
	defer os.Remove(f.Name())
	f.WriteString(r.Text)
	f.Close()
	speed := r.Speed
	if speed <= 0 {
		speed = 1
	}
	return e.b.speak(ctx, f.Name(), r.Voice, speed, r.Output)
}
