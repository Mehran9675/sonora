package settings

import (
	"github.com/mehran9675/sonora/internal/audio/record"
	st "github.com/mehran9675/sonora/internal/settings"
	"github.com/mehran9675/sonora/ui/components/picker"
	"github.com/mehran9675/sonora/ui/core"
)

// microphone is the Recording section's microphone picker. The list is
// read in the background when the page is built and again each time
// the picker opens, so newly plugged-in microphones show up.
type microphone struct {
	c       *core.Ctx
	picker  *picker.Picker
	trigger *picker.Trigger
	devices []record.Device
	loaded  bool
}

func (m *microphone) init(c *core.Ctx) {
	m.c = c
	current := func() string { return c.Settings.Get().Microphone }
	m.picker = picker.New("Microphone", m.items, current, func(it picker.Item) {
		c.Settings.Update(func(s *st.Settings) {
			s.Microphone = it.ID
			s.MicrophoneName = ""
			if it.ID != "" {
				s.MicrophoneName = it.Label
			}
		})
	})
	m.trigger = picker.NewTrigger("Microphone", m.label, nil, func() {
		m.refresh()
		m.picker.Open()
	})
	m.refresh()
}

// refresh reads the microphones off the UI goroutine.
func (m *microphone) refresh() {
	var found []record.Device
	m.c.Background(func() error {
		var err error
		found, err = core.Microphones()
		return err
	}, func(error) {
		// A failed listing still leaves the default to choose.
		m.devices, m.loaded = found, true
	})
}

// connected reports whether the chosen microphone is in the list.
func (m *microphone) connected(id string) bool {
	for _, d := range m.devices {
		if d.ID == id {
			return true
		}
	}
	return false
}

func (m *microphone) items() []picker.Item {
	items := []picker.Item{{ID: "", Label: "System default", Detail: "The microphone your system records from"}}
	for _, d := range m.devices {
		items = append(items, picker.Item{ID: d.ID, Label: d.Name})
	}
	s := m.c.Settings.Get()
	if s.Microphone != "" && m.loaded && !m.connected(s.Microphone) {
		items = append(items, picker.Item{ID: s.Microphone, Label: m.name(s), Detail: "Not connected"})
	}
	return items
}

func (m *microphone) label() (string, string) {
	s := m.c.Settings.Get()
	switch {
	case s.Microphone == "":
		return "System default", "The microphone your system records from"
	case m.loaded && !m.connected(s.Microphone):
		return m.name(s), "Not connected · the default is used"
	}
	return m.name(s), ""
}

// name is the chosen microphone's name, kept for when it's unplugged.
func (m *microphone) name(s st.Settings) string {
	for _, d := range m.devices {
		if d.ID == s.Microphone {
			return d.Name
		}
	}
	if s.MicrophoneName != "" {
		return s.MicrophoneName
	}
	return s.Microphone
}
