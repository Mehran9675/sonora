package settings

import (
	st "github.com/mehran9675/sonora/internal/settings"
	"github.com/mehran9675/sonora/ui/components/picker"
	"github.com/mehran9675/sonora/ui/core"
	"github.com/mehran9675/sonora/ui/tr"
)

// language is the Appearance section's language picker. Choosing one
// switches the whole app at once.
type language struct {
	picker  *picker.Picker
	trigger *picker.Trigger
}

func (l *language) init(c *core.Ctx) {
	current := func() string { return c.Settings.Get().UILanguage }
	l.picker = picker.New(tr.Msg("Language"), func() []picker.Item {
		items := []picker.Item{{ID: "", Label: tr.T("System"), Detail: tr.Name(tr.System())}}
		for _, lang := range tr.Languages {
			items = append(items, picker.Item{ID: lang.Code, Label: lang.Name})
		}
		return items
	}, current, func(it picker.Item) {
		c.Settings.Update(func(s *st.Settings) { s.UILanguage = it.ID })
		tr.Use(it.ID)
	})
	l.trigger = picker.NewTrigger(tr.Msg("Language"), func() (string, string) {
		if code := current(); code != "" {
			return tr.Name(code), ""
		}
		return tr.T("System"), tr.Name(tr.System())
	}, nil, l.picker.Open)
}
