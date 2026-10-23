package engines

import "context"

// Logger receives progress and log lines.
type Logger func(line string)

// Engine is a text-to-speech backend.
type Engine interface {
	Manifest() *Manifest
	// Installed reports whether the engine is ready to load.
	Installed() bool
	// Install prepares the engine (downloads, environments).
	Install(ctx context.Context, log Logger) error
	// Uninstall removes what Install created.
	Uninstall() error
	// Voices lists the speakers, loading the model if needed. opts
	// holds the load-scope option values.
	Voices(ctx context.Context, opts Values, log Logger) ([]Voice, error)
	// Synthesize writes req.Output as a WAV file.
	Synthesize(ctx context.Context, req Request, log Logger) error
	// Loaded reports whether a model is held in memory.
	Loaded() bool
	// Close releases loaded models and helper processes.
	Close()
}
