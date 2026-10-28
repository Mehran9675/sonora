package record

import (
	"math"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/mehran9675/sonora/internal/audio/opusfile"
)

// pcm returns n samples of a 440 Hz tone at half scale as the device
// delivers them.
func pcm(start, n int) []byte {
	b := make([]byte, 2*n)
	for i := 0; i < n; i++ {
		v := int16(16384 * math.Sin(2*math.Pi*440*float64(start+i)/SampleRate))
		b[2*i], b[2*i+1] = byte(v), byte(uint16(v)>>8)
	}
	return b
}

// TestRecordToFile feeds two seconds through the recorder, as a
// microphone would, and reads the Opus file back.
func TestRecordToFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "sub", "take.opus")
	r, err := newRecorder(Options{Path: path, Bitrate: 32000})
	if err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 2*SampleRate; i += 480 { // 10 ms blocks
		r.push(pcm(i, 480))
	}
	if lv := r.Levels(); len(lv) == 0 || lv[len(lv)-1] < 0.8 {
		t.Fatalf("levels %v", lv)
	}
	if got := r.Elapsed(); got != 2*time.Second {
		t.Fatalf("elapsed %v", got)
	}
	out, length, err := r.Stop()
	if err != nil || out != path || length != 2*time.Second {
		t.Fatalf("stop: %q %v %v", out, length, err)
	}
	f, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	n := 0
	if err := opusfile.Decode(f, func(_ int, m []float32) error { n += len(m); return nil }); err != nil {
		t.Fatal(err)
	}
	if secs := float64(n) / 48000; math.Abs(secs-2) > 0.01 {
		t.Fatalf("decoded %.3fs", secs)
	}
}

func TestCancelDeletes(t *testing.T) {
	path := filepath.Join(t.TempDir(), "take.opus")
	r, err := newRecorder(Options{Path: path, Bitrate: 16000})
	if err != nil {
		t.Fatal(err)
	}
	r.push(pcm(0, 4800))
	r.Cancel()
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Fatal("cancelled recording kept")
	}
}
