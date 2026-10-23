//go:build sherpa

package sherpa

import (
	"os"
	"path/filepath"
	"testing"
)

// TestGenerate runs real models when SHERPA_TEST_MODELS points at a
// folder holding unpacked sherpa-onnx model packages.
func TestGenerate(t *testing.T) {
	root := os.Getenv("SHERPA_TEST_MODELS")
	if root == "" {
		t.Skip("set SHERPA_TEST_MODELS to a folder of unpacked models")
	}
	cases := map[string]Config{
		"vits-piper-en_US-amy-low":    {Family: VITS, Model: "en_US-amy-low.onnx", Tokens: "tokens.txt", DataDir: "espeak-ng-data"},
		"kitten-nano-en-v0_8-int8":    {Family: Kitten, Model: "model.int8.onnx", Voices: "voices.bin", Tokens: "tokens.txt", DataDir: "espeak-ng-data"},
		"kokoro-int8-multi-lang-v1_1": {Family: Kokoro, Model: "model.int8.onnx", Voices: "voices.bin", Tokens: "tokens.txt", DataDir: "espeak-ng-data", Lexicon: "lexicon-us-en.txt", DictDir: "dict"},
	}
	for dir, cfg := range cases {
		base := filepath.Join(root, dir)
		if _, err := os.Stat(base); err != nil {
			t.Logf("skip %s", dir)
			continue
		}
		for _, f := range []*string{&cfg.Model, &cfg.Voices, &cfg.Tokens, &cfg.DataDir, &cfg.Lexicon, &cfg.DictDir} {
			if *f != "" {
				*f = filepath.Join(base, *f)
			}
		}
		tts, err := New(cfg)
		if err != nil {
			t.Fatalf("%s: %v", dir, err)
		}
		var last float32
		samples, rate, err := tts.Generate("Hello from Sonora. This runs on the device.", 0, 1, func(p float32) bool { last = p; return true })
		if err != nil || rate <= 0 || len(samples) < rate {
			t.Fatalf("%s: %d samples at %d Hz, %v", dir, len(samples), rate, err)
		}
		t.Logf("%s: %.1fs at %d Hz, %d speakers, last progress %.2f", dir, float64(len(samples))/float64(rate), rate, tts.NumSpeakers(), last)
		// Stopping from the progress callback ends generation.
		if _, _, err := tts.Generate("One. Two. Three. Four.", 0, 1, func(float32) bool { return false }); err != ErrStopped {
			t.Fatalf("%s: stop: %v", dir, err)
		}
		tts.Close()
	}
}
