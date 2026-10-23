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
)

const (
	addAPI     = "api"
	addCommand = "command"
	addPlugin  = "plugin"
)

// addDialog creates API and command engines and explains plugins.
type addDialog struct {
	c    *core.Ctx
	open bool
	kind string
	tabs *segmented.Control
	name *input.Input

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
	d.tabs = segmented.New([]segmented.Option{
		{Value: addAPI, Label: "API server"},
		{Value: addCommand, Label: "Command line"},
		{Value: addPlugin, Label: "Python plugin"},
	}, func() string { return d.kind }, func(v string) { d.kind = v })
	d.name = giode.Input.Text("Name, e.g. “Kokoro server”", theme.Field()).Label("Engine name")
	d.forms = map[string]*setupForm{
		addAPI:     newSetupForm(openai.Template("", "")),
		addCommand: newSetupForm(command.Template("", "")),
	}
	outline := giode.Styles{Color: "theme.text", Background: "theme.border", BorderColor: "theme.border", FontSize: 13,
		Padding: giode.SymmetricInset(14, 9), BorderRadius: theme.RadiusSm + 2}
	primary := giode.Styles{BackgroundGradient: theme.Brand(90), Color: "#ffffff", FontSize: 13,
		Padding: giode.SymmetricInset(16, 9), BorderRadius: theme.RadiusSm + 2}
	d.cancel = giode.Button("Cancel", outline).Variant(giode.VariantOutlined).OnClick(func() { d.open = false })
	d.add = giode.Button("Add engine", primary).Icon(icons.Icon(icons.Plus, 15, "")).OnClick(d.submit)
	d.folder = giode.Button("Open plugins folder", outline).Variant(giode.VariantOutlined).
		Icon(icons.Icon(icons.Folder, 15, "")).OnClick(func() { c.Error(paths.Open(paths.Sub("plugins"))) })
	d.reload = giode.Button("Reload engines", primary).Icon(icons.Icon(icons.Refresh, 15, "")).OnClick(func() {
		c.Manager.Reload()
		d.open = false
		c.Success("Engines reloaded")
	})
	overlay.Add(giode.Dialog(&d.open, d.content, dialogStyle()))
	return d
}

func (d *addDialog) show() {
	d.name.SetText("")
	d.open = true
}

func (d *addDialog) submit() {
	name := strings.TrimSpace(d.name.Text())
	kind := engines.KindOpenAI
	if d.kind == addCommand {
		kind = engines.KindCommand
	}
	if name == "" {
		name = map[engines.Kind]string{engines.KindOpenAI: "API engine", engines.KindCommand: "Command engine"}[kind]
	}
	form := d.forms[d.kind]
	values := engines.Values{}
	for k, v := range form.values {
		values[k] = v
	}
	id := d.c.Manager.AddCustom(kind, name, values)
	d.forms[d.kind] = newSetupForm(map[string]*engines.Manifest{
		addAPI: openai.Template("", ""), addCommand: command.Template("", ""),
	}[d.kind])
	d.open = false
	d.c.Success("Added " + name)
	if rt := d.c.Manager.Get(id); rt != nil {
		d.c.Manager.LoadVoices(rt, true)
	}
}

func (d *addDialog) content() giode.Element {
	els := []giode.Element{
		giode.Text("Add an engine", giode.Styles{Color: "theme.text", FontSize: 18, FontWeight: giode.FontWeightBold}),
		d.tabs.Element(),
	}
	switch d.kind {
	case addPlugin:
		els = append(els,
			giode.Text("Any Python text-to-speech model can become a Sonora engine. Put a folder with an engine.json "+
				"and an adapter.py into the plugins folder, then reload. Sonora installs its packages into a private "+
				"environment, just like the built-in engines.", giode.Styles{Color: "theme.secondary", FontSize: 13, LineHeight: 20}),
			giode.Box(giode.Styles{Width: -1, Padding: giode.UniformInset(12), Background: "theme.background", BorderRadius: theme.RadiusSm},
				giode.Text(paths.Sub("plugins"), theme.MonoText)),
			giode.Text("See docs/plugins.md for the format; the built-in engines are complete examples.", theme.Caption),
			giode.Row(giode.Styles{Width: -1, Gap: 8, Justify: giode.JustifyEnd}, d.folder, d.reload),
		)
	default:
		help := "Connect to anything that speaks OpenAI's /v1/audio/speech API: OpenAI, Kokoro-FastAPI, LocalAI, the Audio8 server…"
		if d.kind == addCommand {
			help = "Run a text-to-speech program installed on this computer. Sonora fills in the placeholders and plays the WAV it writes."
		}
		els = append(els,
			giode.Text(help, giode.Styles{Color: "theme.secondary", FontSize: 13, LineHeight: 20}),
			giode.Stack(giode.Styles{Width: -1, Gap: 8}, giode.Text("Name", theme.Overline), giode.Box(giode.Styles{Width: -1}, d.name)),
		)
		for _, f := range d.forms[d.kind].fields {
			els = append(els, f.Element())
		}
		els = append(els, giode.Row(giode.Styles{Width: -1, Gap: 8, Justify: giode.JustifyEnd}, d.cancel, d.add))
	}
	return giode.Stack(giode.Styles{Width: 520, Gap: 16}, els...)
}

func dialogStyle() giode.Styles {
	return giode.Styles{
		Background: "theme.surface", BorderRadius: theme.RadiusLg, Padding: giode.UniformInset(24),
		BorderWidth: 1, BorderColor: "theme.border",
		BoxShadow: giode.BoxShadow{Y: 24, Blur: 60, Color: "#00000080"},
	}
}
