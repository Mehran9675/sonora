// Package voicerecorder records the user's voice from the microphone
// for cloning: Record, a live meter and timer, Stop.
package voicerecorder

import (
	"errors"
	"fmt"
	"path/filepath"
	"time"

	"github.com/mehran9675/giode"

	"github.com/mehran9675/sonora/internal/audio/record"
	"github.com/mehran9675/sonora/internal/audio/wav"
	"github.com/mehran9675/sonora/ui/components/levelmeter"
	"github.com/mehran9675/sonora/ui/core"
	"github.com/mehran9675/sonora/ui/icons"
	"github.com/mehran9675/sonora/ui/theme"
)

const (
	// MaxLength stops a recording automatically.
	MaxLength = 30 * time.Second
	// MinLength is the shortest recording worth cloning from.
	MinLength = 5 * time.Second
)

// Recorder is stateful; create one per form.
type Recorder struct {
	c      *core.Ctx
	onDone func(path string)

	rec    *record.Recorder
	timer  *time.Timer
	meter  *levelmeter.Meter
	start  *giode.ButtonElement
	stop   *giode.ButtonElement
	cancel *giode.ButtonElement
}

// New returns a recorder that calls onDone with the saved WAV file.
func New(c *core.Ctx, onDone func(path string)) *Recorder {
	r := &Recorder{c: c, onDone: onDone}
	r.meter = &levelmeter.Meter{Color: "theme.danger", Levels: func() []float32 {
		if r.rec == nil {
			return nil
		}
		return r.rec.Levels()
	}}
	r.start = giode.Button("Record", giode.Styles{Color: "theme.text", Background: "theme.border", BorderColor: "theme.border",
		FontSize: 13, Padding: giode.SymmetricInset(12, 7), BorderRadius: theme.RadiusSm}).
		Variant(giode.VariantOutlined).Icon(icons.Icon(icons.Mic, 15, "")).OnClick(r.begin)
	r.stop = giode.Button("Stop", giode.Styles{Background: "theme.danger", Color: "#ffffff", FontSize: 13,
		Padding: giode.SymmetricInset(12, 7), BorderRadius: theme.RadiusSm}).
		Icon(icons.Solid(icons.Stop, 12, "")).OnClick(func() { r.finish(true) })
	r.cancel = giode.Button("", giode.Styles{Color: "theme.muted", Background: "theme.text", Padding: giode.UniformInset(6), BorderRadius: 8}).
		Variant(giode.VariantGhost).Icon(icons.Icon(icons.Close, 14, "")).OnClick(func() { r.finish(false) })
	return r
}

// Recording reports whether the microphone is open.
func (r *Recorder) Recording() bool { return r.rec != nil }

// Button is the Record button shown while idle.
func (r *Recorder) Button() giode.Element { return r.start }

func (r *Recorder) begin() {
	r.c.Player.Pause() // don't record the speakers
	rec, err := record.Start()
	if err != nil {
		r.c.Error(err)
		return
	}
	r.rec = rec
	r.timer = time.AfterFunc(MaxLength, func() { r.c.App.Post(func() { r.finish(true) }) })
}

// finish closes the microphone and, when keep is set, saves the take.
func (r *Recorder) finish(keep bool) {
	rec := r.rec
	if rec == nil {
		return
	}
	r.rec = nil
	r.timer.Stop()
	elapsed := rec.Elapsed()
	var path string
	r.c.Background(func() error {
		clip := rec.Stop()
		if !keep {
			return nil
		}
		if elapsed < MinLength || clip.Duration() < 3*time.Second {
			return errors.New("That recording is too short; speak for at least 5 seconds")
		}
		path = filepath.Join(core.ReferencesDir(), "recording-"+time.Now().Format("20060102-150405")+".wav")
		return wav.Write(path, clip)
	}, func(err error) {
		if err != nil {
			r.c.Error(err)
			return
		}
		if path != "" {
			r.onDone(path)
		}
	})
}

// Element is the live recording panel; lay it out while Recording.
func (r *Recorder) Element() giode.Element {
	if r.rec == nil {
		return giode.Box(giode.Styles{})
	}
	giode.HandleBack(func() { r.finish(false) }) // Back discards the take
	elapsed := r.rec.Elapsed()
	clock := fmt.Sprintf("%s / %s", core.Clock(elapsed), core.Clock(MaxLength))
	hint := "Read naturally, a few sentences, in a quiet room."
	if elapsed >= MinLength {
		hint = "Sounds good — press Stop when you're done."
	}
	return giode.Stack(giode.Styles{Width: -1, Gap: 10, Padding: giode.UniformInset(12),
		Background: "theme.danger/0.08", BorderRadius: theme.RadiusMd, BorderWidth: 1, BorderColor: "theme.danger/0.3"},
		giode.Row(giode.Styles{Width: -1, Gap: 8, Align: giode.AlignCenter, Wrap: giode.WrapNoWrap},
			giode.Box(giode.Styles{Width: 9, Height: 9, BorderRadius: 99, Background: "theme.danger"}),
			giode.Box(giode.Styles{FlexGrow: 1}, giode.Text("Recording", giode.Styles{Color: "theme.text", FontSize: 13, FontWeight: giode.FontWeightBold})),
			giode.Text(clock, giode.Styles{Color: "theme.secondary", FontSize: 12, FontFamily: theme.Mono}),
			r.cancel,
		),
		giode.Box(giode.Styles{Width: -1}, r.meter),
		giode.Row(giode.Styles{Width: -1, Gap: 8, Align: giode.AlignCenter, Wrap: giode.WrapNoWrap},
			giode.Box(giode.Styles{FlexGrow: 1}, giode.Text(hint, giode.Styles{Color: "theme.muted", FontSize: 12})),
			r.stop,
		),
	)
}
