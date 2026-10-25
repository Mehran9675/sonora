//go:build sherpa

package onnx

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/mehran9675/sonora/internal/audio/wav"
	"github.com/mehran9675/sonora/internal/engines"
)

// TestEngines downloads (or reuses) each engine's default model and
// generates a clip. Set SONORA_ONDEVICE_TEST to a folder for the
// downloads; it needs network access the first time.
func TestEngines(t *testing.T) {
	root := os.Getenv("SONORA_ONDEVICE_TEST")
	if root == "" {
		t.Skip("set SONORA_ONDEVICE_TEST to a download folder")
	}
	// Cloning engines copy the voice of the clip made before them.
	const text = "Hello from Sonora, running on this device."
	var reference string
	for _, e := range All(root) {
		e, ok := e.(*Engine)
		if !ok {
			continue
		}
		t.Run(e.m.ID, func(t *testing.T) {
			log := func(line string) { t.Log(line) }
			if err := e.Install(context.Background(), log); err != nil {
				t.Fatal(err)
			}
			if !e.Installed() {
				t.Fatal("not installed after Install")
			}
			out := filepath.Join(root, e.m.ID+".wav")
			req := engines.Request{Text: text, Voice: e.m.DefaultVoice,
				Speed: 1, Options: engines.Resolve(e.m.Options, nil), Output: out}
			if e.m.Clone == engines.CloneRequired {
				if reference == "" {
					t.Skip("no clip to clone")
				}
				req.Reference, req.ReferenceText = reference, text
			}
			if err := e.Synthesize(context.Background(), req, func(string) {}); err != nil {
				t.Fatal(err)
			}
			clip, err := wav.Read(out)
			if err != nil {
				t.Fatal(err)
			}
			if secs := float64(len(clip.Samples)) / float64(clip.SampleRate); secs < 1 {
				t.Fatalf("clip is %.2fs", secs)
			}
			if !e.Loaded() {
				t.Fatal("model not loaded after Synthesize")
			}
			if e.m.Clone == engines.CloneNone {
				reference = out
			}
			ctx, cancel := context.WithCancel(context.Background())
			cancel()
			if err := e.Synthesize(ctx, req, func(string) {}); err == nil {
				t.Fatal("cancelled request succeeded")
			}
			e.Close()
			if e.Loaded() {
				t.Fatal("still loaded after Close")
			}
		})
	}
}

// TestRecognizers speaks a sentence with Piper and transcribes it with
// each speech-to-text engine's default model.
func TestRecognizers(t *testing.T) {
	root := os.Getenv("SONORA_ONDEVICE_TEST")
	if root == "" {
		t.Skip("set SONORA_ONDEVICE_TEST to a download folder")
	}
	all := All(root)
	var piper *Engine
	for _, e := range all {
		if e.Manifest().ID == "ondevice-piper" {
			piper = e.(*Engine)
		}
	}
	speech := filepath.Join(root, "speech.wav")
	text := "The quick brown fox jumps over the lazy dog. It was a sunny morning in the city."
	req := engines.Request{Text: text, Voice: piperVoices[0].ID, Speed: 1, Output: speech}
	if err := piper.Synthesize(context.Background(), req, func(string) {}); err != nil {
		t.Fatal(err)
	}
	clip, err := wav.Read(speech)
	if err != nil {
		t.Fatal(err)
	}
	clip.Samples, clip.SampleRate = clip.Resample(16000), 16000
	audio16k := filepath.Join(root, "speech16k.wav")
	if err := wav.Write(audio16k, clip); err != nil {
		t.Fatal(err)
	}
	for _, e := range all {
		e, ok := e.(*Recognizer)
		if !ok {
			continue
		}
		t.Run(e.m.ID, func(t *testing.T) {
			res, err := e.Transcribe(context.Background(), engines.TranscribeRequest{Audio: audio16k,
				Options: engines.Resolve(e.m.Options, nil)}, func(line string) {})
			if err != nil {
				t.Fatal(err)
			}
			t.Logf("%q (%s, %d segments)", res.Text, res.Language, len(res.Segments))
			lower := strings.ToLower(res.Text)
			for _, w := range []string{"quick", "fox", "dog", "morning"} {
				if !strings.Contains(lower, w) {
					t.Errorf("missing %q", w)
				}
			}
			if len(res.Segments) == 0 || res.Segments[0].End <= res.Segments[0].Start {
				t.Errorf("bad segments %+v", res.Segments)
			}
			ctx, cancel := context.WithCancel(context.Background())
			cancel()
			if _, err := e.Transcribe(ctx, engines.TranscribeRequest{Audio: audio16k}, func(string) {}); err == nil {
				t.Error("cancelled request succeeded")
			}
			e.Close()
		})
	}
}
