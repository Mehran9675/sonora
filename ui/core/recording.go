package core

import (
	"time"

	"github.com/mehran9675/sonora/internal/paths"
)

// RecordQuality is a recording quality setting.
type RecordQuality struct {
	Value, Label string
	Bitrate      int // Opus, bits per second
}

// RecordQualities are the recording quality settings, smallest first.
// Recordings are Opus: even Small keeps speech clear.
var RecordQualities = []RecordQuality{
	{"small", "Small · 16 kbps · about 7 MB an hour", 16000},
	{"standard", "Standard · 32 kbps · about 14 MB an hour", 32000},
	{"high", "High · 64 kbps · about 28 MB an hour", 64000},
	{"best", "Best · 128 kbps · about 56 MB an hour", 128000},
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
