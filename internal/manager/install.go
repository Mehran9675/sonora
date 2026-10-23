package manager

import (
	"context"
	"errors"
)

// Install runs the engine's installer in the background.
func (m *Manager) Install(rt *Runtime, done func(error)) {
	if rt.Busy() {
		return
	}
	ctx, cancel := context.WithCancel(context.Background())
	rt.cancel = cancel
	rt.State, rt.Err, rt.Log, rt.Progress = StateInstalling, "", nil, -1
	rt.Status = "Preparing…"
	sink := m.sink(rt)
	go func() {
		err := explain(rt.Engine.Install(ctx, sink.Log))
		m.post(func() {
			rt.cancel = nil
			rt.Progress = -1
			if err != nil {
				rt.State = StateNotInstalled
				if errors.Is(err, context.Canceled) {
					rt.Status = "Install cancelled"
				} else {
					rt.Err, rt.Status = err.Error(), "Install failed"
				}
			} else {
				rt.State, rt.Status = StateReady, "Installed"
				rt.VoicesLoaded = false
			}
			if done != nil {
				done(err)
			}
		})
	}()
}

// Cancel stops the engine's running job.
func (m *Manager) Cancel(rt *Runtime) {
	if rt.cancel != nil {
		rt.cancel()
	}
}

// Eject unloads the engine's model and stops its helper process,
// freeing the memory it used. The model loads again on next use.
func (m *Manager) Eject(rt *Runtime) error {
	if rt.Busy() {
		return errors.New("wait for " + rt.Manifest().Name + " to finish first")
	}
	rt.Engine.Close()
	rt.Status = "Model unloaded"
	return nil
}

// Uninstall removes an installed engine.
func (m *Manager) Uninstall(rt *Runtime) error {
	if rt.Busy() {
		return errors.New("the engine is busy")
	}
	if err := rt.Engine.Uninstall(); err != nil {
		return err
	}
	rt.State, rt.Status, rt.Err, rt.Log = initialState(rt.Engine), "", "", nil
	rt.Voices, rt.VoicesLoaded = nil, false
	return nil
}
