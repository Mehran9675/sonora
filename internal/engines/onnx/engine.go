package onnx

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"sync"

	"github.com/mehran9675/sonora/internal/audio"
	"github.com/mehran9675/sonora/internal/audio/convert"
	"github.com/mehran9675/sonora/internal/audio/wav"
	"github.com/mehran9675/sonora/internal/engines"
	"github.com/mehran9675/sonora/internal/engines/onnx/sherpa"
)

// Available reports whether this build can run on-device models.
const Available = sherpa.Available

// Engine is one on-device engine: Kitten, Kokoro, Piper or ZipVoice.
type Engine struct {
	m   *engines.Manifest
	dir string // downloaded packages: dir/<package id>/
	// pick returns the package a request needs.
	pick func(voice string, opts engines.Values) model
	// models are all the packages the engine may download.
	models []model

	mu     sync.Mutex
	tts    *sherpa.TTS
	loaded string // package id of tts
}

// All returns the on-device engines, storing models under dataDir.
func All(dataDir string) []engines.Engine {
	dir := filepath.Join(dataDir, "models", "ondevice")
	var piper []model
	for _, v := range piperVoices {
		piper = append(piper, piperModel(v.ID))
	}
	return []engines.Engine{
		&Engine{m: kittenManifest(), dir: dir, pick: byOption(kittenModels), models: kittenModels},
		&Engine{m: kokoroManifest(), dir: dir, pick: byOption(kokoroModels), models: kokoroModels},
		&Engine{m: piperManifest(), dir: dir, models: piper, pick: func(voice string, _ engines.Values) model {
			if voice == "" {
				voice = piperVoices[0].ID
			}
			return piperModel(voice)
		}},
		&Engine{m: zipvoiceManifest(), dir: dir, pick: byOption(zipvoiceModels), models: zipvoiceModels},
		newRecognizer(whisperManifest(), dir, whisperModels),
		newRecognizer(parakeetManifest(), dir, []model{parakeetModel}),
		newRecognizer(senseVoiceManifest(), dir, []model{senseVoiceModel}),
	}
}

// byOption picks the package named by the "model" option.
func byOption(models []model) func(string, engines.Values) model {
	return func(_ string, o engines.Values) model {
		for _, m := range models {
			if m.ID == o.String("model") {
				return m
			}
		}
		return models[0]
	}
}

func (e *Engine) Manifest() *engines.Manifest { return e.m }

// defaultModel is what Install downloads.
func (e *Engine) defaultModel() model {
	return e.pick(e.m.DefaultVoice, engines.Resolve(e.m.Options, nil))
}

// Installed reports whether the default model is downloaded.
func (e *Engine) Installed() bool {
	_, err := os.Stat(filepath.Join(e.dir, e.defaultModel().ID))
	return err == nil
}

// Install downloads the default model.
func (e *Engine) Install(ctx context.Context, log engines.Logger) error {
	_, err := fetch(ctx, e.dir, e.defaultModel(), log)
	return err
}

// Uninstall unloads and deletes every package this engine downloaded.
func (e *Engine) Uninstall() error {
	e.Close()
	return removeModels(e.dir, e.models)
}

// removeModels deletes downloaded packages and their extra files.
func removeModels(dir string, models []model) error {
	var paths []string
	for _, m := range models {
		paths = append(paths, filepath.Join(dir, m.ID))
		for _, url := range m.Extra {
			paths = append(paths, extraPath(dir, url))
		}
	}
	for _, p := range paths {
		if err := os.RemoveAll(p); err != nil {
			return err
		}
	}
	return nil
}

// Voices returns the fixed voice list.
func (e *Engine) Voices(context.Context, engines.Values, engines.Logger) ([]engines.Voice, error) {
	return e.m.Voices, nil
}

// Loaded reports whether a model is in memory.
func (e *Engine) Loaded() bool {
	e.mu.Lock()
	defer e.mu.Unlock()
	return e.tts != nil
}

// Close frees the model's memory.
func (e *Engine) Close() {
	e.mu.Lock()
	defer e.mu.Unlock()
	if e.tts != nil {
		e.tts.Close()
		e.tts, e.loaded = nil, ""
	}
}

// load returns the model for m, downloading and loading it as needed.
func (e *Engine) load(ctx context.Context, m model, log engines.Logger) (*sherpa.TTS, error) {
	dir, err := fetch(ctx, e.dir, m, log)
	if err != nil {
		return nil, err
	}
	e.mu.Lock()
	defer e.mu.Unlock()
	if e.tts != nil && e.loaded == m.ID {
		return e.tts, nil
	}
	if e.tts != nil {
		e.tts.Close()
		e.tts = nil
	}
	cfg, err := config(dir, m)
	if err != nil {
		return nil, err
	}
	if m.Family == sherpa.ZipVoice {
		cfg.Vocoder = extraPath(e.dir, vocos)
	}
	log("Loading " + e.m.Name + "…")
	tts, err := sherpa.New(cfg)
	if err != nil {
		return nil, err
	}
	e.tts, e.loaded = tts, m.ID
	return tts, nil
}

// config finds a package's files.
func config(dir string, m model) (sherpa.Config, error) {
	find := func(pattern string) (string, error) {
		found, _ := filepath.Glob(filepath.Join(dir, pattern))
		if len(found) == 0 {
			return "", errors.New("the downloaded model is incomplete; remove the engine and install it again")
		}
		return found[0], nil
	}
	cfg := sherpa.Config{Family: m.Family,
		Tokens: filepath.Join(dir, "tokens.txt"), DataDir: filepath.Join(dir, "espeak-ng-data")}
	var err error
	if m.Family == sherpa.ZipVoice {
		if cfg.Encoder, err = find("encoder*.onnx"); err != nil {
			return cfg, err
		}
		if cfg.Decoder, err = find("decoder*.onnx"); err != nil {
			return cfg, err
		}
	} else if cfg.Model, err = find("*.onnx"); err != nil {
		return cfg, err
	}
	if m.Family == sherpa.Kitten || m.Family == sherpa.Kokoro {
		cfg.Voices = filepath.Join(dir, "voices.bin")
	}
	if m.Lexicon != "" {
		cfg.Lexicon = filepath.Join(dir, m.Lexicon)
	}
	if m.Dict != "" {
		cfg.DictDir = filepath.Join(dir, m.Dict)
	}
	return cfg, nil
}

// Synthesize generates speech and writes it as WAV.
func (e *Engine) Synthesize(ctx context.Context, r engines.Request, log engines.Logger) error {
	m := e.pick(r.Voice, r.Options)
	p := sherpa.Params{Speed: float32(r.Speed), Steps: m.Steps}
	if p.Speed <= 0 {
		p.Speed = 1
	}
	if m.Family == sherpa.Kitten || m.Family == sherpa.Kokoro {
		p.Speaker, _ = strconv.Atoi(r.Voice)
	}
	if m.Family == sherpa.ZipVoice {
		if r.Reference == "" || r.ReferenceText == "" {
			return errors.New(e.m.Name + " needs a recording of the voice and its transcript")
		}
		clip, err := readReference(r.Reference)
		if err != nil {
			return err
		}
		p.Reference, p.ReferenceRate, p.ReferenceText = clip.Samples, clip.SampleRate, r.ReferenceText
	}
	tts, err := e.load(ctx, m, log)
	if err != nil {
		return err
	}
	last := -1
	samples, rate, err := tts.Generate(r.Text, p, func(p float32) bool {
		if pct := int(p * 100); pct != last {
			last = pct
			log(fmt.Sprintf("Generating… %d%%", pct))
		}
		return ctx.Err() == nil
	})
	if errors.Is(err, sherpa.ErrStopped) && ctx.Err() != nil {
		return ctx.Err()
	}
	if err != nil {
		return err
	}
	return wav.Write(r.Output, &audio.Clip{Samples: samples, SampleRate: rate})
}

// readReference loads a cloning recording. Other formats than WAV are
// converted with ffmpeg where it is installed (not on phones; recordings
// made in the app are WAV).
func readReference(path string) (*audio.Clip, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	if !wav.IsWAV(data) {
		if data, err = convert.ToWAV(data); err != nil {
			return nil, errors.New("this engine needs a WAV recording; record your voice in the app or pick a WAV file")
		}
	}
	return wav.Decode(data)
}
