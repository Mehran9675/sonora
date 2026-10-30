package studio

import (
	"github.com/mehran9675/sonora/internal/engines"
	"github.com/mehran9675/sonora/internal/manager"
	"github.com/mehran9675/sonora/ui/components/picker"
	"github.com/mehran9675/sonora/ui/components/statuspill"
	"github.com/mehran9675/sonora/ui/core"

	"github.com/mehran9675/sonora/ui/tr"
)

// enginePicker lists every text-to-speech engine, ready ones first.
type enginePicker struct {
	*picker.Picker
}

func newEnginePicker(c *core.Ctx, onPick func(id string)) *enginePicker {
	items := func() []picker.Item {
		var ready, other []picker.Item
		for _, rt := range c.Manager.EnginesFor(engines.TaskSpeak) {
			m := rt.Manifest()
			label, _ := statuspill.Describe(rt)
			it := picker.Item{ID: m.ID, Label: m.Name, Detail: joinNonEmpty(core.Parts(m.Vendor), tr.T(label))}
			if rt.State == manager.StateReady || rt.State == manager.StateWorking {
				it.Group = tr.T("Ready to use")
				ready = append(ready, it)
			} else {
				it.Group = tr.T("Available to install")
				other = append(other, it)
			}
		}
		return append(ready, other...)
	}
	selected := func() string { return c.Settings.Get().Engine }
	return &enginePicker{picker.New(tr.Msg("Choose an engine"), items, selected, func(it picker.Item) { onPick(it.ID) })}
}
