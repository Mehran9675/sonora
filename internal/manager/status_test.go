package manager

import "testing"

func TestTidy(t *testing.T) {
	cases := []struct {
		in   string
		want string
		pct  float32
	}{
		{"audio8_tts:  40%|████      | 2/5 [00:35<00:52, 17.6s/it]", "audio8_tts · 40% · 2/5 · 00:52 left", 0.4},
		{"model.safetensors:  45%|████▌     | 450M/1.0G [00:10<00:12, 45MB/s]", "Downloading model.safetensors · 45% · 450M/1.0G · 00:12 left", 0.45},
		{"100%|██████████| 3/3 [00:01<00:00,  2.1it/s]", "Working · 100% · 3/3", 1},
		{"Downloading… 35%", "Downloading… 35%", 0.35},
		{"Loading model", "Loading model", -1},
	}
	for _, c := range cases {
		got, pct := tidy(c.in)
		if got != c.want || pct != c.pct {
			t.Errorf("tidy(%q) = %q, %v; want %q, %v", c.in, got, pct, c.want, c.pct)
		}
	}
}

func TestAppendLineCollapsesProgress(t *testing.T) {
	var log []string
	for _, l := range []string{"start", "a.bin:  10%|█ | 1/10", "a.bin:  50%|█████ | 5/10", "b.bin:  10%|█ | 1/10", "done"} {
		log = appendLine(log, l)
	}
	want := []string{"start", "a.bin:  50%|█████ | 5/10", "b.bin:  10%|█ | 1/10", "done"}
	if len(log) != len(want) {
		t.Fatalf("got %q", log)
	}
	for i := range want {
		if log[i] != want[i] {
			t.Fatalf("got %q", log)
		}
	}
}
