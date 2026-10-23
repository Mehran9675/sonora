package openai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/mehran9675/sonora/internal/audio/convert"
	"github.com/mehran9675/sonora/internal/engines"
)

// Engine is one configured API endpoint. Its setup options come from
// settings through config.
type Engine struct {
	m      *engines.Manifest
	config func() engines.Values
	client *http.Client
}

// New returns the engine; config returns its current option values.
func New(m *engines.Manifest, config func() engines.Values) *Engine {
	return &Engine{m: m, config: config, client: &http.Client{Timeout: 10 * time.Minute}}
}

func (e *Engine) Manifest() *engines.Manifest                   { return e.m }
func (e *Engine) Installed() bool                               { return true }
func (e *Engine) Install(context.Context, engines.Logger) error { return nil }
func (e *Engine) Uninstall() error                              { return nil }
func (e *Engine) Close()                                        {}
func (e *Engine) Loaded() bool                                  { return false }

func (e *Engine) url(path string) string {
	return strings.TrimRight(e.config().String("base_url"), "/") + path
}

func (e *Engine) do(req *http.Request) (*http.Response, error) {
	if key := e.config().String("api_key"); key != "" {
		req.Header.Set("Authorization", "Bearer "+key)
	}
	return e.client.Do(req)
}

// Voices asks the server (GET /audio/voices, offered by several local
// servers) and falls back to the configured list.
func (e *Engine) Voices(ctx context.Context, _ engines.Values, log engines.Logger) ([]engines.Voice, error) {
	if vs := e.serverVoices(ctx); len(vs) > 0 {
		return vs, nil
	}
	var out []engines.Voice
	for _, v := range strings.Split(e.config().String("voices"), ",") {
		if v = strings.TrimSpace(v); v != "" {
			out = append(out, engines.Voice{ID: v, Name: v})
		}
	}
	return out, nil
}

func (e *Engine) serverVoices(ctx context.Context) []engines.Voice {
	ctx, cancel := context.WithTimeout(ctx, 4*time.Second)
	defer cancel()
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, e.url("/audio/voices"), nil)
	resp, err := e.do(req)
	if err != nil {
		return nil
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil
	}
	var body struct {
		Voices []json.RawMessage `json:"voices"`
	}
	if json.NewDecoder(resp.Body).Decode(&body) != nil {
		return nil
	}
	var out []engines.Voice
	for _, raw := range body.Voices {
		var name string
		var obj struct{ ID, Name string }
		if json.Unmarshal(raw, &name) == nil {
			out = append(out, engines.Voice{ID: name, Name: name})
		} else if json.Unmarshal(raw, &obj) == nil && (obj.ID != "" || obj.Name != "") {
			id := obj.ID
			if id == "" {
				id = obj.Name
			}
			out = append(out, engines.Voice{ID: id, Name: obj.Name})
		}
	}
	return out
}

// Synthesize posts the text and saves the returned audio as WAV.
func (e *Engine) Synthesize(ctx context.Context, r engines.Request, log engines.Logger) error {
	cfg := e.config()
	body := map[string]any{
		"model":           cfg.String("model"),
		"input":           r.Text,
		"voice":           r.Voice,
		"response_format": "wav",
	}
	if r.Speed > 0 && r.Speed != 1 {
		body["speed"] = r.Speed
	}
	if s := r.Options.String("instructions"); s != "" {
		body["instructions"] = s
	}
	data, _ := json.Marshal(body)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, e.url("/audio/speech"), bytes.NewReader(data))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	log("POST " + e.url("/audio/speech"))
	resp, err := e.do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	audioBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusOK {
		msg := strings.TrimSpace(string(audioBytes))
		if len(msg) > 400 {
			msg = msg[:400] + "…"
		}
		return fmt.Errorf("server replied %s: %s", resp.Status, msg)
	}
	return convert.SaveAsWAV(r.Output, audioBytes)
}
