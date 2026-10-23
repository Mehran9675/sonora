//go:build windows

package system

import (
	"context"
	"fmt"
	"math"
	"os/exec"
	"strings"

	"github.com/mehran9675/sonora/internal/engines"
	"github.com/mehran9675/sonora/internal/paths"
)

// sapi drives Windows' System.Speech synthesizer through PowerShell.
type sapi struct{}

func platform() (backend, string, string, string) {
	return sapi{}, "Windows SAPI", "", "Microsoft Zira Desktop"
}

func (sapi) available() bool { return true }

func powershell(ctx context.Context, script string) ([]byte, error) {
	cmd := exec.CommandContext(ctx, "powershell", "-NoProfile", "-NonInteractive", "-Command", script)
	paths.HideWindow(cmd)
	return cmd.CombinedOutput()
}

func (sapi) voices(ctx context.Context) ([]engines.Voice, error) {
	out, err := powershell(ctx, `Add-Type -AssemblyName System.Speech;`+
		`$s = New-Object System.Speech.Synthesis.SpeechSynthesizer;`+
		`$s.GetInstalledVoices() | ForEach-Object { $v = $_.VoiceInfo; "$($v.Name)|$($v.Culture.Name)|$($v.Gender)" }`)
	if err != nil {
		return nil, fmt.Errorf("list voices: %v %s", err, out)
	}
	var vs []engines.Voice
	for _, line := range strings.Split(string(out), "\n") {
		f := strings.Split(strings.TrimSpace(line), "|")
		if len(f) == 3 && f[0] != "" {
			vs = append(vs, engines.Voice{ID: f[0], Name: strings.TrimPrefix(f[0], "Microsoft "), Group: f[1], Description: f[2]})
		}
	}
	return vs, nil
}

func (sapi) speak(ctx context.Context, textFile, voice string, speed float64, out string) error {
	rate := int(math.Round(math.Log2(speed) * 10))
	rate = max(-10, min(10, rate))
	q := func(s string) string { return "'" + strings.ReplaceAll(s, "'", "''") + "'" }
	script := `Add-Type -AssemblyName System.Speech;` +
		`$s = New-Object System.Speech.Synthesis.SpeechSynthesizer;`
	if voice != "" {
		script += `$s.SelectVoice(` + q(voice) + `);`
	}
	script += fmt.Sprintf(`$s.Rate = %d;`, rate) +
		`$s.SetOutputToWaveFile(` + q(out) + `);` +
		`$s.Speak([IO.File]::ReadAllText(` + q(textFile) + `, [Text.Encoding]::UTF8));` +
		`$s.Dispose()`
	if msg, err := powershell(ctx, script); err != nil {
		return fmt.Errorf("speech: %v %s", err, msg)
	}
	return nil
}
