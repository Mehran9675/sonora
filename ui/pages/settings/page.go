package settings

import (
	"context"
	"runtime"

	"github.com/mehran9675/giode"
	"github.com/mehran9675/giode/components/input"
	"github.com/mehran9675/giode/components/scroll"

	"github.com/mehran9675/sonora/internal/audio/convert"
	"github.com/mehran9675/sonora/internal/paths"
	"github.com/mehran9675/sonora/internal/pyenv"
	st "github.com/mehran9675/sonora/internal/settings"
	"github.com/mehran9675/sonora/ui/components/chip"
	"github.com/mehran9675/sonora/ui/components/pagehead"
	"github.com/mehran9675/sonora/ui/components/segmented"
	"github.com/mehran9675/sonora/ui/core"
	"github.com/mehran9675/sonora/ui/icons"
	"github.com/mehran9675/sonora/ui/theme"
)

// Version is set at build time with -ldflags "-X …/settings.Version=…".
var Version = "dev"

// Page is the Settings screen. Create it once.
type Page struct {
	c       *core.Ctx
	scroll  *scroll.Scroll
	theme   *segmented.Control
	device  *segmented.Control
	token   *input.Input
	data    *giode.ButtonElement
	library *giode.ButtonElement
	uv      *giode.ButtonElement
	uvInfo  string
}

// New builds the page; applyTheme switches the palette.
func New(c *core.Ctx, applyTheme func(mode string)) *Page {
	p := &Page{c: c}
	if !core.Mobile {
		p.uvInfo = pyenv.UVVersion()
	}
	p.theme = segmented.New([]segmented.Option{{Value: "dark", Label: "Dark"}, {Value: "light", Label: "Light"}},
		func() string { return c.Settings.Get().Theme }, applyTheme)
	p.device = segmented.New([]segmented.Option{{Value: "auto", Label: "Automatic"}, {Value: "gpu", Label: "GPU"}, {Value: "cpu", Label: "CPU only"}},
		func() string { return c.Settings.Get().Device },
		func(v string) {
			c.Settings.Update(func(s *st.Settings) { s.Device = v })
			c.Manager.CloseIdle() // engines restart on the new device on next use
			c.Info("Engines will restart on the " + v + " device")
		})
	p.token = giode.Input.Text("hf_…", theme.Field()).Password(true).Label("Hugging Face token").
		OnChange(func(v string) { c.Settings.Update(func(s *st.Settings) { s.HFToken = v }) })
	p.token.SetText(c.Settings.Get().HFToken)
	outline := giode.Styles{Color: "theme.text", Background: "theme.border", BorderColor: "theme.border", FontSize: 13,
		Padding: giode.SymmetricInset(14, 8), BorderRadius: theme.RadiusSm}
	p.data = giode.Button("Open data folder", outline).Variant(giode.VariantOutlined).
		Icon(icons.Icon(icons.Folder, 15, "")).OnClick(func() { c.Error(paths.Open(paths.Data())) })
	p.library = giode.Button("Open library folder", outline).Variant(giode.VariantOutlined).
		Icon(icons.Icon(icons.Folder, 15, "")).OnClick(func() { c.Error(paths.Open(c.Library.Dir())) })
	p.uv = giode.Button("Set up now", outline).Variant(giode.VariantOutlined).Icon(icons.Icon(icons.Download, 15, ""))
	p.uv.OnClick(func() {
		p.uv.Loading(true)
		c.Background(func() error {
			_, err := pyenv.EnsureUV(context.Background(), func(string) {})
			return err
		}, func(err error) {
			p.uv.Loading(false)
			if err != nil {
				c.Error(err)
				return
			}
			p.uvInfo = pyenv.UVVersion()
			c.Success("Python toolchain ready")
		})
	})
	p.scroll = giode.Scroll(p.content, giode.Styles{ScrollBar: giode.ScrollBar{ThumbColor: "theme.text/0.15", Width: 6}})
	return p
}

func (p *Page) content() giode.Element {
	toolchain := giode.Element(chip.Accent(p.uvInfo, "theme.success"))
	if p.uvInfo == "" {
		toolchain = giode.Row(giode.Styles{Gap: 8, Align: giode.AlignCenter},
			chip.Accent("Downloads on first install", "theme.warning"), p.uv)
	}
	ffmpeg := chip.Accent("Found", "theme.success")
	if !convert.Available() {
		ffmpeg = chip.Accent("Not found — only needed for MP3 from API servers", "theme.muted")
	}
	sections := []giode.Element{
		pagehead.Head(giode.Compact(), "Settings", "Make Sonora yours."),
		section("Appearance", "How Sonora looks.",
			p.setting("Theme", "Dark is easy on the eyes; light suits bright rooms.", p.theme.Element()),
		),
	}
	// Models run locally only on desktops; phones use API servers, so
	// devices, folders and the Python toolchain don't apply there.
	if !core.Mobile {
		sections = append(sections,
			section("Performance", "Where models run. GPUs are much faster for large models.",
				p.setting("Compute device", "Automatic uses CUDA or Apple silicon when available.", p.device.Element()),
				p.setting("Hugging Face token", "Only for gated or private models.", giode.Box(giode.Styles{Width: -1}, p.token)),
			),
			section("Storage", "Engines, models and your clips live here.",
				p.setting("Data folder", paths.Data(), p.data),
				p.setting("Library", p.c.Library.Dir(), p.library),
			),
			section("Toolchain", "Sonora manages Python for you; nothing to install by hand.",
				p.setting("Python manager (uv)", "Creates an isolated environment per engine.", toolchain),
				p.setting("FFmpeg", "Converts compressed audio from some servers.", ffmpeg),
			),
		)
	}
	sections = append(sections,
		section("About", "Sonora "+Version+" · "+runtime.GOOS+"/"+runtime.GOARCH,
			giode.Text("Built with giode, a Go UI library on Gio. Engines are open-source projects by their authors; "+
				"check each model's license before commercial use.", giode.Styles{Color: "theme.secondary", FontSize: 13, LineHeight: 20}),
		),
	)
	return giode.Stack(giode.Styles{Width: -1, Gap: 20, Padding: giode.UniformInset(core.Gutter()), MaxWidth: 900}, sections...)
}

// Element lays out the page.
func (p *Page) Element() giode.Element {
	return giode.Box(giode.Styles{Width: -1, Height: -1}, p.scroll)
}
