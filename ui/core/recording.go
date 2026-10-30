package core

import (
	"time"

	"github.com/mehran9675/sonora/ui/tr"

	"github.com/mehran9675/sonora/internal/paths"
)

// RecordQuality is a recording quality setting.
type RecordQuality struct {
	Value, Name string // Name is English; translate it with tr.T
	Bitrate     int    // Opus, bits per second
}

// Detail describes the quality's bitrate and size, translated.
func (q RecordQuality) Detail() string {
	return tr.T("{kbps} kbps · about {mb} MB an hour", "kbps", q.Bitrate/1000, "mb", q.Bitrate*3600/8/1000000)
}

// RecordQualities are the recording quality settings, smallest first.
// Recordings are Opus: even Small keeps speech clear.
var RecordQualities = []RecordQuality{
	{"small", tr.Msg("Small"), 16000},
	{"standard", tr.Msg("Standard"), 32000},
	{"high", tr.Msg("High"), 64000},
	{"best", tr.Msg("Best"), 128000},
}

// RecordLimits are the choices for the longest recording, in minutes.
var RecordLimits = []int{15, 30, 60, 120, 240, 480}

// defaultRecordMinutes is the longest recording unless changed.
const defaultRecordMinutes = 60

// RecordBitrate returns the bitrate recordings are made at.
func (c *Ctx) RecordBitrate() int {
	q := c.Settings.Get().RecordQuality
	for _, r := range RecordQualities {
		if r.Value == q {
			return r.Bitrate
		}
	}
	return RecordQualities[1].Bitrate
}

// RecordMax returns the longest a recording may run.
func (c *Ctx) RecordMax() time.Duration {
	if m := c.Settings.Get().RecordMinutes; m > 0 {
		return time.Duration(m) * time.Minute
	}
	return defaultRecordMinutes * time.Minute
}

// RecordingsDir is where recordings are saved: the folder chosen in
// Settings, or recordings/ in the data folder.
func (c *Ctx) RecordingsDir() string {
	if d := c.Settings.Get().RecordDir; d != "" && !Mobile {
		return d
	}
	return DefaultRecordingsDir()
}

// DefaultRecordingsDir is the recordings folder unless one is chosen.
func DefaultRecordingsDir() string { return paths.Sub("recordings") }
