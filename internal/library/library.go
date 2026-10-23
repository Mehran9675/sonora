package library

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"time"
)

// Library is the folder of clips plus index.json. Methods are safe
// from any goroutine.
type Library struct {
	mu      sync.Mutex
	dir     string
	entries []Entry
}

// Open loads the library in dir.
func Open(dir string) *Library {
	l := &Library{dir: dir}
	if data, err := os.ReadFile(l.index()); err == nil {
		json.Unmarshal(data, &l.entries)
	}
	return l
}

func (l *Library) index() string { return filepath.Join(l.dir, "index.json") }

// Dir returns the library folder.
func (l *Library) Dir() string { return l.dir }

// NewFile returns a fresh id and the path its WAV should be written to.
func (l *Library) NewFile() (id, path string) {
	id = fmt.Sprintf("%d", time.Now().UnixNano())
	return id, filepath.Join(l.dir, id+".wav")
}

// Path returns the absolute path of an entry's audio.
func (l *Library) Path(e Entry) string { return filepath.Join(l.dir, e.File) }

// Add records an entry, newest first.
func (l *Library) Add(e Entry) {
	l.mu.Lock()
	l.entries = append([]Entry{e}, l.entries...)
	l.mu.Unlock()
	l.save()
}

// Remove deletes an entry and its audio.
func (l *Library) Remove(id string) {
	l.mu.Lock()
	for i, e := range l.entries {
		if e.ID == id {
			os.Remove(filepath.Join(l.dir, e.File))
			l.entries = append(l.entries[:i], l.entries[i+1:]...)
			break
		}
	}
	l.mu.Unlock()
	l.save()
}

// Entries returns a copy of all entries, newest first.
func (l *Library) Entries() []Entry {
	l.mu.Lock()
	defer l.mu.Unlock()
	out := append([]Entry(nil), l.entries...)
	sort.SliceStable(out, func(i, j int) bool { return out[i].Created.After(out[j].Created) })
	return out
}

func (l *Library) save() {
	l.mu.Lock()
	data, _ := json.MarshalIndent(l.entries, "", "  ")
	l.mu.Unlock()
	os.WriteFile(l.index(), data, 0o644)
}
