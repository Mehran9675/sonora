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
	"github.com/mehran9675/sonora/ui/tr"
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
	mic                  microphone
	language             language
}

// New builds the page; applyTheme switches the palette.
func New(c *core.Ctx, applyTheme func(mode string)) *Page {
	p := &Page{c: c}
	if !core.Mobile {
		p.uvInfo = pyenv.UVVersion()
	}
	p.theme = segmented.New([]segmented.Option{{Value: "system", Label: tr.Msg("System")}, {Value: "dark", Label: tr.Msg("Dark")}, {Value: "light", Label: tr.Msg("Light")}},
		func() string { return c.Settings.Get().Theme }, applyTheme)
	p.device = segmented.New([]segmented.Option{{Value: "auto", Label: tr.Msg("Automatic")}, {Value: "gpu", Label: tr.Msg("GPU")}, {Value: "cpu", Label: tr.Msg("CPU only")}},
		func() string { return c.Settings.Get().Device },
		func(v string) {
			c.Settings.Update(func(s *st.Settings) { s.Device = v })
			c.Manager.CloseIdle() // engines restart on the new device on next use
			c.Info(map[string]string{
				"auto": tr.T("Engines will restart and pick the device themselves"),
				"gpu":  tr.T("Engines will restart on the GPU"),
				"cpu":  tr.T("Engines will restart on the CPU"),
			}[v])
		})
	p.token = giode.Input.Text("hf_…", theme.Field()).Password(true).
		OnChange(func(v string) { c.Settings.Update(func(s *st.Settings) { s.HFToken = v }) })
	tr.Bind(func(s string) { p.token.Label(s) }, "Hugging Face token")
	p.token.SetText(c.Settings.Get().HFToken)
	outline := giode.Styles{Color: "theme.text", Background: "theme.border", BorderColor: "theme.border", FontSize: 13,
		Padding: giode.SymmetricInset(14, 8), BorderRadius: theme.RadiusSm}
	p.data = giode.Button("", outline).Variant(giode.VariantOutlined).
		Icon(icons.Icon(icons.Folder, 15, "")).OnClick(func() { c.Error(paths.Open(paths.Data())) })
	tr.Bind(func(s string) { p.data.Label(s) }, "Open data folder")
	p.library = giode.Button("", outline).Variant(giode.VariantOutlined).
		Icon(icons.Icon(icons.Folder, 15, "")).OnClick(func() { c.Error(paths.Open(c.Library.Dir())) })
	tr.Bind(func(s string) { p.library.Label(s) }, "Open library folder")
	p.uv = giode.Button("", outline).Variant(giode.VariantOutlined).Icon(icons.Icon(icons.Download, 15, ""))
	tr.Bind(func(s string) { p.uv.Label(s) }, "Set up now")
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
			c.Success(tr.T("Python toolchain ready"))
		})
	})
	p.recordingControls()
	p.mic.init(c)
	p.language.init(c)
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
	p.quality = picker.New(tr.Msg("Recording quality"), func() []picker.Item {
		var items []picker.Item
		for _, q := range core.RecordQualities {
			items = append(items, picker.Item{ID: q.Value, Label: tr.T(q.Name), Detail: q.Detail()})
		}
		return items
	}, quality, func(it picker.Item) { c.Settings.Update(func(s *st.Settings) { s.RecordQuality = it.ID }) })
	p.qualityT = picker.NewTrigger(tr.Msg("Recording quality"), func() (string, string) {
		for _, q := range core.RecordQualities {
			if q.Value == quality() {
				return tr.T(q.Name), q.Detail()
			}
		}
		return quality(), ""
	}, nil, p.quality.Open)
	minutes := func() string { return strconv.Itoa(int(c.RecordMax().Minutes())) }
	p.limit = picker.New(tr.Msg("Longest recording"), func() []picker.Item {
		var items []picker.Item
		for _, m := range core.RecordLimits {
			items = append(items, picker.Item{ID: strconv.Itoa(m), Label: limitLabel(m)})
		}
		return items
	}, minutes, func(it picker.Item) {
		m, _ := strconv.Atoi(it.ID)
		c.Settings.Update(func(s *st.Settings) { s.RecordMinutes = m })
	})
	p.limitT = picker.NewTrigger(tr.Msg("Longest recording"), func() (string, string) {
		m, _ := strconv.Atoi(minutes())
		return limitLabel(m), tr.T("Then the recording stops by itself")
	}, nil, p.limit.Open)
	p.folder = giode.Input.Text(core.DefaultRecordingsDir(), theme.Field()).
		OnChange(func(v string) { c.Settings.Update(func(s *st.Settings) { s.RecordDir = strings.TrimSpace(v) }) })
	tr.Bind(func(s string) { p.folder.Label(s) }, "Recordings folder")
	p.folder.SetText(c.Settings.Get().RecordDir)
	outline := giode.Styles{Color: "theme.text", Background: "theme.border", BorderColor: "theme.border", FontSize: 13,
		Padding: giode.SymmetricInset(12, 7), BorderRadius: theme.RadiusSm}
	p.openFolder = giode.Button("", outline).Variant(giode.VariantOutlined).Icon(icons.Icon(icons.Folder, 14, "")).
		OnClick(func() {
			dir := c.RecordingsDir()
			if err := os.MkdirAll(dir, 0o755); err != nil {
				c.Error(err)
				return
			}
			c.Error(paths.Open(dir))
		})
	tr.Bind(func(s string) { p.openFolder.Label(s) }, "Open")
	p.defaults = giode.Button("", outline).Variant(giode.VariantOutlined).OnClick(func() {
		p.folder.SetText("")
		c.Settings.Update(func(s *st.Settings) { s.RecordDir = "" })
	})
	tr.Bind(func(s string) { p.defaults.Label(s) }, "Default")
}

// limitLabel names a recording length in minutes.
func limitLabel(m int) string {
	switch {
	case m == 60:
		return tr.T("1 hour")
	case m%60 == 0:
		return tr.N("{n} hours", m/60)
	}
	return tr.N("{n} minutes", m)
}

// recordingSection is the Recording settings card.
func (p *Page) recordingSection() giode.Element {
	rows := []giode.Element{
		p.setting(tr.T("Microphone"), tr.T("Where recordings come from. An unplugged microphone falls back to the default."),
			p.mic.trigger.Element()),
		p.setting(tr.T("Quality"), tr.T("Recordings are Opus files, a fraction of a WAV's size. Higher quality suits voices you'll clone."),
			p.qualityT.Element()),
		p.setting(tr.T("Longest recording"), tr.T("A safety stop, so a forgotten recording doesn't fill the storage."), p.limitT.Element()),
	}
	// Phones keep recordings in Sonora's own storage; export them from
	// the Library.
	if !core.Mobile {
		rows = append(rows, p.setting(tr.T("Folder"), tr.T("Where recordings are saved. Leave empty for Sonora's data folder."),
			giode.Stack(giode.Styles{Width: -1, Gap: 8},
				giode.Box(giode.Styles{Width: -1}, p.folder),
				giode.Row(giode.Styles{Gap: 8}, p.openFolder, p.defaults),
			)))
	}
	return section(tr.T("Recording"), tr.T("Voices you record to clone or to transcribe."), rows...)
}

func (p *Page) content() giode.Element {
	toolchain := giode.Element(chip.Accent(p.uvInfo, "theme.success"))
	if p.uvInfo == "" {
		toolchain = giode.Row(giode.Styles{Gap: 8, Align: giode.AlignCenter},
			chip.Accent(tr.T("Downloads on first install"), "theme.warning"), p.uv)
	}
	ffmpeg := chip.Accent(tr.T("Found"), "theme.success")
	if !convert.Available() {
		ffmpeg = chip.Accent(tr.T("Not found — needed for M4A, Opus and video files"), "theme.muted")
	}
	sections := []giode.Element{
		pagehead.Head(giode.Compact(), tr.T("Settings"), tr.T("Make Sonora yours.")),
		section(tr.T("Appearance"), tr.T("How Sonora looks."),
			p.setting(tr.T("Theme"), tr.T("System follows your device's light or dark setting."), p.theme.Element()),
			p.setting(tr.T("Language"), tr.T("System uses your device's language when Sonora has it."), p.language.trigger.Element()),
		),
		p.recordingSection(),
	}
	// Phones run only on-device models, so the compute device, folders
	// and the Python toolchain don't apply there.
	if !core.Mobile {
		sections = append(sections,
			section(tr.T("Performance"), tr.T("Where models run. GPUs are much faster for large models."),
				p.setting(tr.T("Compute device"), tr.T("Automatic uses CUDA or Apple silicon when available."), p.device.Element()),
				p.setting(tr.T("Hugging Face token"), tr.T("Only for gated or private models."), giode.Box(giode.Styles{Width: -1}, p.token)),
			),
			section(tr.T("Storage"), tr.T("Engines, models and your clips live here."),
				p.setting(tr.T("Data folder"), paths.Data(), p.data),
				p.setting(tr.T("Library"), p.c.Library.Dir(), p.library),
			),
			section(tr.T("Toolchain"), tr.T("Sonora manages Python for you; nothing to install by hand."),
				p.setting(tr.T("Python manager (uv)"), tr.T("Creates an isolated environment per engine."), toolchain),
				p.setting(tr.T("FFmpeg"), tr.T("Reads audio formats Sonora can't decode itself."), ffmpeg),
			),
		)
	}
	sections = append(sections,
		section(tr.T("About"), "Sonora "+Version+" · "+runtime.GOOS+"/"+runtime.GOARCH,
			giode.Text(tr.T("Built with giode, a Go UI library on Gio. Engines are open-source projects by their authors; "+
				"check each model's license before commercial use."), giode.Styles{Color: "theme.secondary", FontSize: 13, LineHeight: 20}),
		),
	)
	return giode.Stack(giode.Styles{Width: -1, Gap: 20, Padding: giode.UniformInset(core.Gutter()), MaxWidth: 900}, sections...)
}

// Element lays out the page.
func (p *Page) Element() giode.Element {
	return giode.Box(giode.Styles{Width: -1, Height: -1}, p.scroll)
}
