// Package library keeps every generated clip with the settings that
// produced it.
package library

import "time"

// Entry is one generated clip.
type Entry struct {
	ID         string        `json:"id"`
	Text       string        `json:"text"`
	EngineID   string        `json:"engine_id"`
	EngineName string        `json:"engine_name"`
	Voice      string        `json:"voice"`
	VoiceName  string        `json:"voice_name"`
	Created    time.Time     `json:"created"`
	Duration   time.Duration `json:"duration"`
	File       string        `json:"file"` // name inside the library folder
}
