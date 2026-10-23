package sysdlg

import (
	"errors"
	"os/exec"
	"strings"

	"github.com/mehran9675/sonora/internal/paths"
)

func run(cmd *exec.Cmd) (string, error) {
	paths.HideWindow(cmd)
	out, err := cmd.Output()
	path := strings.TrimSpace(string(out))
	if path == "" {
		if err != nil {
			var ee *exec.ExitError
			if errors.As(err, &ee) {
				return "", ErrCancelled
			}
			return "", err
		}
		return "", ErrCancelled
	}
	return path, nil
}

func globs(f Filter, sep string) string {
	if len(f.Extensions) == 0 {
		return "*"
	}
	var g []string
	for _, e := range f.Extensions {
		g = append(g, "*."+e)
	}
	return strings.Join(g, sep)
}

func winFilter(f Filter) string {
	if len(f.Extensions) == 0 {
		return "All files|*.*"
	}
	return f.Name + "|" + globs(f, ";") + "|All files|*.*"
}

func esc(s string) string   { return strings.ReplaceAll(strings.ReplaceAll(s, `\`, `\\`), `"`, `\"`) }
func psEsc(s string) string { return strings.ReplaceAll(s, `'`, `''`) }
