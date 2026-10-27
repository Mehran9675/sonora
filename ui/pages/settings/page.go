package settings

import (
	"context"
	"os"
	"runtime"
	"strconv"
	"strings"

	"github.com/mehran9675/giode"
	"github.com/mehran9675/giode/components/input"
	"github.com/mehran9675/giode/components/scroll"

	"github.com/mehran9675/sonora/internal/audio/convert"
	"github.com/mehran9675/sonora/internal/paths"
	"github.com/mehran9675/sonora/internal/pyenv"
	st "github.com/mehran9675/sonora/internal/settings"
	"github.com/mehran9675/sonora/ui/components/chip"
	"github.com/mehran9675/sonora/ui/components/pagehead"
	"github.com/mehran9675/sonora/ui/components/picker"
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

	quality, limit       *picker.Picker
	qualityT, limitT     *picker.Trigger
	folder               *input.Input
	openFolder, defaults *giode.ButtonElement
}

// New builds the page; applyTheme switches the palette.
func New(c *core.Ctx, applyTheme func(mode string)) *Page {
	p := &Page{c: c}
	if !core.Mobile {
		p.uvInfo = pyenv.UVVersion()
	}
	p.theme = segmented.New([]segmented.Option{{Value: "system", Label: "System"}, {Value: "dark", Label: "Dark"}, {Value: "light", Label: "Light"}},
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
	p.recordingControls()
	p.scroll = giode.Scroll(p.content, theme.Scrolling(6))
	return p
}

// recordingControls builds the Recording section's controls.
func (p *Page) recordingControls() {
	c := p.c
	quality := func() string {
		q := c.Settings.Get().RecordQuality
		if q == "" {
			return core.RecordQualities[1].Value
		}
		return q
	}
	p.quality = picker.New("Recording quality", func() []picker.Item {
		var items []picker.Item
		for _, q := range core.RecordQualities {
			label, detail, _ := strings.Cut(q.Label, " · ")
			items = append(items, picker.Item{ID: q.Value, Label: label, Detail: detail})
		}
		return items
	}, quality, func(it picker.Item) { c.Settings.Update(func(s *st.Settings) { s.RecordQuality = it.ID }) })
	p.qualityT = picker.NewTrigger("Recording quality", func() (string, string) {
		for _, q := range core.RecordQualities {
			if q.Value == quality() {
				label, detail, _ := strings.Cut(q.Label, " · ")
				return label, detail
			}
		}
		return quality(), ""
	}, nil, p.quality.Open)
	minutes := func() string { return strconv.Itoa(int(c.RecordMax().Minutes())) }
	p.limit = picker.New("Longest recording", func() []picker.Item {
		var items []picker.Item
		for _, m := range core.RecordLimits {
			items = append(items, picker.Item{ID: strconv.Itoa(m), Label: limitLabel(m)})
		}
		return items
	}, minutes, func(it picker.Item) {
		m, _ := strconv.Atoi(it.ID)
		c.Settings.Update(func(s *st.Settings) { s.RecordMinutes = m })
	})
	p.limitT = picker.NewTrigger("Longest recording", func() (string, string) {
		m, _ := strconv.Atoi(minutes())
		return limitLabel(m), "Then the recording stops by itself"
	}, nil, p.limit.Open)
	p.folder = giode.Input.Text(core.DefaultRecordingsDir(), theme.Field()).Label("Recordings folder").
		OnChange(func(v string) { c.Settings.Update(func(s *st.Settings) { s.RecordDir = strings.TrimSpace(v) }) })
	p.folder.SetText(c.Settings.Get().RecordDir)
	outline := giode.Styles{Color: "theme.text", Background: "theme.border", BorderColor: "theme.border", FontSize: 13,
		Padding: giode.SymmetricInset(12, 7), BorderRadius: theme.RadiusSm}
	p.openFolder = giode.Button("Open", outline).Variant(giode.VariantOutlined).Icon(icons.Icon(icons.Folder, 14, "")).
		OnClick(func() {
			dir := c.RecordingsDir()
			if err := os.MkdirAll(dir, 0o755); err != nil {
				c.Error(err)
				return
			}
			c.Error(paths.Open(dir))
		})
	p.defaults = giode.Button("Default", outline).Variant(giode.VariantOutlined).OnClick(func() {
		p.folder.SetText("")
		c.Settings.Update(func(s *st.Settings) { s.RecordDir = "" })
	})
}

// limitLabel names a recording length in minutes.
func limitLabel(m int) string {
	switch {
	case m == 60:
		return "1 hour"
	case m%60 == 0:
		return strconv.Itoa(m/60) + " hours"
	}
	return strconv.Itoa(m) + " minutes"
}

// recordingSection is the Recording settings card.
func (p *Page) recordingSection() giode.Element {
	rows := []giode.Element{
		p.setting("Quality", "Recordings are Opus files, a fraction of a WAV's size. Higher quality suits voices you'll clone.",
			p.qualityT.Element()),
		p.setting("Longest recording", "A safety stop, so a forgotten recording doesn't fill the storage.", p.limitT.Element()),
	}
	// Phones keep recordings in Sonora's own storage; export them from
	// the Library.
	if !core.Mobile {
		rows = append(rows, p.setting("Folder", "Where recordings are saved. Leave empty for Sonora's data folder.",
			giode.Stack(giode.Styles{Width: -1, Gap: 8},
				giode.Box(giode.Styles{Width: -1}, p.folder),
				giode.Row(giode.Styles{Gap: 8}, p.openFolder, p.defaults),
			)))
	}
	return section("Recording", "Voices you record to clone or to transcribe.", rows...)
}

func (p *Page) content() giode.Element {
	toolchain := giode.Element(chip.Accent(p.uvInfo, "theme.success"))
	if p.uvInfo == "" {
		toolchain = giode.Row(giode.Styles{Gap: 8, Align: giode.AlignCenter},
			chip.Accent("Downloads on first install", "theme.warning"), p.uv)
	}
	ffmpeg := chip.Accent("Found", "theme.success")
	if !convert.Available() {
		ffmpeg = chip.Accent("Not found — needed for M4A, Opus and video files", "theme.muted")
	}
	sections := []giode.Element{
		pagehead.Head(giode.Compact(), "Settings", "Make Sonora yours."),
		section("Appearance", "How Sonora looks.",
			p.setting("Theme", "System follows your device's light or dark setting.", p.theme.Element()),
		),
		p.recordingSection(),
	}
	// Phones run only on-device models, so the compute device, folders
	// and the Python toolchain don't apply there.
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
				p.setting("FFmpeg", "Reads audio formats Sonora can't decode itself.", ffmpeg),
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
