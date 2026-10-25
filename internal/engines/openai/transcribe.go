package openai

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/mehran9675/sonora/internal/engines"
)

// TranscribeTemplate returns the manifest for a new transcription API
// engine.
func TranscribeTemplate(id, name string) *engines.Manifest {
	return &engines.Manifest{
		ID:          id,
		Name:        name,
		Vendor:      "OpenAI-compatible API",
		Description: "Transcribes through an OpenAI-compatible /v1/audio/transcriptions endpoint - OpenAI, Groq, a local Whisper server…",
		Tags:        []string{"API", "Remote"},
		Accent:      []string{"#94a3b8", "#475569"},
		Kind:        engines.KindOpenAITranscribe,
		Task:        engines.TaskTranscribe,
		Translate:   true,
		Custom:      true,
		Options: []engines.Option{
			{ID: "base_url", Label: "Server URL", Type: engines.OptionText, Scope: engines.ScopeSetup,
				Default: "https://api.openai.com/v1", Placeholder: "https://api.openai.com/v1",
				Help: "Everything before /audio/transcriptions", Keyboard: "url"},
			{ID: "api_key", Label: "API key", Type: engines.OptionSecret, Scope: engines.ScopeSetup, Placeholder: "optional for local servers"},
			{ID: "model", Label: "Model", Type: engines.OptionText, Scope: engines.ScopeSetup, Default: "whisper-1",
				Suggestions: []string{"whisper-1", "gpt-4o-transcribe", "gpt-4o-mini-transcribe", "whisper-large-v3-turbo", "Systran/faster-whisper-large-v3"}},
			{ID: "prompt", Label: "Vocabulary hint", Type: engines.OptionText, Placeholder: "Names, jargon, spelling…"},
		},
	}
}

// Transcriber is one configured transcription endpoint.
type Transcriber struct {
	*Engine
}

// NewTranscriber returns the engine; config returns its option values.
func NewTranscriber(m *engines.Manifest, config func() engines.Values) *Transcriber {
	return &Transcriber{&Engine{m: m, config: config, client: &http.Client{Timeout: 30 * time.Minute}}}
}

func (e *Transcriber) Voices(context.Context, engines.Values, engines.Logger) ([]engines.Voice, error) {
	return nil, nil
}

func (e *Transcriber) Synthesize(context.Context, engines.Request, engines.Logger) error {
	return errors.New(e.m.Name + " turns speech into text; it can't speak")
}

// maxUpload is OpenAI's file size limit.
const maxUpload = 25 << 20

// Transcribe uploads the recording and parses the verbose JSON reply;
// servers that ignore response_format get a plain-text fallback.
func (e *Transcriber) Transcribe(ctx context.Context, r engines.TranscribeRequest, log engines.Logger) (*engines.Transcript, error) {
	cfg := e.config()
	audio, err := os.ReadFile(r.Audio)
	if err != nil {
		return nil, err
	}
	if len(audio) > maxUpload {
		return nil, fmt.Errorf("the recording is %d MB once converted; most APIs take at most 25 MB (about 13 minutes), so use an on-device or Python engine for longer files", len(audio)>>20)
	}
	path := "/audio/transcriptions"
	if r.Translate {
		path = "/audio/translations"
	}
	var body bytes.Buffer
	mw := multipart.NewWriter(&body)
	fw, _ := mw.CreateFormFile("file", filepath.Base(r.Audio))
	fw.Write(audio)
	model := cfg.String("model")
	mw.WriteField("model", model)
	// The gpt-4o transcription models only return plain JSON.
	verbose := !strings.HasPrefix(model, "gpt-4o")
	if verbose {
		mw.WriteField("response_format", "verbose_json")
		mw.WriteField("timestamp_granularities[]", "segment")
	} else {
		mw.WriteField("response_format", "json")
	}
	if r.Language != "" && !r.Translate {
		mw.WriteField("language", r.Language)
	}
	if p := r.Options.String("prompt"); p != "" {
		mw.WriteField("prompt", p)
	}
	mw.Close()
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, e.url(path), &body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", mw.FormDataContentType())
	log("Uploading to " + e.url(path) + "…")
	resp, err := e.do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		msg := strings.TrimSpace(string(data))
		if len(msg) > 400 {
			msg = msg[:400] + "…"
		}
		return nil, fmt.Errorf("server replied %s: %s", resp.Status, msg)
	}
	var reply struct {
		Text     string            `json:"text"`
		Language string            `json:"language"`
		Segments []engines.Segment `json:"segments"`
	}
	if json.Unmarshal(data, &reply) != nil {
		reply.Text = string(data)
	}
	for i := range reply.Segments {
		reply.Segments[i].Text = strings.TrimSpace(reply.Segments[i].Text)
	}
	return &engines.Transcript{Text: strings.TrimSpace(reply.Text), Language: languageCode(reply.Language), Segments: reply.Segments}, nil
}

// languageCode turns the language name OpenAI reports ("english") into
// its code.
func languageCode(name string) string {
	for _, l := range engines.Languages {
		if l.Value != "" && strings.EqualFold(l.Label, name) {
			return l.Value
		}
	}
	return name
}
