package core

import "github.com/mehran9675/sonora/internal/paths"

// ReferencesDir holds voice recordings chosen for cloning.
func ReferencesDir() string { return paths.Sub("references") }
