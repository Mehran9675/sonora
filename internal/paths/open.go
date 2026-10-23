package paths

import (
	"os/exec"
	"path/filepath"
	"runtime"
)

// Open opens a file or folder with the default application.
func Open(path string) error {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("explorer", path)
	case "darwin":
		cmd = exec.Command("open", path)
	default:
		cmd = exec.Command("xdg-open", path)
	}
	HideWindow(cmd)
	return cmd.Start()
}

// Reveal shows a file selected in the file manager where supported,
// otherwise opens its folder.
func Reveal(path string) error {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("explorer", "/select,", path)
	case "darwin":
		cmd = exec.Command("open", "-R", path)
	default:
		return Open(filepath.Dir(path))
	}
	HideWindow(cmd)
	return cmd.Start()
}
