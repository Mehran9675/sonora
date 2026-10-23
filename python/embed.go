// Package python embeds the worker runtime and the built-in engine
// definitions, which the app unpacks into its data folder at startup.
package python

import "embed"

// Files holds runtime/ (the worker and its helper package) and
// engines/ (one folder per built-in engine: engine.json + adapter.py).
//
//go:embed all:runtime engines
var Files embed.FS
