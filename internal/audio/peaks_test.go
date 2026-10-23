package audio

import "testing"

func TestPeaksNormalized(t *testing.T) {
	c := &Clip{SampleRate: 10, Samples: []float32{0, 0.1, -0.2, 0.4, 0, 0, 0.8, -0.8}}
	p := c.Peaks(4)
	if len(p) != 4 {
		t.Fatalf("got %d peaks", len(p))
	}
	var top float32
	for _, v := range p {
		top = max(top, v)
	}
	if top != 1 {
		t.Fatalf("loudest bucket is %f, want 1", top)
	}
}

func TestResample(t *testing.T) {
	c := &Clip{SampleRate: 24000, Samples: make([]float32, 24000)}
	if n := len(c.Resample(48000)); n != 48000 {
		t.Fatalf("got %d samples", n)
	}
}
