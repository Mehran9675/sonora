package command

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/mehran9675/sonora/internal/audio/convert"
	"github.com/mehran9675/sonora/internal/engines"
	"github.com/mehran9675/sonora/internal/paths"
	"github.com/mehran9675/sonora/internal/pyenv"
)

// Engine runs a configured command.
type Engine struct {
	m      *engines.Manifest
	config func() engines.Values
}

// New returns the engine; config returns its current option values.
func New(m *engines.Manifest, config func() engines.Values) *Engine {
	return &Engine{m: m, config: config}
}

func (e *Engine) Manifest() *engines.Manifest                   { return e.m }
func (e *Engine) Installed() bool                               { return true }
func (e *Engine) Install(context.Context, engines.Logger) error { return nil }
func (e *Engine) Uninstall() error                              { return nil }
func (e *Engine) Close()                                        {}
func (e *Engine) Loaded() bool                                  { return false }

// Voices returns the configured voice list.
func (e *Engine) Voices(context.Context, engines.Values, engines.Logger) ([]engines.Voice, error) {
	var out []engines.Voice
	for _, v := range strings.Split(e.config().String("voices"), ",") {
		if v = strings.TrimSpace(v); v != "" {
			out = append(out, engines.Voice{ID: v, Name: filepath.Base(v)})
		}
	}
	return out, nil
}

// Synthesize runs the command and collects the audio from {output} or
// standard output.
func (e *Engine) Synthesize(ctx context.Context, r engines.Request, log engines.Logger) error {
	tmpl := e.config().String("command")
	args, err := splitArgs(tmpl)
	if err != nil {
		return err
	}
	textFile, err := os.CreateTemp("", "sonora-*.txt")
	if err != nil {
		return err
	}
	defer os.Remove(textFile.Name())
	textFile.WriteString(r.Text)
	textFile.Close()

	speed := r.Speed
	if speed <= 0 {
		speed = 1
	}
	raw := strings.TrimSuffix(r.Output, ".wav") + ".raw"
	defer os.Remove(raw)
	args = expand(args, map[string]string{
		"text":         r.Text,
		"text_file":    textFile.Name(),
		"output":       raw,
		"voice":        r.Voice,
		"speed":        fmt.Sprintf("%.2f", speed),
		"length_scale": fmt.Sprintf("%.2f", 1/speed),
		"wpm":          fmt.Sprintf("%d", int(175*speed)),
	})
	log("$ " + strings.Join(args, " "))
	cmd := exec.CommandContext(ctx, args[0], args[1:]...)
	paths.HideWindow(cmd)
	var stdout, stderr bytes.Buffer
	cmd.Stdout, cmd.Stderr = &stdout, &stderr
	err = cmd.Run()
	pyenv.ScanLines(&stderr, func(l string) { log(l) })
	if err != nil {
		return fmt.Errorf("%s: %v", args[0], err)
	}
	if strings.Contains(tmpl, "{output}") {
		data, err := os.ReadFile(raw)
		if err != nil {
			return fmt.Errorf("the command did not write {output}: %w", err)
		}
		return convert.SaveAsWAV(r.Output, data)
	}
	if stdout.Len() == 0 {
		return fmt.Errorf("the command produced no audio")
	}
	return convert.SaveAsWAV(r.Output, stdout.Bytes())
}
