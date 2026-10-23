//go:build darwin

package system

import (
	"context"
	"fmt"
	"os/exec"
	"regexp"
	"strings"

	"github.com/mehran9675/sonora/internal/engines"
)

// say drives macOS's speech synthesizer.
type say struct{}

func platform() (backend, string, string, string) { return say{}, "macOS", "", "Samantha" }

func (say) available() bool { return true }

// voiceLine matches `say -v ?` output such as
// "Eddy (English (UK))   en_GB    # Hello! My name is Eddy.".
var voiceLine = regexp.MustCompile(`^(.+?)\s+([a-z]{2,3}[_-][A-Za-z0-9]+)\s+#\s*(.*)$`)

func (say) voices(ctx context.Context) ([]engines.Voice, error) {
	out, err := exec.CommandContext(ctx, "say", "-v", "?").Output()
	if err != nil {
		return nil, err
	}
	var vs []engines.Voice
	for _, line := range strings.Split(string(out), "\n") {
		m := voiceLine.FindStringSubmatch(strings.TrimSpace(line))
		if m == nil {
			continue
		}
		vs = append(vs, engines.Voice{ID: m[1], Name: m[1], Group: m[2], Description: m[3]})
	}
	return vs, nil
}

func (say) speak(ctx context.Context, textFile, voice string, speed float64, out string) error {
	args := []string{"-o", out, "--file-format=WAVE", "--data-format=LEI16@22050", "-r", fmt.Sprint(int(175 * speed)), "-f", textFile}
	if voice != "" {
		args = append(args, "-v", voice)
	}
	if msg, err := exec.CommandContext(ctx, "say", args...).CombinedOutput(); err != nil {
		return fmt.Errorf("say: %v %s", err, msg)
	}
	return nil
}
