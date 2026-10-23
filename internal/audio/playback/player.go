package playback

import (
	"io"
	"sync"
	"time"

	"github.com/ebitengine/oto/v3"

	"github.com/mehran9675/sonora/internal/audio"
)

// Player plays one clip at a time. All methods are safe from any
// goroutine.
type Player struct {
	mu     sync.Mutex
	src    *source
	p      *oto.Player
	length time.Duration
	key    string
}

// New returns an idle player.
func New() *Player { return &Player{} }

// Load replaces the current clip; key identifies it (usually the file
// path) so callers can tell what is loaded.
func (pl *Player) Load(c *audio.Clip, key string) error {
	dev, err := device()
	if err != nil {
		return err
	}
	pl.mu.Lock()
	defer pl.mu.Unlock()
	pl.closeLocked()
	pl.src = &source{samples: c.Resample(deviceRate)}
	pl.p = dev.NewPlayer(pl.src)
	pl.p.SetBufferSize(deviceRate * frameBytes / 10)
	pl.length, pl.key = c.Duration(), key
	return nil
}

// Loaded returns the key of the loaded clip.
func (pl *Player) Loaded() string {
	pl.mu.Lock()
	defer pl.mu.Unlock()
	return pl.key
}

// Unload stops and forgets the current clip.
func (pl *Player) Unload() {
	pl.mu.Lock()
	defer pl.mu.Unlock()
	pl.closeLocked()
	pl.p, pl.src, pl.key, pl.length = nil, nil, "", 0
}

func (pl *Player) closeLocked() {
	if pl.p != nil {
		pl.p.Pause()
		pl.p.Close()
	}
}

// Play starts or resumes playback, rewinding a finished clip.
func (pl *Player) Play() {
	pl.mu.Lock()
	defer pl.mu.Unlock()
	if pl.p == nil {
		return
	}
	if pl.src.position() >= pl.src.total() && pl.p.BufferedSize() == 0 {
		pl.p.Seek(0, io.SeekStart)
	}
	pl.p.Play()
}

// Pause pauses playback.
func (pl *Player) Pause() {
	pl.mu.Lock()
	defer pl.mu.Unlock()
	if pl.p != nil {
		pl.p.Pause()
	}
}

// Toggle plays or pauses.
func (pl *Player) Toggle() {
	if pl.Playing() {
		pl.Pause()
	} else {
		pl.Play()
	}
}

// Playing reports whether audio is being played.
func (pl *Player) Playing() bool {
	pl.mu.Lock()
	defer pl.mu.Unlock()
	return pl.p != nil && pl.p.IsPlaying()
}

// Length returns the loaded clip's duration.
func (pl *Player) Length() time.Duration {
	pl.mu.Lock()
	defer pl.mu.Unlock()
	return pl.length
}

// Progress returns the playback position as a fraction in [0, 1].
func (pl *Player) Progress() float32 {
	pl.mu.Lock()
	defer pl.mu.Unlock()
	if pl.p == nil || pl.src.total() == 0 {
		return 0
	}
	pos := pl.src.position() - int64(pl.p.BufferedSize())
	return float32(max(0, min(1, float64(pos)/float64(pl.src.total()))))
}

// Seek moves playback to the fraction f of the clip.
func (pl *Player) Seek(f float32) {
	pl.mu.Lock()
	defer pl.mu.Unlock()
	if pl.p == nil {
		return
	}
	pl.p.Seek(int64(float64(pl.src.total())*float64(max(0, min(1, f)))), io.SeekStart)
}
