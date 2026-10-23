package sysdlg

import (
	"os"
	"os/exec"
	"runtime"
)

// OpenFile asks for an existing file. Blocks; call it off the UI
// goroutine.
func OpenFile(title string, f Filter) (string, error) {
	switch runtime.GOOS {
	case "darwin":
		script := `POSIX path of (choose file with prompt "` + esc(title) + `")`
		return run(exec.Command("osascript", "-e", script))
	case "windows":
		ps := `Add-Type -AssemblyName System.Windows.Forms;` +
			`$d = New-Object System.Windows.Forms.OpenFileDialog;` +
			`$d.Title = '` + psEsc(title) + `';` +
			`$d.Filter = '` + psEsc(winFilter(f)) + `';` +
			`if ($d.ShowDialog() -eq 'OK') { [Console]::Out.Write($d.FileName) }`
		return run(exec.Command("powershell", "-NoProfile", "-STA", "-Command", ps))
	default:
		if p, err := exec.LookPath("zenity"); err == nil {
			args := []string{"--file-selection", "--title=" + title}
			if len(f.Extensions) > 0 {
				args = append(args, "--file-filter="+f.Name+" | "+globs(f, " "), "--file-filter=All files | *")
			}
			return run(exec.Command(p, args...))
		}
		if p, err := exec.LookPath("kdialog"); err == nil {
			home, _ := os.UserHomeDir()
			return run(exec.Command(p, "--title", title, "--getopenfilename", home, globs(f, " ")+"|"+f.Name))
		}
	}
	return "", ErrUnavailable
}
