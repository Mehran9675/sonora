// Package python runs engines implemented as Python adapters, each in
// its own uv-managed virtual environment and worker process.
package python

import (
	"context"
	"encoding/json"
	"errors"
	"path/filepath"
	"reflect"
	"sync"
	"sync/atomic"

	"github.com/mehran9675/sonora/internal/engines"
	"github.com/mehran9675/sonora/internal/pyenv"
)

// Engine is a Python-backed engine.
type Engine struct {
	m      *engines.Manifest
	dir    string // install folder: venv/, src/, models/
	script string // worker.py
	env    func() engines.Environment

	mu sync.Mutex
	w  *worker
	// current mirrors w for Loaded, which must not wait on mu while a
	// model is loading.
	current atomic.Pointer[worker]
	loaded  engines.Values
	voices  []engines.Voice
}

// New returns the engine for m, installed under dataDir/engines/<id>
// and run by the worker script at script.
func New(m *engines.Manifest, dataDir, script string, env func() engines.Environment) *Engine {
	return &Engine{m: m, dir: filepath.Join(dataDir, "engines", m.ID), script: script, env: env}
}

// Manifest returns the engine's manifest.
func (e *Engine) Manifest() *engines.Manifest { return e.m }

// ensure starts the worker and loads the model with opts, reusing a
// running worker when nothing that needs a reload changed.
func (e *Engine) ensure(ctx context.Context, opts engines.Values, log engines.Logger) (*worker, error) {
	e.mu.Lock()
	defer e.mu.Unlock()
	if e.w != nil && e.w.alive() && reflect.DeepEqual(e.loaded, opts) {
		e.w.setLog(log)
		return e.w, nil
	}
	if e.w != nil {
		e.w.close()
		e.w = nil
	}
	env := map[string]string{}
	for k, v := range e.m.Env {
		env[k] = v
	}
	settings := e.env()
	env["SONORA_DEVICE"] = settings.Device
	if settings.HFToken != "" {
		env["HF_TOKEN"] = settings.HFToken
	}
	log("Starting " + e.m.Name + "…")
	w, err := startWorker(ctx, pyenv.VenvPython(e.venv()), e.script,
		filepath.Join(e.m.Dir, "adapter.py"), e.dir, env, log)
	if err != nil {
		return nil, err
	}
	raw, err := w.call(ctx, "load", opts, nil)
	if err != nil {
		w.close()
		return nil, err
	}
	var res loadResult
	json.Unmarshal(raw, &res)
	e.voices = nil
	for _, v := range res.Voices {
		e.voices = append(e.voices, engines.Voice{ID: v.ID, Name: v.Name, Group: v.Group, Description: v.Description})
	}
	e.w, e.loaded = w, opts
	e.current.Store(w)
	log(e.m.Name + " is ready.")
	return w, nil
}

// Voices returns the manifest's voices, or loads the model and asks
// the adapter for them.
func (e *Engine) Voices(ctx context.Context, opts engines.Values, log engines.Logger) ([]engines.Voice, error) {
	if len(e.m.Voices) > 0 {
		return e.m.Voices, nil
	}
	if _, err := e.ensure(ctx, opts, log); err != nil {
		return nil, err
	}
	e.mu.Lock()
	defer e.mu.Unlock()
	return e.voices, nil
}

// Synthesize runs one request through the adapter.
func (e *Engine) Synthesize(ctx context.Context, req engines.Request, log engines.Logger) error {
	w, err := e.ensure(ctx, engines.Subset(e.m.Options, req.Options, engines.ScopeLoad), log)
	if err != nil {
		return err
	}
	_, err = w.call(ctx, "synthesize", nil, req)
	return err
}

// Transcribe runs one speech-to-text request through the adapter.
func (e *Engine) Transcribe(ctx context.Context, req engines.TranscribeRequest, log engines.Logger) (*engines.Transcript, error) {
	w, err := e.ensure(ctx, engines.Subset(e.m.Options, req.Options, engines.ScopeLoad), log)
	if err != nil {
		return nil, err
	}
	raw, err := w.call(ctx, "transcribe", nil, req)
	if err != nil {
		return nil, err
	}
	var t engines.Transcript
	if err := json.Unmarshal(raw, &t); err != nil {
		return nil, errors.New("the engine returned an unreadable transcript")
	}
	return &t, nil
}

// Loaded reports whether a worker with a loaded model is running.
func (e *Engine) Loaded() bool {
	w := e.current.Load()
	return w != nil && w.alive()
}

// Close stops the worker, freeing the model's memory.
func (e *Engine) Close() {
	if w := e.current.Swap(nil); w != nil {
		// Stop it first, so a Close during a long load or generation
		// doesn't wait for the lock that work holds.
		w.close()
	}
	e.mu.Lock()
	defer e.mu.Unlock()
	if e.w != nil {
		e.w.close()
		e.w = nil
	}
}
