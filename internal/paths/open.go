package paths

import (
	"os/exec"
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
