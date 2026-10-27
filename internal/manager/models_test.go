package manager

import (
	"context"
	"testing"

	"github.com/mehran9675/sonora/internal/engines"
	"github.com/mehran9675/sonora/internal/settings"
)

// modelEngine is an engine with three models, some downloaded.
type modelEngine struct {
	engines.Engine
	m    *engines.Manifest
	have map[string]bool
}

func (e *modelEngine) Manifest() *engines.Manifest       { return e.m }
func (e *modelEngine) Installed() bool                   { return len(e.have) > 0 }
func (e *modelEngine) DownloadedModels() map[string]bool { return e.have }
func (e *modelEngine) InstallModel(_ context.Context, id string, _ engines.Logger) error {
	e.have[id] = true
	return nil
}

func TestModelDefaults(t *testing.T) {
	t.Setenv("SONORA_HOME", t.TempDir())
	man := &engines.Manifest{ID: "e", Kind: engines.KindOnDevice, Options: []engines.Option{{
		ID: "model", Type: engines.OptionSelect, Scope: engines.ScopeLoad, Default: "small",
		Choices: []engines.Choice{{Value: "small"}, {Value: "medium"}, {Value: "large"}},
	}}}
	eng := &modelEngine{m: man, have: map[string]bool{"large": true}}
	rt := &Runtime{Engine: eng, State: StateReady}
	posted := make(chan func(), 10)
	m := &Manager{post: func(f func()) { posted <- f }, settings: settings.Open(), byID: map[string]*Runtime{"e": rt}}

	// Untouched, the setting shows the downloaded model, not the default.
	if got := m.Values(man).String("model"); got != "large" {
		t.Fatalf("default: %q", got)
	}
	// Downloading another from the engine card selects it.
	done := make(chan error, 1)
	m.InstallModel(rt, "medium", func(err error) { done <- err })
	for len(done) == 0 {
		(<-posted)()
	}
	if err := <-done; err != nil {
		t.Fatal(err)
	}
	if got := m.Values(man).String("model"); got != "medium" || rt.State != StateReady {
		t.Fatalf("after download: %q, state %v", got, rt.State)
	}
	// A model picked on the main page sticks, downloaded or not; it
	// downloads when it's used.
	m.settings.SetValue("e", "model", "small")
	if got := m.Values(man).String("model"); got != "small" {
		t.Fatalf("picked: %q", got)
	}
}
