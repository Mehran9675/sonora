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
	"github.com/mehran9675/sonora/internal/audio/wav"
	"github.com/mehran9675/sonora/internal/engines"
	"github.com/mehran9675/sonora/internal/engines/onnx/sherpa"
)

// Available reports whether this build can run on-device models.
const Available = sherpa.Available

// Engine is one on-device engine: Kitten, Kokoro or Piper.
type Engine struct {
	m   *engines.Manifest
	dir string // downloaded packages: dir/<package id>/
	// pick returns the package a request needs.
	pick func(voice string, opts engines.Values) model

	mu     sync.Mutex
	tts    *sherpa.TTS
	loaded string // package id of tts
}

// All returns the on-device engines, storing models under dataDir.
func All(dataDir string) []engines.Engine {
	dir := filepath.Join(dataDir, "models", "ondevice")
	return []engines.Engine{
		&Engine{m: kittenManifest(), dir: dir, pick: func(_ string, o engines.Values) model {
			for _, m := range kittenModels {
				if m.ID == o.String("model") {
					return m
				}
			}
			return kittenModels[0]
		}},
		&Engine{m: kokoroManifest(), dir: dir, pick: func(string, engines.Values) model { return kokoroModel }},
		&Engine{m: piperManifest(), dir: dir, pick: func(voice string, _ engines.Values) model {
			if voice == "" {
				voice = piperVoices[0].ID
			}
			return piperModel(voice)
		}},
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
	ids := []string{e.defaultModel().ID}
	switch e.m.ID {
	case "ondevice-kitten":
		for _, m := range kittenModels {
			ids = append(ids, m.ID)
		}
	case "ondevice-piper":
		for _, v := range piperVoices {
			ids = append(ids, piperModel(v.ID).ID)
		}
	}
	for _, id := range ids {
		if err := os.RemoveAll(filepath.Join(e.dir, id)); err != nil {
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
	onnx, _ := filepath.Glob(filepath.Join(dir, "*.onnx"))
	if len(onnx) == 0 {
		return sherpa.Config{}, errors.New("the downloaded model has no .onnx file; remove the engine and install it again")
	}
	cfg := sherpa.Config{Family: m.Family, Model: onnx[0],
		Tokens: filepath.Join(dir, "tokens.txt"), DataDir: filepath.Join(dir, "espeak-ng-data")}
	if m.Family != sherpa.VITS {
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
	tts, err := e.load(ctx, m, log)
	if err != nil {
		return err
	}
	sid := 0
	if m.Family != sherpa.VITS {
		sid, _ = strconv.Atoi(r.Voice)
	}
	speed := float32(r.Speed)
	if speed <= 0 {
		speed = 1
	}
	last := -1
	samples, rate, err := tts.Generate(r.Text, sid, speed, func(p float32) bool {
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
