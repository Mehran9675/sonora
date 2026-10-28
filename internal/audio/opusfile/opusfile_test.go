package opusfile

import (
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"
)

// tone writes secs seconds of a 440 Hz tone at rate in uneven chunks,
// as a microphone delivers them.
func tone(t *testing.T, path string, rate, bitrate int, secs float64) {
	w, err := Create(path, rate, bitrate)
	if err != nil {
		t.Fatal(err)
	}
	n := int(secs * float64(rate))
	for i := 0; i < n; {
		chunk := make([]float32, min(n-i, 333))
		for j := range chunk {
			chunk[j] = 0.5 * float32(math.Sin(2*math.Pi*440*float64(i+j)/float64(rate)))
		}
		if err := w.Write(chunk); err != nil {
			t.Fatal(err)
		}
		i += len(chunk)
	}
	if err := w.Close(); err != nil {
		t.Fatal(err)
	}
}

func TestRoundTrip(t *testing.T) {
	for _, rate := range []int{48000, 16000} {
		path := filepath.Join(t.TempDir(), "t.opus")
		tone(t, path, rate, 32000, 3.25)
		st, _ := os.Stat(path)
		f, _ := os.Open(path)
		head := make([]byte, 64)
		f.Read(head)
		f.Seek(0, 0)
		if !IsOpus(head) {
			t.Fatal("not recognized as Opus")
		}
		var got []float32
		if err := Decode(f, func(r int, m []float32) error {
			if r != 48000 {
				t.Fatalf("rate %d", r)
			}
			got = append(got, m...)
			return nil
		}); err != nil {
			t.Fatal(err)
		}
		f.Close()
		secs := float64(len(got)) / 48000
		crossings := 0
		for i := 48000; i < len(got)-4800; i++ { // skip the encoder's warm-up
			if (got[i-1] < 0) != (got[i] < 0) {
				crossings++
			}
		}
		hz := float64(crossings) / 2 / (float64(len(got)-52800) / 48000)
		t.Logf("%d Hz input: %d bytes (%.1f kbps), %.3fs decoded, %.0f Hz", rate, st.Size(), float64(st.Size()*8)/3.25/1000, secs, hz)
		if d, err := Duration(path); err != nil || math.Abs(d.Seconds()-3.25) > 0.001 {
			t.Errorf("Duration: %v %v", d, err)
		}
		if secs < 3.24 || secs > 3.30 {
			t.Errorf("duration %.3fs", secs)
		}
		if math.Abs(hz-440) > 5 {
			t.Errorf("pitch %.0f Hz", hz)
		}
		// An independent decoder agrees, when ffprobe is installed.
		if _, err := exec.LookPath("ffprobe"); err == nil {
			out, err := exec.Command("ffprobe", "-v", "error", "-show_entries", "format=duration", "-of", "csv=p=0", path).Output()
			d, _ := strconv.ParseFloat(strings.TrimSpace(string(out)), 64)
			if err != nil || math.Abs(d-3.25) > 0.03 {
				t.Errorf("ffprobe: %q %v", out, err)
			}
			if out, err := exec.Command("ffmpeg", "-v", "error", "-i", path, "-f", "null", "-").CombinedOutput(); err != nil || len(out) > 0 {
				t.Errorf("ffmpeg decode: %s %v", out, err)
			}
		}
	}
}

// Low bitrates use the speech (SILK) and hybrid modes; every quality
// setting must round-trip.
func TestBitrates(t *testing.T) {
	for _, br := range []int{12000, 16000, 24000, 64000, 128000} {
		path := filepath.Join(t.TempDir(), "t.opus")
		start := time.Now()
		tone(t, path, 48000, br, 10)
		took := time.Since(start)
		f, _ := os.Open(path)
		n := 0
		if err := Decode(f, func(_ int, m []float32) error { n += len(m); return nil }); err != nil {
			t.Fatalf("%d bps: %v", br, err)
		}
		f.Close()
		st, _ := os.Stat(path)
		t.Logf("%3d kbps: %6d bytes for 10 s, encoded %.0fx real time, %.2fs decoded", br/1000, st.Size(), 10/took.Seconds(), float64(n)/48000)
		if math.Abs(float64(n)/48000-10) > 0.01 {
			t.Errorf("%d bps: decoded %d samples", br, n)
		}
	}
}
