// Package openai talks to any server implementing OpenAI's
// /v1/audio/speech endpoint - OpenAI itself, and local servers such as
// Kokoro-FastAPI, LocalAI, Audio8's server or KittenTTS's Docker image.
package openai

import "github.com/mehran9675/sonora/internal/engines"

// Template returns the manifest for a new API engine.
func Template(id, name string) *engines.Manifest {
	return &engines.Manifest{
		ID:          id,
		Name:        name,
		Vendor:      "OpenAI-compatible API",
		Description: "Speaks through an OpenAI-compatible /v1/audio/speech endpoint - a local server or a cloud service.",
		Tags:        []string{"API", "Remote"},
		Accent:      []string{"#94a3b8", "#475569"},
		Kind:        engines.KindOpenAI,
		Speed:       true,
		Custom:      true,
		Options: []engines.Option{
			{ID: "base_url", Label: "Server URL", Type: engines.OptionText, Scope: engines.ScopeSetup,
				Default: "http://localhost:8880/v1", Placeholder: "http://localhost:8880/v1",
				Help: "Everything before /audio/speech"},
			{ID: "api_key", Label: "API key", Type: engines.OptionSecret, Scope: engines.ScopeSetup, Placeholder: "optional for local servers"},
			{ID: "model", Label: "Model", Type: engines.OptionText, Scope: engines.ScopeSetup, Default: "tts-1",
				Suggestions: []string{"tts-1", "tts-1-hd", "gpt-4o-mini-tts", "kokoro"}},
			{ID: "voices", Label: "Voices", Type: engines.OptionText, Scope: engines.ScopeSetup,
				Default: "alloy, echo, fable, onyx, nova, shimmer", Help: "Comma-separated; the server's own list is used when it offers one"},
			{ID: "instructions", Label: "Instructions", Type: engines.OptionText, Placeholder: "e.g. speak warmly and slowly",
				Help: "Sent to models that accept style instructions"},
		},
	}
}
