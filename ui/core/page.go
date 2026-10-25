package core

// Page identifies a top-level screen.
type Page int

const (
	PageStudio Page = iota
	PageTranscribe
	PageLibrary
	PageEngines
	PageSettings
)
