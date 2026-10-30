package core

import (
	"bytes"
	"context"
	"errors"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/mehran9675/giode"

	"github.com/mehran9675/sonora/ui/tr"
)

// PickFile shows the system file picker for one file of the given
// types (extensions or MIME types) and returns it with its name. It
// blocks: call it off the UI goroutine. A desktop without a picker
// service gets a hint to use another way in: otherwise, an English
// sentence translated here.
func (c *Ctx) PickFile(types []string, otherwise string) (giode.PickedFile, error) {
	files, err := c.App.PickFiles(context.Background(), giode.PickOptions{Types: types})
	switch {
	case errors.Is(err, giode.ErrCanceled):
		return giode.PickedFile{}, err
	case err != nil:
		// Linux systems without the XDG desktop portal have no picker.
		msg := tr.T("Couldn't open a file picker on this system ({error}).", "error", err)
		if otherwise != "" {
			msg += " " + tr.T(otherwise)
		}
		return giode.PickedFile{}, errors.New(msg)
	case len(files) == 0:
		return giode.PickedFile{}, giode.ErrCanceled
	}
	return files[0], nil
}

// SavePicked copies a picked file into dir and returns its path. A
// file without a name (or an extension) is named prefix plus the time,
// with the extension taken from the content.
func SavePicked(f giode.PickedFile, dir, prefix string) (string, error) {
	name := filepath.Base(f.Name)
	head := make([]byte, 12)
	n, _ := io.ReadFull(f, head)
	head = head[:n]
	if f.Name == "" || filepath.Ext(name) == "" {
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
	_, err = io.Copy(out, f)
	return dst, err
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
