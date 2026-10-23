package manager

import (
	"context"
	"errors"
	"os"
	"strings"
	"time"

	"github.com/mehran9675/sonora/internal/audio/wav"
	"github.com/mehran9675/sonora/internal/engines"
	"github.com/mehran9675/sonora/internal/library"
)

// Job is a generation in progress.
type Job struct {
	Runtime *Runtime
	Started time.Time
}

// Input is what the Studio asks for.
type Input struct {
	Text          string
	Voice         engines.Voice
	Speed         float64
	Reference     string
	ReferenceText string
}

// Generate synthesizes in the background and adds the clip to the
// library; done runs on the UI goroutine.
func (m *Manager) Generate(rt *Runtime, in Input, done func(library.Entry, error)) {
	if m.Job != nil || rt.Busy() {
		done(library.Entry{}, errors.New("already generating"))
		return
	}
	man := rt.Manifest()
	id, out := m.library.NewFile()
	req := engines.Request{
		Text:          strings.TrimSpace(in.Text),
		Voice:         in.Voice.ID,
		Speed:         in.Speed,
		Options:       m.values(man),
		Output:        out,
		Reference:     in.Reference,
		ReferenceText: strings.TrimSpace(in.ReferenceText),
	}
	if !man.Speed {
		req.Speed = 0
	}
	ctx, cancel := context.WithCancel(context.Background())
	rt.cancel = cancel
	rt.State, rt.Err, rt.Progress, rt.Status = StateWorking, "", -1, "Generating…"
	m.Job = &Job{Runtime: rt, Started: time.Now()}
	sink := m.sink(rt)
	go func() {
		err := explain(rt.Engine.Synthesize(ctx, req, sink.Log))
		var entry library.Entry
		if err == nil {
			entry, err = m.record(id, out, man, in, req)
		}
		if err != nil {
			os.Remove(out)
		}
		m.post(func() {
			rt.cancel, m.Job = nil, nil
			rt.State, rt.Progress = StateReady, -1
			switch {
			case errors.Is(err, context.Canceled):
				rt.Status = "Stopped"
			case err != nil:
				rt.Err, rt.Status = err.Error(), "Generation failed"
			default:
				rt.Status = "Done"
			}
			done(entry, err)
		})
	}()
}

// record checks the engine's output and adds it to the library.
func (m *Manager) record(id, out string, man *engines.Manifest, in Input, req engines.Request) (library.Entry, error) {
	clip, err := wav.Read(out)
	if err != nil {
		return library.Entry{}, errors.New("the engine did not produce playable audio: " + err.Error())
	}
	if len(clip.Samples) == 0 {
		return library.Entry{}, errors.New("the engine produced an empty clip")
	}
	voice := in.Voice.Label()
	if req.Reference != "" {
		voice = "Cloned voice"
	}
	e := library.Entry{
		ID: id, Text: req.Text, EngineID: man.ID, EngineName: man.Name,
		Voice: in.Voice.ID, VoiceName: voice, Created: time.Now(),
		Duration: clip.Duration(), File: id + ".wav",
	}
	m.library.Add(e)
	return e, nil
}
