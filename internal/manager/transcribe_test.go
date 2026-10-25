//go:build sherpa

package manager

import (
	"os"
	"strings"
	"testing"

	"github.com/mehran9675/sonora/internal/library"
	"github.com/mehran9675/sonora/internal/paths"
	"github.com/mehran9675/sonora/internal/settings"
	"github.com/mehran9675/sonora/internal/transcripts"
)

// TestTranscribe runs a whole transcription - decoding an MP3, the
// on-device Whisper engine, storing the result - when SONORA_HOME
// holds downloaded models and SONORA_TEST_AUDIO names a recording of
// "...lazy dog...".
func TestTranscribe(t *testing.T) {
	audio := os.Getenv("SONORA_TEST_AUDIO")
	if audio == "" || os.Getenv("SONORA_HOME") == "" {
		t.Skip("set SONORA_HOME and SONORA_TEST_AUDIO")
	}
	posted := make(chan func(), 100)
	ts := transcripts.Open(paths.Sub("transcripts"))
	m := New(func(f func()) { posted <- f }, settings.Open(), library.Open(paths.Sub("library")), ts)
	rt := m.Get("ondevice-whisper")
	if rt == nil || rt.State != StateReady {
		t.Fatal("on-device Whisper isn't installed")
	}
	done := make(chan error, 1)
	m.Transcribe(rt, TranscribeInput{Source: audio}, func(e transcripts.Entry, err error) {
		if err == nil && !strings.Contains(strings.ToLower(e.Text), "lazy dog") {
			t.Errorf("transcript %q", e.Text)
		}
		done <- err
	})
	for {
		select {
		case f := <-posted:
			f()
		case err := <-done:
			if err != nil {
				t.Fatal(err)
			}
			es := ts.Entries()
			if len(es) == 0 || es[0].Duration.Seconds() < 3 {
				t.Fatalf("stored %+v", es)
			}
			if _, err := os.Stat(ts.AudioPath(es[0])); err != nil {
				t.Fatal(err)
			}
			if m.TranscribeJob != nil || rt.Busy() {
				t.Fatal("job not cleared")
			}
			return
		}
	}
}
