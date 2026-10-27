// Package settings persists the user's choices to settings.json in the
// data folder.
package settings

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"

	"github.com/mehran9675/sonora/internal/engines"
	"github.com/mehran9675/sonora/internal/paths"
)

// Settings is everything Sonora remembers between runs.
type Settings struct {
	Theme   string `json:"theme"`  // "system", "dark" or "light"
	Device  string `json:"device"` // "auto", "cpu" or "gpu"
	HFToken string `json:"hf_token,omitempty"`

	Engine string                    `json:"engine"`
	Voices map[string]string         `json:"voices"` // engine id → voice id
	Values map[string]engines.Values `json:"values"` // engine id → option values
	Speed  map[string]float64        `json:"speed"`  // engine id → speed
	Custom []*engines.Manifest       `json:"custom"` // user-added API/command engines
	Draft  string                    `json:"draft"`  // the editor's text

	// Transcription: the engine, the spoken language ("" detects it)
	// and whether to translate into English.
	TranscribeEngine string `json:"transcribe_engine,omitempty"`
	Language         string `json:"language,omitempty"`
	Translate        bool   `json:"translate,omitempty"`

	// Recording: quality ("small", "standard", "high", "best"; empty is
	// standard), the folder ("" is recordings/ in the data folder) and
	// the longest a recording runs, in minutes (0 is an hour).
	RecordQuality string `json:"record_quality,omitempty"`
	RecordDir     string `json:"record_dir,omitempty"`
	RecordMinutes int    `json:"record_minutes,omitempty"`
}

// Store loads and saves Settings. Methods are safe from any goroutine.
type Store struct {
	mu   sync.Mutex
	path string
	s    Settings
}

// Open reads settings.json, falling back to defaults.
func Open() *Store {
	st := &Store{path: filepath.Join(paths.Data(), "settings.json")}
	st.s = Settings{Theme: "system", Device: "auto", Engine: "system"}
	if data, err := os.ReadFile(st.path); err == nil {
		json.Unmarshal(data, &st.s)
	}
	if st.s.Voices == nil {
		st.s.Voices = map[string]string{}
	}
	if st.s.Values == nil {
		st.s.Values = map[string]engines.Values{}
	}
	if st.s.Speed == nil {
		st.s.Speed = map[string]float64{}
	}
	return st
}

// Get returns a copy of the current settings (maps are shared; change
// them through Update).
func (st *Store) Get() Settings {
	st.mu.Lock()
	defer st.mu.Unlock()
	return st.s
}

// Update changes the settings and saves them.
func (st *Store) Update(fn func(s *Settings)) {
	st.mu.Lock()
	fn(&st.s)
	data, _ := json.MarshalIndent(st.s, "", "  ")
	st.mu.Unlock()
	tmp := st.path + ".tmp"
	if os.WriteFile(tmp, data, 0o600) == nil {
		os.Rename(tmp, st.path)
	}
}

// Values returns a copy of an engine's option values.
func (st *Store) Values(id string) engines.Values {
	st.mu.Lock()
	defer st.mu.Unlock()
	out := engines.Values{}
	for k, v := range st.s.Values[id] {
		out[k] = v
	}
	return out
}

// SetValue stores one option value.
func (st *Store) SetValue(id, key string, v any) {
	st.Update(func(s *Settings) {
		if s.Values[id] == nil {
			s.Values[id] = engines.Values{}
		}
		s.Values[id][key] = v
	})
}
