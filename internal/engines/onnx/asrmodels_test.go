package onnx

import "testing"

// sherpa-onnx ends the process on a language a model doesn't know, so
// only known hints may reach it.
func TestSupportedLanguage(t *testing.T) {
	tiny, turbo := whisperModels[0], whisperModels[len(whisperModels)-1]
	cases := []struct {
		m          model
		lang, want string
	}{
		{tiny, "de", "de"}, {tiny, "yue", ""}, {turbo, "yue", "yue"}, {tiny, "xx", ""}, {tiny, "", ""},
		{senseVoiceModel, "ja", "ja"}, {senseVoiceModel, "de", ""}, {parakeetModel, "de", ""},
	}
	for _, c := range cases {
		if got := supportedLanguage(c.m, c.lang); got != c.want {
			t.Errorf("%s %q: got %q, want %q", c.m.ID, c.lang, got, c.want)
		}
	}
}
