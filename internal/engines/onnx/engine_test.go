//go:build sherpa

package onnx

import (
	"context"
	"os"
	"path/filepath"
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
	for _, e := range All(root) {
		e := e.(*Engine)
		t.Run(e.m.ID, func(t *testing.T) {
			log := func(line string) { t.Log(line) }
			if err := e.Install(context.Background(), log); err != nil {
				t.Fatal(err)
			}
			if !e.Installed() {
				t.Fatal("not installed after Install")
			}
			out := filepath.Join(t.TempDir(), "out.wav")
			req := engines.Request{Text: "Hello from Sonora, running on this device.", Voice: e.m.DefaultVoice,
				Speed: 1, Options: engines.Resolve(e.m.Options, nil), Output: out}
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
