package core

import "github.com/mehran9675/sonora/internal/paths"

// ReferencesDir holds voice recordings chosen for cloning.
func ReferencesDir() string { return paths.Sub("references") }

// PickedDir holds copies of audio files picked on phones, kept until
// they're transcribed.
func PickedDir() string { return paths.Sub("cache", "picked") }

// CacheDir holds temporary files: shared exports, converted audio.
func CacheDir() string { return paths.Sub("cache") }
