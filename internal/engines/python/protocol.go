package python

import "encoding/json"

// request is a message to the worker.
type request struct {
	ID      int            `json:"id"`
	Cmd     string         `json:"cmd"`
	Options map[string]any `json:"options,omitempty"`
	Request any            `json:"request,omitempty"`
}

// message is a reply (ID set) or an event from the worker.
type message struct {
	ID      int             `json:"id"`
	OK      bool            `json:"ok"`
	Result  json.RawMessage `json:"result"`
	Error   string          `json:"error"`
	Event   string          `json:"event"`
	Message string          `json:"message"`
	Value   float64         `json:"value"`
}

// loadResult is what an adapter's load() returns.
type loadResult struct {
	Voices []struct {
		ID          string `json:"id"`
		Name        string `json:"name"`
		Group       string `json:"group"`
		Description string `json:"description"`
	} `json:"voices"`
}
