package core

import (
	"github.com/mehran9675/sonora/internal/engines"
	"github.com/mehran9675/sonora/internal/manager"
	"github.com/mehran9675/sonora/ui/components/optionfield"

	"github.com/mehran9675/sonora/ui/tr"
)

// OptionStore edits engine man's settings for its option fields. Values
// read as the manager resolves them, so an untouched model setting
// shows a model that's downloaded, and model choices say whether
// they're downloaded. onSet (optional) runs after a change.
func (c *Ctx) OptionStore(man *engines.Manifest, onSet func(id string)) optionfield.Store {
	return optionfield.Store{
		Get: func(id string) any { return c.Manager.Values(man)[id] },
		Set: func(id string, v any) {
			c.Settings.SetValue(man.ID, id, v)
			if onSet != nil {
				onSet(id)
			}
		},
		Detail: func(id, value string) string {
			o := man.ModelOption()
			if o == nil || o.ID != id {
				return ""
			}
			if c.Manager.Downloaded(man.ID)[value] {
				return tr.T("Downloaded")
			}
			return tr.T("Not downloaded yet")
		},
	}
}

// InstallModel installs an engine if needed and downloads one of its
// models, keeping the device awake meanwhile; done (optional) runs
// afterwards on the UI goroutine.
func (c *Ctx) InstallModel(rt *manager.Runtime, model string, done func(error)) {
	awake := c.StayAwake(tr.T("Downloading a model for {name}", "name", rt.Manifest().Name))
	c.Manager.InstallModel(rt, model, func(err error) {
		awake()
		if err == nil {
			c.NotifyIfAway(tr.T("{name} is ready", "name", rt.Manifest().Name), tr.T("The model is downloaded."))
		}
		if done != nil {
			done(err)
		}
	})
}
