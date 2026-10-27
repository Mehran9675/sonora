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
	if man.Transcribes() {
		rt.VoicesLoaded = true
		return
	}
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
// Until the user picks a model, the model option names one that's
// downloaded, rather than the manifest's default.
func (m *Manager) values(man *engines.Manifest) engines.Values {
	stored := m.settings.Values(man.ID)
	v := engines.Resolve(man.Options, stored)
	if o := man.ModelOption(); o != nil {
		if _, picked := stored[o.ID]; !picked {
			if id := m.firstDownloaded(man); id != "" {
				v[o.ID] = id
			}
		}
	}
	return v
}

// Values is values for the UI: what a setting shows when the user
// hasn't changed it.
func (m *Manager) Values(man *engines.Manifest) engines.Values { return m.values(man) }

// firstDownloaded returns the first model choice on this device, or "".
func (m *Manager) firstDownloaded(man *engines.Manifest) string {
	have := m.Downloaded(man.ID)
	for _, c := range man.ModelOption().Choices {
		if have[c.Value] {
			return c.Value
		}
	}
	return ""
}

// Downloaded reports which models of the engine with id are on this
// device (nil for engines without a model choice).
func (m *Manager) Downloaded(id string) map[string]bool {
	rt := m.byID[id]
	if rt == nil || rt.Manifest().ModelOption() == nil {
		return nil
	}
	if s, ok := rt.Engine.(engines.ModelStore); ok {
		return s.DownloadedModels()
	}
	return nil
}
