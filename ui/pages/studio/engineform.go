package studio

import (
	"io"
	"os"
	"path/filepath"

	"github.com/mehran9675/giode"
	"github.com/mehran9675/giode/components/input"
	"github.com/mehran9675/giode/components/slider"

	"github.com/mehran9675/sonora/internal/engines"
	"github.com/mehran9675/sonora/internal/manager"
	"github.com/mehran9675/sonora/internal/settings"
	"github.com/mehran9675/sonora/internal/sysdlg"
	"github.com/mehran9675/sonora/ui/components/focusarea"
	"github.com/mehran9675/sonora/ui/components/optionfield"
	"github.com/mehran9675/sonora/ui/components/picker"
	"github.com/mehran9675/sonora/ui/components/segmented"
	"github.com/mehran9675/sonora/ui/components/voicerecorder"
	"github.com/mehran9675/sonora/ui/core"
	"github.com/mehran9675/sonora/ui/icons"
	"github.com/mehran9675/sonora/ui/theme"
)

// engineForm holds the per-engine controls of the voice panel.
type engineForm struct {
	c  *core.Ctx
	rt *manager.Runtime

	fields       []*optionfield.Field
	voices       *picker.Picker
	voiceTrigger *picker.Trigger
	speed        *slider.Slider

	// cloneMode is set when the user chose "Clone a voice".
	cloneMode      bool
	modes          *segmented.Control
	reference      string
	refPick        *giode.ButtonElement
	refPlay        *giode.PressableElement
	recorder       *voicerecorder.Recorder
	refClear       *giode.ButtonElement
	transcript     *input.Input
	transcriptArea *focusarea.Area
}

func newEngineForm(c *core.Ctx, rt *manager.Runtime) *engineForm {
	f := &engineForm{c: c, rt: rt}
	m := rt.Manifest()
	store := optionfield.Store{
		Get: func(id string) any { return c.Settings.Values(m.ID)[id] },
		Set: func(id string, v any) {
			c.Settings.SetValue(m.ID, id, v)
			if o := m.Option(id); o != nil && o.Scope == engines.ScopeLoad && len(m.Voices) == 0 {
				rt.VoicesLoaded = false
			}
		},
	}
	for _, o := range m.Options {
		if o.Scope != engines.ScopeSetup {
			f.fields = append(f.fields, optionfield.New(o, store))
		}
	}

	f.voices = picker.New("Choose a voice", f.voiceItems, func() string { return f.voice().ID }, func(it picker.Item) {
		c.Settings.Update(func(s *settings.Settings) { s.Voices[m.ID] = it.ID })
	})
	f.voiceTrigger = picker.NewTrigger("Voice", func() (string, string) {
		if rt.VoicesBusy {
			return "Loading voices…", "Starting " + m.Name
		}
		v := f.voice()
		detail := v.Group
		if v.Description != "" {
			detail = joinNonEmpty(v.Group, v.Description)
		}
		return v.Label(), detail
	}, func() giode.Element {
		return giode.Box(giode.Styles{Width: 30, Height: 30, BorderRadius: 99, Background: "theme.primary/0.15",
			Align: giode.AlignCenter, Justify: giode.JustifyCenter}, icons.Icon(icons.Mic, 15, "theme.primary"))
	}, func() {
		if !rt.VoicesBusy {
			f.voices.Open()
		}
	})

	f.speed = giode.Input.Slider(giode.Styles{Color: "theme.primary", Background: "theme.text/0.12", BorderColor: "theme.primary"}).
		OnChange(func(v float32) {
			c.Settings.Update(func(s *settings.Settings) { s.Speed[m.ID] = sliderToSpeed(v) })
		})
	f.speed.SetValue(speedToSlider(f.speedValue()))

	small := giode.Styles{Color: "theme.text", Background: "theme.border", BorderColor: "theme.border", FontSize: 13,
		Padding: giode.SymmetricInset(12, 7), BorderRadius: theme.RadiusSm}
	f.refPick = giode.Button("Choose file…", small).Variant(giode.VariantOutlined).
		Icon(icons.Icon(icons.Upload, 15, "")).OnClick(f.pickReference)
	f.refClear = giode.Button("", giode.Styles{Color: "theme.muted", Background: "theme.text", Padding: giode.UniformInset(5), BorderRadius: 8}).
		Variant(giode.VariantGhost).Icon(icons.Icon(icons.Close, 14, "")).OnClick(func() { f.reference = "" })
	f.recorder = voicerecorder.New(c, func(path string) {
		f.reference = path
		f.transcript.SetText("")
		c.Success("Recording saved — it will be used as the voice")
	})
	f.refPlay = giode.Pressable(func(s giode.PressState) giode.Element {
		icon := icons.Solid(icons.Play, 12, "theme.primary")
		if c.IsPlaying(f.reference) {
			icon = icons.Solid(icons.Pause, 12, "theme.primary")
		}
		bg := "theme.primary/0.14"
		if s.Hovered {
			bg = "theme.primary/0.26"
		}
		return giode.Box(giode.Styles{Width: 28, Height: 28, BorderRadius: 99, Background: bg,
			Align: giode.AlignCenter, Justify: giode.JustifyCenter}, icon)
	}).Label("Play recording").OnClick(func(giode.ClickEvent) {
		c.PlayPause(f.reference, "Voice sample", filepath.Base(f.reference))
	})
	f.transcript = giode.Input.Text("What is said in the recording", theme.Field()).Label("Reference transcript")
	f.transcriptArea = focusarea.New(f.transcript, giode.Styles{Width: -1})
	f.cloneMode = m.Clone == engines.CloneRequired
	f.modes = segmented.New([]segmented.Option{{Value: "voice", Label: "Built-in voice"}, {Value: "clone", Label: "Clone a voice"}},
		func() string {
			if f.cloneMode {
				return "clone"
			}
			return "voice"
		},
		func(v string) { f.cloneMode = v == "clone" })
	return f
}

func (f *engineForm) voiceList() []engines.Voice {
	if len(f.rt.Voices) > 0 {
		return f.rt.Voices
	}
	return f.rt.Manifest().Voices
}

func (f *engineForm) voiceItems() []picker.Item {
	var out []picker.Item
	for _, v := range f.voiceList() {
		out = append(out, picker.Item{ID: v.ID, Label: v.Label(), Group: v.Group, Detail: v.Description})
	}
	return out
}

// voice returns the chosen voice, defaulting sensibly.
func (f *engineForm) voice() engines.Voice {
	vs := f.voiceList()
	want := f.c.Settings.Get().Voices[f.rt.Manifest().ID]
	if want == "" {
		want = f.rt.Manifest().DefaultVoice
	}
	for _, v := range vs {
		if v.ID == want {
			return v
		}
	}
	if len(vs) > 0 {
		return vs[0]
	}
	return engines.Voice{ID: want, Name: orDefault(want)}
}

func (f *engineForm) speedValue() float64 {
	if s, ok := f.c.Settings.Get().Speed[f.rt.Manifest().ID]; ok && s > 0 {
		return s
	}
	return 1
}

// pickReference asks for a recording and copies it into the engine's
// folder so the library entry keeps working if the original moves.
func (f *engineForm) pickReference() {
	var picked string
	f.c.Background(func() error {
		p, err := sysdlg.OpenFile("Choose a voice recording", sysdlg.Filter{Name: "Audio", Extensions: []string{"wav", "mp3", "flac", "ogg", "m4a"}})
		if err != nil {
			return err
		}
		picked, err = copyReference(p)
		return err
	}, func(err error) {
		if err != nil {
			f.c.Error(err)
			return
		}
		f.reference = picked
	})
}

func copyReference(src string) (string, error) {
	in, err := os.Open(src)
	if err != nil {
		return "", err
	}
	defer in.Close()
	dst := filepath.Join(core.ReferencesDir(), filepath.Base(src))
	out, err := os.Create(dst)
	if err != nil {
		return "", err
	}
	defer out.Close()
	_, err = io.Copy(out, in)
	return dst, err
}

func sliderToSpeed(v float32) float64 { return float64(int((0.5+float64(v)*1.5)*20+0.5)) / 20 }
func speedToSlider(s float64) float32 { return float32((s - 0.5) / 1.5) }

func orDefault(s string) string {
	if s == "" {
		return "Default voice"
	}
	return s
}

func joinNonEmpty(a, b string) string {
	switch {
	case a == "":
		return b
	case b == "":
		return a
	}
	return a + " · " + b
}
