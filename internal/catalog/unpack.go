// Package catalog finds every available engine: the built-in ones
// embedded in the binary, plugins dropped into the data folder, and
// API/command engines the user added.
package catalog

import (
	"io/fs"
	"os"
	"path/filepath"

	"github.com/mehran9675/sonora/python"
)

// Unpack writes the embedded Python runtime and built-in engines to
// dir (replacing older copies, so app updates take effect) and returns
// the worker script path.
func Unpack(dir string) (string, error) {
	os.RemoveAll(dir)
	err := fs.WalkDir(python.Files, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		data, err := python.Files.ReadFile(path)
		if err != nil {
			return err
		}
		out := filepath.Join(dir, filepath.FromSlash(path))
		if err := os.MkdirAll(filepath.Dir(out), 0o755); err != nil {
			return err
		}
		return os.WriteFile(out, data, 0o644)
	})
	return filepath.Join(dir, "runtime", "worker.py"), err
}
