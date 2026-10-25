// Package transcripts keeps every transcription with its audio.
package transcripts

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"time"

	"github.com/mehran9675/sonora/internal/engines"
)

// Entry is one transcribed recording.
type Entry struct {
	ID         string            `json:"id"`
	Title      string            `json:"title"` // the source file's name
	Text       string            `json:"text"`
	Language   string            `json:"language,omitempty"`
	Translated bool              `json:"translated,omitempty"`
	Segments   []engines.Segment `json:"segments,omitempty"`
	EngineID   string            `json:"engine_id"`
	EngineName string            `json:"engine_name"`
	Created    time.Time         `json:"created"`
	Duration   time.Duration     `json:"duration"`
	Audio      string            `json:"audio"` // 16 kHz WAV inside the folder
}

// Store is the folder of transcripts plus index.json. Methods are safe
// from any goroutine.
type Store struct {
	mu      sync.Mutex
	dir     string
	entries []Entry
}

// Open loads the store in dir.
func Open(dir string) *Store {
	s := &Store{dir: dir}
	os.MkdirAll(dir, 0o755)
	if data, err := os.ReadFile(s.index()); err == nil {
		json.Unmarshal(data, &s.entries)
	}
	return s
}

func (s *Store) index() string { return filepath.Join(s.dir, "index.json") }

// NewFile returns a fresh id and the path its audio should be written
// to.
func (s *Store) NewFile() (id, path string) {
	id = fmt.Sprintf("%d", time.Now().UnixNano())
	return id, filepath.Join(s.dir, id+".wav")
}

// AudioPath returns the absolute path of an entry's audio.
func (s *Store) AudioPath(e Entry) string { return filepath.Join(s.dir, e.Audio) }

// Add records an entry, newest first.
func (s *Store) Add(e Entry) {
	s.mu.Lock()
	s.entries = append([]Entry{e}, s.entries...)
	s.mu.Unlock()
	s.save()
}

// SetText replaces an entry's text (after the user edits it).
func (s *Store) SetText(id, text string) {
	s.mu.Lock()
	for i := range s.entries {
		if s.entries[i].ID == id {
			s.entries[i].Text = text
		}
	}
	s.mu.Unlock()
	s.save()
}

// Get returns the entry with id.
func (s *Store) Get(id string) (Entry, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, e := range s.entries {
		if e.ID == id {
			return e, true
		}
	}
	return Entry{}, false
}

// Remove deletes an entry and its audio.
func (s *Store) Remove(id string) {
	s.mu.Lock()
	for i, e := range s.entries {
		if e.ID == id {
			os.Remove(filepath.Join(s.dir, e.Audio))
			s.entries = append(s.entries[:i], s.entries[i+1:]...)
			break
		}
	}
	s.mu.Unlock()
	s.save()
}

// Entries returns a copy of all entries, newest first.
func (s *Store) Entries() []Entry {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := append([]Entry(nil), s.entries...)
	sort.SliceStable(out, func(i, j int) bool { return out[i].Created.After(out[j].Created) })
	return out
}

func (s *Store) save() {
	s.mu.Lock()
	data, _ := json.MarshalIndent(s.entries, "", "  ")
	s.mu.Unlock()
	tmp := s.index() + ".tmp"
	if os.WriteFile(tmp, data, 0o644) == nil {
		os.Rename(tmp, s.index())
	}
}
