// Package pyenv manages the Python environments engines run in. It
// uses uv (https://github.com/astral-sh/uv), downloading it on first
// use, which in turn downloads a matching Python - so users never
// install Python, pip or anything else by hand.
package pyenv

import (
	"os"
	"path/filepath"
	"runtime"

	"github.com/mehran9675/sonora/internal/paths"
)

// Logger receives progress lines.
type Logger func(line string)

func exe(name string) string {
	if runtime.GOOS == "windows" {
		return name + ".exe"
	}
	return name
}

// NoTorchBackend, as a key of Env's extra map, leaves PyTorch to be
// installed from PyPI instead of PyTorch's own index.
const NoTorchBackend = "SONORA_NO_TORCH_BACKEND"

// Env returns the environment for uv and Python child processes.
func Env(extra map[string]string) []string {
	env := append(os.Environ(),
		"UV_PYTHON_INSTALL_DIR="+filepath.Join(paths.Data(), "python"),
		"PYTHONIOENCODING=utf-8",
		"PYTHONUNBUFFERED=1",
	)
	if _, off := extra[NoTorchBackend]; !off {
		// Pick the PyTorch build matching the machine's GPU (or CPU).
		env = append(env, "UV_TORCH_BACKEND=auto")
	}
	for k, v := range extra {
		if k != NoTorchBackend {
			env = append(env, k+"="+v)
		}
	}
	return env
}
