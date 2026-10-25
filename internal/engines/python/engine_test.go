package python

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/mehran9675/sonora/internal/audio/wav"
	"github.com/mehran9675/sonora/internal/catalog"
	"github.com/mehran9675/sonora/internal/engines"
)

// adapter is a dependency-free engine: it reports two voices and writes
// a short silent clip whose length depends on the text.
const adapter = `
import wave
class Engine:
    def __init__(self, ctx):
        self.ctx = ctx
    def load(self, options):
        print("noise on stdout must not break the protocol")
        self.ctx.log("loaded " + str(options.get("model")))
        return {"voices": [{"id": "a", "name": "Alpha"}, {"id": "b", "name": "Beta"}]}
    def synthesize(self, req):
        if req["text"] == "fail":
            raise ValueError("asked to fail")
        with wave.open(req["output"], "wb") as w:
            w.setnchannels(1); w.setsampwidth(2); w.setframerate(16000)
            w.writeframes(b"\x00\x00" * 160 * len(req["text"]))
        return {}
    def transcribe(self, req):
        with wave.open(req["audio"]) as w:
            secs = w.getnframes() / w.getframerate()
        return {"text": "heard %.2fs" % secs, "language": req.get("language") or "en",
                "segments": [{"start": 0, "end": secs, "text": "heard"}]}
`

// TestWorkerRoundTrip runs the real worker.py against a fake adapter
// using the system Python in place of an installed environment.
func TestWorkerRoundTrip(t *testing.T) {
	py, err := exec.LookPath("python3")
	if err != nil || runtime.GOOS == "windows" {
		t.Skip("needs python3 on a Unix system")
	}
	data := t.TempDir()
	script, err := catalog.Unpack(filepath.Join(data, "runtime"))
	if err != nil {
		t.Fatal(err)
	}
	engDir := filepath.Join(data, "def")
	os.MkdirAll(engDir, 0o755)
	os.WriteFile(filepath.Join(engDir, "adapter.py"), []byte(adapter), 0o644)
	m := &engines.Manifest{ID: "fake", Name: "Fake", Kind: engines.KindPython, Dir: engDir,
		Options: []engines.Option{{ID: "model", Scope: engines.ScopeLoad, Default: "m1"}}}
	e := New(m, data, script, func() engines.Environment { return engines.Environment{Device: "cpu"} })
	defer e.Close()

	// Stand in for "Install": point venv/bin/python at the system Python.
	os.MkdirAll(filepath.Join(e.venv(), "bin"), 0o755)
	if err := os.Symlink(py, filepath.Join(e.venv(), "bin", "python")); err != nil {
		t.Fatal(err)
	}

	var logs []string
	log := func(l string) { logs = append(logs, l) }
	ctx := context.Background()
	voices, err := e.Voices(ctx, engines.Values{"model": "m1"}, log)
	if err != nil {
		t.Fatal(err)
	}
	if len(voices) != 2 || voices[1].Name != "Beta" {
		t.Fatalf("voices: %+v", voices)
	}
	if !strings.Contains(strings.Join(logs, "\n"), "loaded m1") {
		t.Fatalf("adapter log missing: %q", logs)
	}

	out := filepath.Join(t.TempDir(), "o.wav")
	req := engines.Request{Text: "hello", Voice: "a", Output: out, Options: engines.Values{"model": "m1"}}
	if err := e.Synthesize(ctx, req, log); err != nil {
		t.Fatal(err)
	}
	clip, err := wav.Read(out)
	if err != nil || len(clip.Samples) != 800 {
		t.Fatalf("clip: %v, %v", clip, err)
	}

	tr, err := e.Transcribe(ctx, engines.TranscribeRequest{Audio: out, Language: "de", Options: engines.Values{"model": "m1"}}, log)
	if err != nil {
		t.Fatal(err)
	}
	if tr.Text != "heard 0.05s" || tr.Language != "de" || len(tr.Segments) != 1 || tr.Segments[0].End != 0.05 {
		t.Fatalf("transcript: %+v", tr)
	}

	if !e.Loaded() {
		t.Fatal("the model should be loaded after a request")
	}
	e.Close()
	if e.Loaded() {
		t.Fatal("Close must unload the model")
	}

	req.Text = "fail"
	if err := e.Synthesize(ctx, req, log); err == nil || !strings.Contains(err.Error(), "asked to fail") {
		t.Fatalf("adapter errors must come back: %v", err)
	}
}
