package engines

import (
	"strings"

	"github.com/mehran9675/giode"
	"github.com/mehran9675/giode/components/input"

	"github.com/mehran9675/sonora/internal/engines"
	"github.com/mehran9675/sonora/internal/engines/command"
	"github.com/mehran9675/sonora/internal/engines/openai"
	"github.com/mehran9675/sonora/internal/paths"
	"github.com/mehran9675/sonora/ui/components/optionfield"
	"github.com/mehran9675/sonora/ui/components/overlay"
	"github.com/mehran9675/sonora/ui/components/segmented"
	"github.com/mehran9675/sonora/ui/core"
	"github.com/mehran9675/sonora/ui/icons"
	"github.com/mehran9675/sonora/ui/theme"
	"github.com/mehran9675/sonora/ui/tr"
)

const (
	addAPI        = "api"
	addTranscribe = "transcribe"
	addCommand    = "command"
	addPlugin     = "plugin"
)

// addKinds maps the dialog's tabs to the engines they create; name is
// the default engine name, translated when it's given.
var addKinds = map[string]struct {
	kind     engines.Kind
	name     string
	template func(id, name string) *engines.Manifest
}{
	addAPI:        {engines.KindOpenAI, tr.Msg("API engine"), openai.Template},
	addTranscribe: {engines.KindOpenAITranscribe, tr.Msg("Transcription API"), openai.TranscribeTemplate},
	addCommand:    {engines.KindCommand, tr.Msg("Command engine"), command.Template},
}

// addDialog creates API and command engines and explains plugins.
type addDialog struct {
	c          *core.Ctx
	open       bool
	kind       string
	tabs       *segmented.Control
	mobileTabs *segmented.Control
	name       *input.Input

	forms map[string]*setupForm

	cancel, add, folder, reload *giode.ButtonElement
}

// setupForm is a set of setup fields editing a values map.
type setupForm struct {
	values engines.Values
	fields []*optionfield.Field
}

func newSetupForm(m *engines.Manifest) *setupForm {
	f := &setupForm{values: engines.Resolve(m.Options, nil)}
	st := optionfield.Store{Get: func(id string) any { return f.values[id] }, Set: func(id string, v any) { f.values[id] = v }}
	for _, o := range m.Options {
		if o.Scope == engines.ScopeSetup {
			f.fields = append(f.fields, optionfield.New(o, st))
		}
	}
	return f
}

func newAddDialog(c *core.Ctx) *addDialog {
	d := &addDialog{c: c, kind: addAPI}
	selected, change := func() string { return d.kind }, func(v string) { d.kind = v }
	d.tabs = segmented.New([]segmented.Option{
		{Value: addAPI, Label: tr.Msg("Voice API")},
		{Value: addTranscribe, Label: tr.Msg("Transcription API")},
		{Value: addCommand, Label: tr.Msg("Command line")},
		{Value: addPlugin, Label: tr.Msg("Python plugin")},
	}, selected, change)
	d.mobileTabs = segmented.New([]segmented.Option{
		{Value: addAPI, Label: tr.Msg("Voice API")},
		{Value: addTranscribe, Label: tr.Msg("Transcription API")},
	}, selected, change)
	d.name = giode.Input.Text("", theme.Field())
	tr.Bind(func(s string) { d.name.Placeholder(s) }, "Name, e.g. “Kokoro server”")
	tr.Bind(func(s string) { d.name.Label(s) }, "Engine name")
	d.forms = map[string]*setupForm{}
	for k, a := range addKinds {
		d.forms[k] = newSetupForm(a.template("", ""))
	}
	outline := giode.Styles{Color: "theme.text", Background: "theme.border", BorderColor: "theme.border", FontSize: 13,
		Padding: giode.SymmetricInset(14, 9), BorderRadius: theme.RadiusSm + 2}
	primary := giode.Styles{BackgroundGradient: theme.Brand(90), Color: "#ffffff", FontSize: 13,
		Padding: giode.SymmetricInset(16, 9), BorderRadius: theme.RadiusSm + 2}
	d.cancel = giode.Button("", outline).Variant(giode.VariantOutlined).OnClick(func() { d.open = false })
	d.add = giode.Button("", primary).Icon(icons.Icon(icons.Plus, 15, "")).OnClick(d.submit)
	d.folder = giode.Button("", outline).Variant(giode.VariantOutlined).
		Icon(icons.Icon(icons.Folder, 15, "")).OnClick(func() { c.Error(paths.Open(paths.Sub("plugins"))) })
	d.reload = giode.Button("", primary).Icon(icons.Icon(icons.Refresh, 15, "")).OnClick(func() {
		c.Manager.Reload()
		d.open = false
		c.Success(tr.T("Engines reloaded"))
	})
	tr.Bind(func(s string) { d.cancel.Label(s) }, "Cancel")
	tr.Bind(func(s string) { d.add.Label(s) }, "Add engine")
	tr.Bind(func(s string) { d.folder.Label(s) }, "Open plugins folder")
	tr.Bind(func(s string) { d.reload.Label(s) }, "Reload engines")
	overlay.Add(giode.Dialog(&d.open, d.content, dialogStyle()))
	return d
}

func (d *addDialog) show() {
	d.name.SetText("")
	d.open = true
}

func (d *addDialog) submit() {
	name := strings.TrimSpace(d.name.Text())
	a := addKinds[d.kind]
	if name == "" {
		name = tr.T(a.name)
	}
	form := d.forms[d.kind]
	values := engines.Values{}
	for k, v := range form.values {
		values[k] = v
	}
	id := d.c.Manager.AddCustom(a.kind, name, values)
	d.forms[d.kind] = newSetupForm(a.template("", ""))
	d.open = false
	d.c.Success(tr.T("Added {name}", "name", name))
	if rt := d.c.Manager.Get(id); rt != nil {
		d.c.Manager.LoadVoices(rt, true)
	}
}

func (d *addDialog) content() giode.Element {
	els := []giode.Element{
		giode.Text(tr.T("Add an engine"), giode.Styles{Color: "theme.text", FontSize: 18, FontWeight: giode.FontWeightBold}),
	}
	// Phones can't run local programs or Python, so only API servers
	// are offered there.
	if core.Mobile {
		els = append(els, d.mobileTabs.Element())
	} else {
		els = append(els, d.tabs.Element())
	}
	switch d.kind {
	case addPlugin:
		els = append(els,
			giode.Text(tr.T("Any Python speech model - text to speech or speech to text - can become a Sonora engine. Put a folder with an engine.json and an adapter.py into the plugins folder, then reload. Sonora installs its packages into a private environment, just like the built-in engines."),
				giode.Styles{Color: "theme.secondary", FontSize: 13, LineHeight: 20}),
			giode.Box(giode.Styles{Width: -1, Padding: giode.UniformInset(12), Background: "theme.background", BorderRadius: theme.RadiusSm},
				giode.Text(paths.Sub("plugins"), theme.MonoText)),
			giode.Text(tr.T("See docs/plugins.md for the format; the built-in engines are complete examples."), theme.Caption),
			giode.Row(giode.Styles{Width: -1, Gap: 8, Justify: giode.JustifyEnd}, d.folder, d.reload),
		)
	default:
		help := map[string]string{
			addAPI:        tr.T("Connect to anything that speaks OpenAI's /v1/audio/speech API: OpenAI, Kokoro-FastAPI, LocalAI, the Audio8 server…"),
			addTranscribe: tr.T("Transcribe through anything that speaks OpenAI's /v1/audio/transcriptions API: OpenAI, Groq, a local Whisper server such as Speaches…"),
			addCommand:    tr.T("Run a text-to-speech program installed on this computer. Sonora fills in the placeholders and plays the WAV it writes."),
		}[d.kind]
		els = append(els,
			giode.Text(help, giode.Styles{Color: "theme.secondary", FontSize: 13, LineHeight: 20}),
			giode.Stack(giode.Styles{Width: -1, Gap: 8}, giode.Text(tr.T("Name"), theme.Overline), giode.Box(giode.Styles{Width: -1}, d.name)),
		)
		for _, f := range d.forms[d.kind].fields {
			els = append(els, f.Element())
		}
		els = append(els, giode.Row(giode.Styles{Width: -1, Gap: 8, Justify: giode.JustifyEnd}, d.cancel, d.add))
	}
	return giode.Stack(giode.Styles{Width: -1, MaxWidth: 520, Gap: 16}, els...)
}

func dialogStyle() giode.Styles {
	return giode.Styles{
		Background: "theme.surface", BorderRadius: theme.RadiusLg, Padding: giode.UniformInset(24),
		BorderWidth: 1, BorderColor: "theme.border",
		BoxShadow: giode.BoxShadow{Y: 24, Blur: 60, Color: "#00000080"},
	}
}
