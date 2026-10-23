package pyenv

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"

	"github.com/mehran9675/sonora/internal/paths"
)

var uvMu sync.Mutex

// UVPath returns the uv binary to use, or "" if none is available yet.
// A copy managed by Sonora wins over one on PATH so versions stay
// predictable.
func UVPath() string {
	own := filepath.Join(paths.Data(), "bin", exe("uv"))
	if _, err := os.Stat(own); err == nil {
		return own
	}
	if p, err := exec.LookPath("uv"); err == nil {
		return p
	}
	return ""
}

// UVVersion returns `uv --version`, or "" when uv is missing.
func UVVersion() string {
	p := UVPath()
	if p == "" {
		return ""
	}
	cmd := exec.Command(p, "--version")
	paths.HideWindow(cmd)
	out, err := cmd.Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}

// EnsureUV returns a usable uv binary, downloading it if needed.
func EnsureUV(ctx context.Context, log Logger) (string, error) {
	uvMu.Lock()
	defer uvMu.Unlock()
	if p := UVPath(); p != "" {
		return p, nil
	}
	target, err := uvTarget()
	if err != nil {
		return "", err
	}
	ext := ".tar.gz"
	if runtime.GOOS == "windows" {
		ext = ".zip"
	}
	log("Downloading the Python toolchain (uv)…")
	data, err := Download(ctx, "https://github.com/astral-sh/uv/releases/latest/download/uv-"+target+ext, log)
	if err != nil {
		return "", fmt.Errorf("download uv: %w", err)
	}
	bin := filepath.Join(paths.Data(), "bin")
	want := map[string]bool{exe("uv"): true, exe("uvx"): true}
	err = Extract(data, ext, func(name string) string {
		if base := filepath.Base(name); want[base] {
			return filepath.Join(bin, base)
		}
		return ""
	})
	if err != nil {
		return "", fmt.Errorf("unpack uv: %w", err)
	}
	p := filepath.Join(bin, exe("uv"))
	if _, err := os.Stat(p); err != nil {
		return "", fmt.Errorf("uv archive did not contain %s", exe("uv"))
	}
	log("Python toolchain ready.")
	return p, nil
}

func uvTarget() (string, error) {
	arch := map[string]string{"amd64": "x86_64", "arm64": "aarch64"}[runtime.GOARCH]
	if arch == "" {
		return "", fmt.Errorf("unsupported CPU architecture %s", runtime.GOARCH)
	}
	switch runtime.GOOS {
	case "linux":
		return arch + "-unknown-linux-gnu", nil
	case "darwin":
		return arch + "-apple-darwin", nil
	case "windows":
		return arch + "-pc-windows-msvc", nil
	}
	return "", fmt.Errorf("unsupported OS %s", runtime.GOOS)
}
