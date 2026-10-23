package studio

import (
	"path/filepath"

	"github.com/mehran9675/giode"

	"github.com/mehran9675/sonora/internal/engines"
	"github.com/mehran9675/sonora/ui/theme"
)

// voiceChoice is the Voice section: the voice picker, plus a switch to
// voice cloning on engines that can copy a voice from a recording.
func voiceChoice(f *engineForm) giode.Element {
	m := f.rt.Manifest()
	switch m.Clone {
	case engines.CloneNone:
		return f.voiceTrigger.Element()
	case engines.CloneRequired:
		return cloneBody(f)
	}
	body := f.voiceTrigger.Element()
	if f.cloneMode {
		body = cloneBody(f)
	}
	return giode.Stack(giode.Styles{Width: -1, Gap: 12}, f.modes.Element(), body)
}

// cloneBody picks the recording to copy and, where the engine uses it,
// its transcript.
func cloneBody(f *engineForm) giode.Element {
	m := f.rt.Manifest()
	body := []giode.Element{giode.Text("Record yourself or pick a file: 5–30 seconds of one person speaking clearly, "+
		"with no music or noise.", giode.Styles{Color: "theme.muted", FontSize: 12, LineHeight: 17})}
	switch {
	case f.recorder.Recording():
		body = append(body, f.recorder.Element())
	case f.reference == "":
		body = append(body, giode.Row(giode.Styles{Width: -1, Gap: 8}, f.recorder.Button(), f.refPick))
	default:
		body = append(body, giode.Row(giode.Styles{Width: -1, Gap: 10, Align: giode.AlignCenter, Wrap: giode.WrapNoWrap,
			Padding: giode.SymmetricInset(10, 8), Background: "theme.primary/0.1", BorderRadius: theme.RadiusSm + 2},
			f.refPlay,
			giode.Box(giode.Styles{FlexGrow: 1}, giode.Text(filepath.Base(f.reference),
				giode.Styles{Color: "theme.text", FontSize: 13, TextOverflow: giode.TextOverflowEllipsis})),
			f.refClear,
		))
	}
	if m.Transcript != engines.CloneNone {
		label := "Transcript · optional, detected automatically"
		if m.Transcript == engines.CloneRequired {
			label = "Transcript · required"
		}
		body = append(body, giode.Text(label, theme.Caption), giode.Box(giode.Styles{Width: -1}, f.transcript))
	}
	return giode.Stack(giode.Styles{Width: -1, Gap: 10}, body...)
}
