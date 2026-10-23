// Package manager connects engines to the UI: it builds the engine
// list, runs installs, voice loading and generation in the background,
// and publishes their progress as Runtime values on the UI goroutine.
package manager

import (
	"fmt"
	"log"
	"path/filepath"
	"time"

	"github.com/mehran9675/sonora/internal/catalog"
	"github.com/mehran9675/sonora/internal/engines"
	"github.com/mehran9675/sonora/internal/engines/command"
	"github.com/mehran9675/sonora/internal/engines/openai"
	"github.com/mehran9675/sonora/internal/engines/python"
	"github.com/mehran9675/sonora/internal/engines/system"
	"github.com/mehran9675/sonora/internal/library"
	"github.com/mehran9675/sonora/internal/paths"
	"github.com/mehran9675/sonora/internal/settings"
)

// Manager owns every engine's Runtime.
type Manager struct {
	post     func(func())
	settings *settings.Store
	library  *library.Library
	script   string

	list []*Runtime
	byID map[string]*Runtime

	// Job is the generation in progress, if any.
	Job *Job
}

// New unpacks the Python runtime and builds the engine list. post must
// run a function on the UI goroutine (giode's App.Post).
func New(post func(func()), st *settings.Store, lib *library.Library) *Manager {
	m := &Manager{post: post, settings: st, library: lib, byID: map[string]*Runtime{}}
	if !mobile {
		script, err := catalog.Unpack(paths.Sub("runtime"))
		if err != nil {
			log.Printf("unpack runtime: %v", err)
		}
		m.script = script
	}
	m.Reload()
	return m
}

// Reload rebuilds the engine list (after adding or removing engines),
// keeping the runtimes of engines that still exist.
func (m *Manager) Reload() {
	old := m.byID
	m.list, m.byID = nil, map[string]*Runtime{}
	add := func(e engines.Engine) {
		id := e.Manifest().ID
		if _, dup := m.byID[id]; dup {
			return
		}
		rt := old[id]
		if rt == nil {
			rt = &Runtime{Engine: e, Progress: -1}
			rt.State = initialState(e)
		}
		m.list = append(m.list, rt)
		m.byID[id] = rt
	}
	// Phones can't run local programs or Python: only API engines.
	if !mobile {
		add(system.New())
		builtin, errs := catalog.LoadDir(filepath.Join(paths.Data(), "runtime", "engines"))
		plugins, perrs := catalog.LoadDir(paths.Sub("plugins"))
		for _, err := range append(errs, perrs...) {
			log.Printf("engine manifest: %v", err)
		}
		for _, man := range append(builtin, plugins...) {
			add(python.New(man, paths.Data(), m.script, m.environment))
		}
	}
	for _, man := range m.settings.Get().Custom {
		id := man.ID
		cfg := func() engines.Values { return engines.Resolve(m.manifest(id).Options, m.settings.Values(id)) }
		switch man.Kind {
		case engines.KindOpenAI:
			add(openai.New(withOptions(man, openai.Template), cfg))
		case engines.KindCommand:
			if mobile {
				continue
			}
			add(command.New(withOptions(man, command.Template), cfg))
		}
	}
}

// withOptions refreshes a stored custom manifest's option definitions
// from the current template, keeping the user's name.
func withOptions(stored *engines.Manifest, tmpl func(id, name string) *engines.Manifest) *engines.Manifest {
	return tmpl(stored.ID, stored.Name)
}

func initialState(e engines.Engine) State {
	switch {
	case e.Installed():
		return StateReady
	case e.Manifest().NeedsInstall():
		return StateNotInstalled
	default:
		return StateUnavailable
	}
}

func (m *Manager) environment() engines.Environment {
	s := m.settings.Get()
	return engines.Environment{Device: s.Device, HFToken: s.HFToken}
}

func (m *Manager) manifest(id string) *engines.Manifest {
	if rt := m.byID[id]; rt != nil {
		return rt.Manifest()
	}
	return &engines.Manifest{}
}

// Engines returns every engine in display order.
func (m *Manager) Engines() []*Runtime { return m.list }

// Get returns the engine with id, or nil.
func (m *Manager) Get(id string) *Runtime { return m.byID[id] }

// Ready returns the engines that can generate now.
func (m *Manager) Ready() []*Runtime {
	var out []*Runtime
	for _, rt := range m.list {
		if rt.State == StateReady || rt.State == StateWorking {
			out = append(out, rt)
		}
	}
	return out
}

// AddCustom stores a new API or command engine and returns its id.
func (m *Manager) AddCustom(kind engines.Kind, name string, values engines.Values) string {
	id := fmt.Sprintf("%s-%d", kind, time.Now().UnixNano())
	m.settings.Update(func(s *settings.Settings) {
		s.Custom = append(s.Custom, &engines.Manifest{ID: id, Name: name, Kind: kind, Custom: true})
		s.Values[id] = values
	})
	m.Reload()
	return id
}

// RemoveCustom deletes a user-added engine.
func (m *Manager) RemoveCustom(id string) {
	m.settings.Update(func(s *settings.Settings) {
		for i, c := range s.Custom {
			if c.ID == id {
				s.Custom = append(s.Custom[:i], s.Custom[i+1:]...)
				break
			}
		}
		delete(s.Values, id)
	})
	m.Reload()
}

// Shutdown stops every engine's helper process.
func (m *Manager) Shutdown() {
	for _, rt := range m.list {
		if rt.cancel != nil {
			rt.cancel()
		}
		rt.Engine.Close()
	}
}

// CloseIdle stops the helper processes of engines that aren't busy, so
// they restart with new settings on next use.
func (m *Manager) CloseIdle() {
	for _, rt := range m.list {
		if !rt.Busy() {
			rt.Engine.Close()
		}
	}
}
