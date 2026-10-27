package onnx

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"strings"
	"sync"
	"unicode"
	"unicode/utf8"

	"github.com/mehran9675/sonora/internal/audio/wav"
	"github.com/mehran9675/sonora/internal/engines"
	"github.com/mehran9675/sonora/internal/engines/onnx/sherpa"
)

// maxUtterance is the longest stretch, in seconds, handed to a
// recognizer at once; Whisper takes at most 30.
const maxUtterance = 20

// Recognizer is an on-device speech-to-text engine.
type Recognizer struct {
	m      *engines.Manifest
	dir    string
	pick   func(engines.Values) model
	models []model

	mu     sync.Mutex
	rec    *sherpa.Recognizer
	vad    *sherpa.VAD
	loaded string // what rec was created for, see key
}

func newRecognizer(m *engines.Manifest, dir string, models []model) *Recognizer {
	pick := byOption(models)
	return &Recognizer{m: m, dir: dir, models: models, pick: func(o engines.Values) model { return pick("", o) }}
}

func (e *Recognizer) Manifest() *engines.Manifest { return e.m }

func (e *Recognizer) defaultModel() model {
	return e.pick(engines.Resolve(e.m.Options, nil))
}

// Installed reports whether any of the engine's models is downloaded.
func (e *Recognizer) Installed() bool { return anyDownloaded(e.dir, e.models) }

// DownloadedModels reports which packages are on this device.
func (e *Recognizer) DownloadedModels() map[string]bool { return downloaded(e.dir, e.models) }

// InstallModel downloads the package with the given id.
func (e *Recognizer) InstallModel(ctx context.Context, id string, log engines.Logger) error {
	return fetchModel(ctx, e.dir, e.models, id, log)
}

// Install downloads the default model.
func (e *Recognizer) Install(ctx context.Context, log engines.Logger) error {
	_, err := fetch(ctx, e.dir, e.defaultModel(), log)
	return err
}

// Uninstall unloads and deletes every package this engine downloaded.
func (e *Recognizer) Uninstall() error {
	e.Close()
	return removeModels(e.dir, e.models)
}

func (e *Recognizer) Voices(context.Context, engines.Values, engines.Logger) ([]engines.Voice, error) {
	return nil, nil
}

func (e *Recognizer) Synthesize(context.Context, engines.Request, engines.Logger) error {
	return errors.New(e.m.Name + " turns speech into text; it can't speak")
}

// Loaded reports whether a model is in memory.
func (e *Recognizer) Loaded() bool {
	e.mu.Lock()
	defer e.mu.Unlock()
	return e.rec != nil
}

// Close frees the model's memory.
func (e *Recognizer) Close() {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.unload()
}

func (e *Recognizer) unload() {
	if e.rec != nil {
		e.rec.Close()
		e.rec, e.loaded = nil, ""
	}
	if e.vad != nil {
		e.vad.Close()
		e.vad = nil
	}
}

// load returns the recognizer and voice detector for a request,
// downloading and (re)loading as needed. Whisper's language and task
// are fixed when it loads, so changing them reloads it.
func (e *Recognizer) load(ctx context.Context, r engines.TranscribeRequest, log engines.Logger) (*sherpa.Recognizer, *sherpa.VAD, error) {
	m := e.pick(r.Options)
	dir, err := fetch(ctx, e.dir, m, log)
	if err != nil {
		return nil, nil, err
	}
	cfg, err := recognizerConfig(dir, m)
	if err != nil {
		return nil, nil, err
	}
	cfg.Language = supportedLanguage(m, r.Language)
	cfg.Translate = r.Translate && m.Family == sherpa.Whisper
	key := fmt.Sprintf("%s|%s|%v", m.ID, cfg.Language, cfg.Translate)
	e.mu.Lock()
	defer e.mu.Unlock()
	if e.rec != nil && e.loaded == key {
		return e.rec, e.vad, nil
	}
	e.unload()
	log("Loading " + e.m.Name + "…")
	vad, err := sherpa.NewVAD(extraPath(e.dir, sileroVAD), maxUtterance)
	if err != nil {
		return nil, nil, err
	}
	rec, err := sherpa.NewRecognizer(cfg)
	if err != nil {
		vad.Close()
		return nil, nil, err
	}
	e.rec, e.vad, e.loaded = rec, vad, key
	return rec, vad, nil
}

// recognizerConfig finds a package's files.
func recognizerConfig(dir string, m model) (sherpa.RecognizerConfig, error) {
	cfg := sherpa.RecognizerConfig{Family: m.Family}
	find := func(dst *string, patterns ...string) error {
		for _, p := range patterns {
			if found, _ := filepath.Glob(filepath.Join(dir, p)); len(found) > 0 {
				*dst = found[0]
				return nil
			}
		}
		return errors.New("the downloaded model is incomplete; remove the engine and install it again")
	}
	var errs []error
	switch m.Family {
	case sherpa.Whisper:
		errs = append(errs, find(&cfg.Encoder, "*-encoder.int8.onnx"), find(&cfg.Decoder, "*-decoder.int8.onnx"),
			find(&cfg.Tokens, "*-tokens.txt"))
	case sherpa.Transducer:
		errs = append(errs, find(&cfg.Encoder, "encoder*.onnx"), find(&cfg.Decoder, "decoder*.onnx"),
			find(&cfg.Joiner, "joiner*.onnx"), find(&cfg.Tokens, "tokens.txt"))
	case sherpa.SenseVoice:
		errs = append(errs, find(&cfg.Model, "model*.onnx"), find(&cfg.Tokens, "tokens.txt"))
	}
	return cfg, errors.Join(errs...)
}

// Transcribe cuts the recording into utterances where there is speech
// and recognizes each, so segments carry their times.
func (e *Recognizer) Transcribe(ctx context.Context, r engines.TranscribeRequest, log engines.Logger) (*engines.Transcript, error) {
	clip, err := wav.Read(r.Audio)
	if err != nil {
		return nil, err
	}
	if clip.SampleRate != 16000 {
		clip.Samples, clip.SampleRate = clip.Resample(16000), 16000
	}
	rec, vad, err := e.load(ctx, r, log)
	if err != nil {
		return nil, err
	}
	total := len(clip.Samples)
	out := &engines.Transcript{}
	last := -1
	var failure error
	recognize := func(start int, speech []float32) bool {
		if ctx.Err() != nil {
			return false
		}
		if pct := int(int64(start) * 100 / int64(max(total, 1))); pct != last {
			last = pct
			log(fmt.Sprintf("Transcribing… %d%%", pct))
		}
		res, err := rec.Recognize(speech, 16000)
		if err != nil {
			failure = err
			return false
		}
		if text := strings.TrimSpace(res.Text); text != "" {
			out.Segments = append(out.Segments, engines.Segment{
				Start: float64(start) / 16000, End: float64(start+len(speech)) / 16000, Text: text})
		}
		// Only Whisper's language detection is reliable enough to show.
		if out.Language == "" && e.pick(r.Options).Family == sherpa.Whisper {
			out.Language = strings.Trim(res.Language, "<|>")
		}
		return true
	}
	// The detector reports speech a little after it starts; widen each
	// stretch so first syllables aren't lost.
	const before, after = 16000 * 4 / 10, 16000 * 2 / 10
	vad.Split(clip.Samples, func(start int, speech []float32) bool {
		from, to := max(0, start-before), min(total, start+len(speech)+after)
		return recognize(from, clip.Samples[from:to])
	})
	// Speech too quiet for the detector: fall back to fixed windows.
	if len(out.Segments) == 0 && failure == nil && ctx.Err() == nil {
		step := maxUtterance * 16000
		for i := 0; i < total && recognize(i, clip.Samples[i:min(total, i+step)]); i += step {
		}
	}
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}
	if failure != nil {
		return nil, failure
	}
	log("Transcribing… 100%")
	parts := make([]string, len(out.Segments))
	for i, s := range out.Segments {
		parts[i] = s.Text
	}
	out.Text = JoinText(parts)
	return out, nil
}

// JoinText joins transcript pieces with spaces, except between pieces
// of languages written without them (Chinese, Japanese…).
func JoinText(parts []string) string {
	var b strings.Builder
	for i, p := range parts {
		if i > 0 {
			prev, _ := utf8.DecodeLastRuneInString(parts[i-1])
			next, _ := utf8.DecodeRuneInString(p)
			if !(unspaced(prev) && unspaced(next)) {
				b.WriteByte(' ')
			}
		}
		b.WriteString(p)
	}
	return b.String()
}

func unspaced(r rune) bool {
	return unicode.In(r, unicode.Han, unicode.Hiragana, unicode.Katakana, unicode.Thai) ||
		strings.ContainsRune("。，！？、；：」』）", r)
}
