package sysdlg

import (
	"os/exec"
	"path/filepath"
	"runtime"

	"github.com/mehran9675/sonora/internal/paths"
)

// SaveFile asks where to save a file, suggesting name.
func SaveFile(title, name string, f Filter) (string, error) {
	switch runtime.GOOS {
	case "darwin":
		script := `POSIX path of (choose file name with prompt "` + esc(title) + `" default name "` + esc(name) + `")`
		return run(exec.Command("osascript", "-e", script))
	case "windows":
		ps := `Add-Type -AssemblyName System.Windows.Forms;` +
			`$d = New-Object System.Windows.Forms.SaveFileDialog;` +
			`$d.Title = '` + psEsc(title) + `';` +
			`$d.FileName = '` + psEsc(name) + `';` +
			`$d.Filter = '` + psEsc(winFilter(f)) + `';` +
			`if ($d.ShowDialog() -eq 'OK') { [Console]::Out.Write($d.FileName) }`
		return run(exec.Command("powershell", "-NoProfile", "-STA", "-Command", ps))
	default:
		start := filepath.Join(paths.Downloads(), name)
		if p, err := exec.LookPath("zenity"); err == nil {
			return run(exec.Command(p, "--file-selection", "--save", "--confirm-overwrite", "--title="+title, "--filename="+start))
		}
		if p, err := exec.LookPath("kdialog"); err == nil {
			return run(exec.Command(p, "--title", title, "--getsavefilename", start, globs(f, " ")+"|"+f.Name))
		}
	}
	return "", ErrUnavailable
}
