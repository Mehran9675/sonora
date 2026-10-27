package python

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/mehran9675/sonora/internal/engines"
	"github.com/mehran9675/sonora/internal/pyenv"
)

// marker records a finished install; its absence means "not installed".
type marker struct {
	Installed time.Time  `json:"installed"`
	Recipe    [][]string `json:"recipe"`
}

func (e *Engine) markerPath() string { return filepath.Join(e.dir, "installed.json") }
func (e *Engine) venv() string       { return filepath.Join(e.dir, "venv") }
func (e *Engine) src() string        { return filepath.Join(e.dir, "src") }

// Installed reports whether Install has completed.
func (e *Engine) Installed() bool {
	_, err := os.Stat(e.markerPath())
	return err == nil
}

// Install creates the environment, fetches sources and installs the
// packages listed in the manifest.
func (e *Engine) Install(ctx context.Context, log engines.Logger) error {
	e.Close()
	os.Remove(e.markerPath())
	uv, err := pyenv.EnsureUV(ctx, pyenv.Logger(log))
	if err != nil {
		return err
	}
	log("Creating the Python environment…")
	if err := pyenv.CreateVenv(ctx, uv, e.venv(), e.m.Python, e.m.Env, pyenv.Logger(log)); err != nil {
		return err
	}
	if e.m.Source != nil {
		log("Downloading " + e.m.Name + " sources…")
		data, err := pyenv.Download(ctx, e.m.Source.URL, pyenv.Logger(log))
		if err != nil {
			return err
		}
		os.RemoveAll(e.src())
		ext := ".zip"
		if strings.HasSuffix(e.m.Source.URL, ".tar.gz") {
			ext = ".tar.gz"
		}
		if err := pyenv.Extract(data, ext, pyenv.StripTo(e.src())); err != nil {
			return fmt.Errorf("unpack sources: %w", err)
		}
	}
	for i, step := range e.m.Install {
		args := make([]string, len(step))
		for j, a := range step {
			args[j] = strings.ReplaceAll(a, "{src}", e.src())
		}
		log(fmt.Sprintf("Installing packages (step %d of %d)…", i+1, len(e.m.Install)))
		err := pyenv.PipInstall(ctx, uv, e.venv(), args, e.m.Env, pyenv.Logger(log))
		if err != nil && strings.Contains(err.Error(), "download.pytorch.org") {
			// PyTorch's own index (used to match the GPU) is unreachable
			// from some networks; PyPI carries a build that works too.
			log("PyTorch's download server is unreachable; retrying with PyPI…")
			env := map[string]string{pyenv.NoTorchBackend: ""}
			for k, v := range e.m.Env {
				env[k] = v
			}
			err = pyenv.PipInstall(ctx, uv, e.venv(), args, env, pyenv.Logger(log))
		}
		if err != nil {
			return err
		}
	}
	data, _ := json.MarshalIndent(marker{Installed: time.Now(), Recipe: e.m.Install}, "", "  ")
	if err := os.WriteFile(e.markerPath(), data, 0o644); err != nil {
		return err
	}
	log(e.m.Name + " is installed.")
	return nil
}

// Uninstall deletes the engine's environment, sources and models.
func (e *Engine) Uninstall() error {
	e.Close()
	e.modelsMu.Lock()
	e.models = nil
	e.modelsMu.Unlock()
	return os.RemoveAll(e.dir)
}
