package openai

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/mehran9675/sonora/internal/engines"
)

func TestTranscribe(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/audio/transcriptions" || r.Header.Get("Authorization") != "Bearer k" {
			http.Error(w, "bad request "+r.URL.Path, http.StatusBadRequest)
			return
		}
		r.ParseMultipartForm(1 << 20)
		if r.FormValue("model") != "whisper-1" || r.FormValue("language") != "de" || r.FormValue("response_format") != "verbose_json" {
			http.Error(w, "bad form", http.StatusBadRequest)
			return
		}
		w.Write([]byte(`{"text":" Hallo Welt ","language":"german","segments":[{"start":0,"end":1.5,"text":" Hallo Welt"}]}`))
	}))
	defer srv.Close()
	audio := filepath.Join(t.TempDir(), "a.wav")
	os.WriteFile(audio, []byte("RIFF"), 0o644)
	m := TranscribeTemplate("x", "X")
	e := NewTranscriber(m, func() engines.Values {
		return engines.Values{"base_url": srv.URL + "/v1", "api_key": "k", "model": "whisper-1"}
	})
	got, err := e.Transcribe(context.Background(), engines.TranscribeRequest{Audio: audio, Language: "de"}, func(string) {})
	if err != nil {
		t.Fatal(err)
	}
	if got.Text != "Hallo Welt" || got.Language != "de" || len(got.Segments) != 1 || got.Segments[0].Text != "Hallo Welt" {
		t.Fatalf("%+v", got)
	}
}
