package transcripts

import (
	"fmt"
	"strings"

	"github.com/mehran9675/sonora/internal/engines"
)

// SRT renders the entry as SubRip subtitles. Without segments the
// whole text becomes one cue.
func SRT(e Entry) string {
	segs := e.Segments
	if len(segs) == 0 {
		segs = []engines.Segment{{Start: 0, End: e.Duration.Seconds(), Text: e.Text}}
	}
	var b strings.Builder
	for i, s := range segs {
		fmt.Fprintf(&b, "%d\n%s --> %s\n%s\n\n", i+1, stamp(s.Start, ","), stamp(s.End, ","), strings.TrimSpace(s.Text))
	}
	return b.String()
}

// VTT renders the entry as WebVTT subtitles.
func VTT(e Entry) string {
	segs := e.Segments
	if len(segs) == 0 {
		segs = []engines.Segment{{Start: 0, End: e.Duration.Seconds(), Text: e.Text}}
	}
	var b strings.Builder
	b.WriteString("WEBVTT\n\n")
	for _, s := range segs {
		fmt.Fprintf(&b, "%s --> %s\n%s\n\n", stamp(s.Start, "."), stamp(s.End, "."), strings.TrimSpace(s.Text))
	}
	return b.String()
}

// stamp formats seconds as HH:MM:SS,mmm (sep is "," for SRT, "." for
// WebVTT).
func stamp(secs float64, sep string) string {
	ms := int64(secs*1000 + 0.5)
	if ms < 0 {
		ms = 0
	}
	return fmt.Sprintf("%02d:%02d:%02d%s%03d", ms/3600000, ms/60000%60, ms/1000%60, sep, ms%1000)
}
