// Package convert turns audio in other formats (MP3, OGG, AIFF…) into
// WAV using ffmpeg when it is installed.
package convert

import (
	"bytes"
	"errors"
	"os/exec"

	"github.com/mehran9675/sonora/internal/paths"
)

// ErrNoFFmpeg means the audio isn't WAV and ffmpeg isn't available to
// convert it.
var ErrNoFFmpeg = errors.New("the engine returned compressed audio; install ffmpeg or ask it for WAV")

// Available reports whether ffmpeg is on PATH.
func Available() bool {
	_, err := exec.LookPath("ffmpeg")
	return err == nil
}

// ToWAV converts encoded audio bytes to WAV bytes.
func ToWAV(data []byte) ([]byte, error) {
	if !Available() {
		return nil, ErrNoFFmpeg
	}
	cmd := exec.Command("ffmpeg", "-hide_banner", "-loglevel", "error", "-i", "pipe:0", "-f", "wav", "-acodec", "pcm_s16le", "pipe:1")
	paths.HideWindow(cmd)
	cmd.Stdin = bytes.NewReader(data)
	var out, errb bytes.Buffer
	cmd.Stdout, cmd.Stderr = &out, &errb
	if err := cmd.Run(); err != nil {
		return nil, errors.New("ffmpeg: " + errb.String())
	}
	return out.Bytes(), nil
}

// FileToWAV converts the file at in to a WAV file at out.
func FileToWAV(in, out string) error {
	if !Available() {
		return ErrNoFFmpeg
	}
	cmd := exec.Command("ffmpeg", "-hide_banner", "-loglevel", "error", "-y", "-i", in, "-acodec", "pcm_s16le", out)
	paths.HideWindow(cmd)
	if b, err := cmd.CombinedOutput(); err != nil {
		return errors.New("ffmpeg: " + string(b))
	}
	return nil
}
