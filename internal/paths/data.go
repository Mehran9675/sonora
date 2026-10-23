// Package paths locates Sonora's data folder and opens files, folders
// and links with the default application.
package paths

import (
	"os"
	"path/filepath"
	"runtime"
)

// Data returns the folder holding engines, models and generated audio,
// creating it if needed. SONORA_HOME overrides it.
func Data() string {
	dir := os.Getenv("SONORA_HOME")
	if dir == "" {
		home, _ := os.UserHomeDir()
		switch runtime.GOOS {
		case "windows":
			base := os.Getenv("LOCALAPPDATA")
			if base == "" {
				base = filepath.Join(home, "AppData", "Local")
			}
			dir = filepath.Join(base, "Sonora")
		case "android", "ios":
			dir = filepath.Join(mobileDataDir(), "sonora")
		case "darwin":
			dir = filepath.Join(home, "Library", "Application Support", "Sonora")
		default:
			base := os.Getenv("XDG_DATA_HOME")
			if base == "" {
				base = filepath.Join(home, ".local", "share")
			}
			dir = filepath.Join(base, "sonora")
		}
	}
	os.MkdirAll(dir, 0o755)
	return dir
}

// Sub returns a folder inside Data, creating it if needed.
func Sub(parts ...string) string {
	dir := filepath.Join(append([]string{Data()}, parts...)...)
	os.MkdirAll(dir, 0o755)
	return dir
}
