package manager

import (
	"context"

	"github.com/mehran9675/sonora/internal/engines"
)

// LoadVoices fetches the engine's voice list in the background once;
// pass force to refresh it (after changing the model, say).
func (m *Manager) LoadVoices(rt *Runtime, force bool) {
	if rt.VoicesBusy || (rt.VoicesLoaded && !force) || rt.State != StateReady {
		return
	}
	man := rt.Manifest()
	if len(man.Voices) > 0 {
		rt.Voices, rt.VoicesLoaded = man.Voices, true
		return
	}
	rt.VoicesBusy = true
	opts := engines.Subset(man.Options, m.values(man), engines.ScopeLoad)
	sink := m.sink(rt)
	go func() {
		vs, err := rt.Engine.Voices(context.Background(), opts, sink.Log)
		m.post(func() {
			rt.VoicesBusy, rt.VoicesLoaded = false, true
			if err != nil {
				rt.Err = explain(err).Error()
				return
			}
			rt.Voices = vs
		})
	}()
}

// values returns an engine's option values with defaults filled in.
func (m *Manager) values(man *engines.Manifest) engines.Values {
	return engines.Resolve(man.Options, m.settings.Values(man.ID))
}
