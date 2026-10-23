package paths

import (
	"os/exec"
	"syscall"
)

// HideWindow keeps console programs started by the GUI from flashing a
// console window.
func HideWindow(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true, CreationFlags: 0x08000000}
}
