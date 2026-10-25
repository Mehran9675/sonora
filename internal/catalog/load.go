package catalog

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/mehran9675/sonora/internal/engines"
)

// order puts the headline engines first on the Engines page.
var order = map[string]int{
	"kitten-tts-2": 1, "audio8": 2, "indextts": 3, "kitten-tts-lite": 4, "kokoro": 5,
	"chatterbox": 6, "piper": 7, "coqui": 8, "transformers": 9,
	"faster-whisper": 20, "transformers-asr": 21,
}

// LoadDir reads every <dir>/<id>/engine.json. Broken manifests are
// reported in errs instead of failing the whole catalog.
func LoadDir(dir string) (ms []*engines.Manifest, errs []error) {
	entries, _ := os.ReadDir(dir)
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		folder := filepath.Join(dir, e.Name())
		m, err := Read(folder)
		if err != nil {
			errs = append(errs, err)
			continue
		}
		ms = append(ms, m)
	}
	sort.SliceStable(ms, func(i, j int) bool {
		oi, oj := order[ms[i].ID], order[ms[j].ID]
		if oi == 0 {
			oi = 100
		}
		if oj == 0 {
			oj = 100
		}
		if oi != oj {
			return oi < oj
		}
		return ms[i].Name < ms[j].Name
	})
	return ms, errs
}

// Read parses folder/engine.json.
func Read(folder string) (*engines.Manifest, error) {
	data, err := os.ReadFile(filepath.Join(folder, "engine.json"))
	if err != nil {
		return nil, err
	}
	var m engines.Manifest
	if err := json.Unmarshal(data, &m); err != nil {
		return nil, fmt.Errorf("%s: %w", folder, err)
	}
	if m.ID == "" {
		m.ID = filepath.Base(folder)
	}
	if m.Kind == "" {
		m.Kind = engines.KindPython
	}
	if m.Kind == engines.KindPython {
		if _, err := os.Stat(filepath.Join(folder, "adapter.py")); err != nil {
			return nil, fmt.Errorf("%s: adapter.py is missing", folder)
		}
	}
	m.Dir = folder
	return &m, nil
}
