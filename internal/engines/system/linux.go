//go:build !windows && !darwin

package system

import (
	"context"
	"fmt"
	"os/exec"
	"strings"

	"github.com/mehran9675/sonora/internal/engines"
)

// espeak drives eSpeak NG (or classic eSpeak).
type espeak struct{}

func platform() (backend, string, string, string) {
	return espeak{}, "eSpeak NG", "Needs eSpeak NG: install the espeak-ng package from your distribution.", "en-us"
}

func (espeak) bin() string {
	for _, name := range []string{"espeak-ng", "espeak"} {
		if p, err := exec.LookPath(name); err == nil {
			return p
		}
	}
	return ""
}

func (b espeak) available() bool { return b.bin() != "" }

// voices parses `espeak-ng --voices`:
//
//	Pty Language       Age/Gender VoiceName          File                 Other Languages
//	 5  af              --/M      Afrikaans          gmw/af
func (b espeak) voices(ctx context.Context) ([]engines.Voice, error) {
	out, err := exec.CommandContext(ctx, b.bin(), "--voices").Output()
	if err != nil {
		return nil, err
	}
	var vs []engines.Voice
	for i, line := range strings.Split(string(out), "\n") {
		f := strings.Fields(line)
		if i == 0 || len(f) < 4 {
			continue
		}
		gender := map[string]string{"M": "Male", "F": "Female"}[strings.TrimPrefix(f[2], "--/")]
		vs = append(vs, engines.Voice{
			ID:          f[1],
			Name:        strings.ReplaceAll(f[3], "_", " "),
			Group:       strings.ToUpper(strings.SplitN(f[1], "-", 2)[0]),
			Description: gender,
		})
	}
	return vs, nil
}

func (b espeak) speak(ctx context.Context, textFile, voice string, speed float64, out string) error {
	args := []string{"-w", out, "-s", fmt.Sprint(int(175 * speed)), "-f", textFile}
	if voice != "" {
		args = append(args, "-v", voice)
	}
	if msg, err := exec.CommandContext(ctx, b.bin(), args...).CombinedOutput(); err != nil {
		return fmt.Errorf("espeak: %v %s", err, msg)
	}
	return nil
}
