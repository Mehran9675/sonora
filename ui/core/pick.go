package core

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"time"
)

// SavePicked copies a file from the system picker into dir and returns
// its path. The Android picker hands over a stream without a file
// name, so the name is then prefix plus the time, with the extension
// taken from the content.
func SavePicked(r io.Reader, dir, prefix string) (string, error) {
	name := PickedName(r)
	head := make([]byte, 12)
	n, _ := io.ReadFull(r, head)
	head = head[:n]
	if name == "" || filepath.Ext(name) == "" {
		name = prefix + "-" + time.Now().Format("20060102-150405") + sniffExt(head)
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}
	dst := filepath.Join(dir, name)
	out, err := os.Create(dst)
	if err != nil {
		return "", err
	}
	defer out.Close()
	if _, err := out.Write(head); err != nil {
		return "", err
	}
	_, err = io.Copy(out, r)
	return dst, err
}

// PickedName returns the base name of a picked file, or "" when the
// platform doesn't tell.
func PickedName(r io.Reader) string {
	if n, ok := r.(interface{ Name() string }); ok {
		return filepath.Base(n.Name())
	}
	return ""
}

// sniffExt guesses a file extension from the first bytes of a file.
func sniffExt(head []byte) string {
	switch {
	case bytes.HasPrefix(head, []byte("RIFF")):
		return ".wav"
	case bytes.HasPrefix(head, []byte("fLaC")):
		return ".flac"
	case bytes.HasPrefix(head, []byte("OggS")):
		return ".ogg"
	case len(head) >= 8 && string(head[4:8]) == "ftyp":
		return ".m4a"
	case bytes.HasPrefix(head, []byte("ID3")), len(head) >= 2 && head[0] == 0xFF && head[1]&0xE0 == 0xE0:
		return ".mp3"
	case bytes.HasPrefix(head, []byte("%PDF")):
		return ".pdf"
	}
	return ""
}
