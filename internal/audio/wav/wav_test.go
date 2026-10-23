package wav

import (
	"math"
	"path/filepath"
	"testing"

	"github.com/mehran9675/sonora/internal/audio"
)

func TestRoundTrip(t *testing.T) {
	in := &audio.Clip{SampleRate: 24000}
	for i := 0; i < 2400; i++ {
		in.Samples = append(in.Samples, float32(0.5*math.Sin(float64(i)/10)))
	}
	path := filepath.Join(t.TempDir(), "a.wav")
	if err := Write(path, in); err != nil {
		t.Fatal(err)
	}
	out, err := Read(path)
	if err != nil {
		t.Fatal(err)
	}
	if out.SampleRate != 24000 || len(out.Samples) != len(in.Samples) {
		t.Fatalf("got %d Hz × %d samples", out.SampleRate, len(out.Samples))
	}
	for i := range in.Samples {
		if d := math.Abs(float64(in.Samples[i] - out.Samples[i])); d > 1e-3 {
			t.Fatalf("sample %d differs by %f", i, d)
		}
	}
	if got := out.Duration().Milliseconds(); got != 100 {
		t.Fatalf("duration %d ms", got)
	}
}

func TestRejectsNonWAV(t *testing.T) {
	if _, err := Decode([]byte("ID3 not a wav file")); err == nil {
		t.Fatal("expected an error")
	}
}
