package manager

import (
	"context"

	"github.com/mehran9675/sonora/internal/engines"
)

// State is an engine's lifecycle stage as shown in the UI.
type State int

const (
	// StateNotInstalled: the engine needs its one-click install.
	StateNotInstalled State = iota
	// StateUnavailable: a system dependency is missing (see Hint).
	StateUnavailable
	// StateInstalling: Install is running.
	StateInstalling
	// StateReady: the engine can generate.
	StateReady
	// StateWorking: the engine is loading a model or generating.
	StateWorking
)

// Runtime is the UI-facing state of one engine. It is only changed on
// the UI goroutine (through the manager's post function), so views can
// read it freely.
type Runtime struct {
	Engine   engines.Engine
	State    State
	Status   string  // latest log line, for status bars
	Progress float32 // latest percentage in [0, 1], or -1 when unknown
	Log      []string
	Err      string // last failure

	Voices       []engines.Voice
	VoicesLoaded bool
	VoicesBusy   bool

	cancel context.CancelFunc
}

// Manifest is a shortcut for r.Engine.Manifest().
func (r *Runtime) Manifest() *engines.Manifest { return r.Engine.Manifest() }

// Busy reports whether a job is running on the engine.
func (r *Runtime) Busy() bool { return r.cancel != nil }
