package manager

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/mehran9675/sonora/internal/audio/decode"
	"github.com/mehran9675/sonora/internal/audio/wav"
	"github.com/mehran9675/sonora/internal/engines"
	"github.com/mehran9675/sonora/internal/transcripts"
)

// TranscribeInput is what the Transcribe page asks for.
type TranscribeInput struct {
	// Source is the audio file, in any format decode reads; Title names
	// it in the list.
	Source    string
	Title     string
	Language  string
	Translate bool
}

// Transcribe converts the recording to 16 kHz, transcribes it in the
// background and stores the result; done runs on the UI goroutine.
func (m *Manager) Transcribe(rt *Runtime, in TranscribeInput, done func(transcripts.Entry, error)) {
	tr, ok := rt.Engine.(engines.Transcriber)
	man := rt.Manifest()
	switch {
	case !ok || !man.Transcribes():
		done(transcripts.Entry{}, errors.New(man.Name+" can't transcribe"))
		return
	case m.TranscribeJob != nil || rt.Busy():
		done(transcripts.Entry{}, errors.New("already transcribing"))
		return
	}
	ctx, cancel := context.WithCancel(context.Background())
	rt.cancel = cancel
	rt.State, rt.Err, rt.Progress, rt.Status = StateWorking, "", -1, "Reading the audio…"
	m.TranscribeJob = &Job{Runtime: rt, Started: time.Now()}
	sink := m.sink(rt)
	id, audioPath := m.transcripts.NewFile()
	go func() {
		entry, err := m.transcribe(ctx, tr, man, id, audioPath, in, sink.Log)
		if err != nil {
			os.Remove(audioPath)
		}
		m.post(func() {
			rt.cancel, m.TranscribeJob = nil, nil
			rt.State, rt.Progress = StateReady, -1
			switch {
			case errors.Is(err, context.Canceled):
				rt.Status = "Stopped"
			case err != nil:
				rt.Err, rt.Status = err.Error(), "Transcription failed"
			default:
				rt.Status = "Done"
			}
			done(entry, err)
		})
	}()
}

func (m *Manager) transcribe(ctx context.Context, tr engines.Transcriber, man *engines.Manifest, id, audioPath string,
	in TranscribeInput, log engines.Logger) (transcripts.Entry, error) {
	log("Reading the audio…")
	clip, err := decode.File(in.Source, decode.SpeechRate)
	if err != nil {
		return transcripts.Entry{}, err
	}
	if err := wav.Write(audioPath, clip); err != nil {
		return transcripts.Entry{}, err
	}
	if ctx.Err() != nil {
		return transcripts.Entry{}, ctx.Err()
	}
	req := engines.TranscribeRequest{Audio: audioPath, Language: in.Language, Translate: in.Translate && man.Translate,
		Options: m.values(man)}
	res, err := tr.Transcribe(ctx, req, log)
	if err != nil {
		return transcripts.Entry{}, explain(err)
	}
	text := strings.TrimSpace(res.Text)
	if text == "" {
		return transcripts.Entry{}, errors.New("no speech was recognized in the recording")
	}
	lang := res.Language
	if lang == "" {
		lang = in.Language
	}
	title := in.Title
	if title == "" {
		title = filepath.Base(in.Source)
	}
	e := transcripts.Entry{
		ID: id, Title: title, Text: text, Language: lang, Translated: req.Translate, Segments: res.Segments,
		EngineID: man.ID, EngineName: man.Name, Created: time.Now(), Duration: clip.Duration(),
		Audio: filepath.Base(audioPath),
	}
	m.transcripts.Add(e)
	return e, nil
}
