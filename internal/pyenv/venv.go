package pyenv

import (
	"context"
	"path/filepath"
	"runtime"
)

// VenvPython returns the interpreter inside a virtual environment.
func VenvPython(venv string) string {
	if runtime.GOOS == "windows" {
		return filepath.Join(venv, "Scripts", "python.exe")
	}
	return filepath.Join(venv, "bin", "python")
}

// CreateVenv creates a virtual environment with the given Python
// version ("3.11"); uv fetches that Python if it isn't installed.
func CreateVenv(ctx context.Context, uv, venv, python string, env map[string]string, log Logger) error {
	if python == "" {
		python = "3.11"
	}
	return Run(ctx, log, env, "", uv, "venv", "--python", python, "--seed", "--allow-existing", venv)
}

// PipInstall installs packages (and pip arguments such as -r file or
// --no-deps) into the environment.
func PipInstall(ctx context.Context, uv, venv string, args []string, env map[string]string, log Logger) error {
	full := append([]string{"pip", "install", "--python", VenvPython(venv)}, args...)
	return Run(ctx, log, env, "", uv, full...)
}
