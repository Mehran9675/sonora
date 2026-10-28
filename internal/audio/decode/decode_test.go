package decode

import (
	"errors"
	"math"
	"path/filepath"
	"testing"

	"github.com/mehran9675/sonora/internal/audio/convert"
)

// TestFormats decodes a one-second 440 Hz tone in each format and checks
// length, level and pitch at 16 kHz. The tone is ffmpeg's sine at -6 dB,
// RMS 0.044; the stereo files carry it 3 dB lower in each channel.
func TestFormats(t *testing.T) {
	for _, name := range []string{"tone8k.wav", "tone.mp3", "tone.flac", "tone.ogg", "tone.opus", "tone.m4a"} {
		t.Run(name, func(t *testing.T) {
			clip, err := File(filepath.Join("testdata", name), SpeechRate)
			if errors.Is(err, ErrFormat) && !convert.Available() {
				t.Skip("needs ffmpeg")
			}
			if err != nil {
				t.Fatal(err)
			}
			if clip.SampleRate != SpeechRate {
				t.Fatalf("rate %d", clip.SampleRate)
			}
			if secs := clip.Duration().Seconds(); secs < 0.95 || secs > 1.1 {
				t.Fatalf("duration %.3fs", secs)
			}
			// MP3 pads both ends with silence; measure the tone itself.
			first, last := -1, 0
			for i, s := range clip.Samples {
				if math.Abs(float64(s)) > 0.01 {
					if first < 0 {
						first = i
					}
					last = i
				}
			}
			tone := clip.Samples[first : last+1]
			var sum float64
			crossings := 0
			for i, s := range tone {
				sum += float64(s) * float64(s)
				if i > 0 && (tone[i-1] < 0) != (s < 0) {
					crossings++
				}
			}
			rms := math.Sqrt(sum / float64(len(tone)))
			if rms < 0.028 || rms > 0.047 {
				t.Fatalf("rms %.3f, want 0.031–0.044", rms)
			}
			hz := float64(crossings) / 2 / (float64(len(tone)) / SpeechRate)
			if math.Abs(hz-440) > 10 {
				t.Fatalf("pitch %.0f Hz", hz)
			}
		})
	}
}

func TestResampleUp(t *testing.T) {
	r := newResampler(16000)
	r.write(8000, []float32{0, 1, 0, -1})
	got := r.clip().Samples
	want := []float32{0, 0.5, 1, 0.5, 0, -0.5, -1}
	if len(got) != len(want) {
		t.Fatalf("got %v", got)
	}
	for i := range want {
		if math.Abs(float64(got[i]-want[i])) > 1e-6 {
			t.Fatalf("got %v", got)
		}
	}
}
