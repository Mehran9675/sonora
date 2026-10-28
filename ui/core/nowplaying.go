package core

import (
	"time"

	"github.com/mehran9675/sonora/internal/audio/decode"
)

// peakCount is the waveform resolution kept for display.
const peakCount = 600

// NowPlaying is the clip loaded into the player.
type NowPlaying struct {
	Path     string
	Title    string
	Subtitle string
	Peaks    []float32
	Duration time.Duration
}

// Open decodes path in the background and loads it into the player,
// starting playback when autoplay is set.
func (c *Ctx) Open(path, title, subtitle string, autoplay bool) {
	go func() {
		clip, err := decode.File(path, 0) // WAV clips, Opus recordings…
		if err == nil {
			err = c.Player.Load(clip, path)
		}
		var peaks []float32
		if clip != nil {
			peaks = clip.Peaks(peakCount)
		}
		c.App.Post(func() {
			if clip != nil {
				c.Now = &NowPlaying{Path: path, Title: title, Subtitle: subtitle, Peaks: peaks, Duration: clip.Duration()}
			}
			if err != nil {
				c.Error(err)
				return
			}
			if autoplay {
				c.Player.Play()
			}
		})
	}()
}

// IsPlaying reports whether path is loaded and playing.
func (c *Ctx) IsPlaying(path string) bool {
	return c.Player.Loaded() == path && c.Player.Playing()
}

// PlayPause toggles path, loading it first when another clip is loaded.
func (c *Ctx) PlayPause(path, title, subtitle string) {
	if c.Player.Loaded() == path {
		c.Player.Toggle()
		return
	}
	c.Open(path, title, subtitle, true)
}
