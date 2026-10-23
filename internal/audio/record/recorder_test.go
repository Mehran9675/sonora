package record

import "testing"

func TestTrimSilence(t *testing.T) {
	rate := 1000
	s := make([]float32, 3000)
	for i := 1000; i < 2000; i++ {
		s[i] = 0.5
	}
	got := trimSilence(s, rate)
	// 1000 loud samples plus 100 samples of padding on each side.
	if len(got) != 1199 {
		t.Fatalf("got %d samples", len(got))
	}
	if quiet := make([]float32, 50); len(trimSilence(quiet, rate)) != 50 {
		t.Fatal("an all-quiet recording is kept as is")
	}
}

func TestPushLevels(t *testing.T) {
	r := &Recorder{}
	loud := make([]byte, 400)
	for i := 0; i < len(loud); i += 2 {
		loud[i], loud[i+1] = 0x00, 0x40 // 0.5 full scale
	}
	r.push(loud)
	if lv := r.Levels(); len(lv) != 1 || lv[0] < 0.8 {
		t.Fatalf("levels %v", lv)
	}
	if len(r.samples) != 200 {
		t.Fatalf("samples %d", len(r.samples))
	}
}
