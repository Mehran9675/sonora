package core

import (
	"context"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/mehran9675/giode"

	"github.com/mehran9675/sonora/ui/icons"
)

// SaveClip asks where to save the WAV at path (the system picker on
// Android, a native dialog on desktops) and copies it there. A desktop
// without a dialog service (a Linux system without the XDG portal)
// gets the file in its Downloads folder instead.
func (c *Ctx) SaveClip(path, title string) {
	var fallback string
	c.Background(func() error {
		w, err := c.App.SaveFile(context.Background(), clipName(title))
		if err != nil && !errors.Is(err, giode.ErrCanceled) && !Mobile {
			fallback = filepath.Join(downloadsDir(), clipName(title))
			w, err = os.Create(fallback)
		}
		if err != nil {
			return err
		}
		in, err := os.Open(path)
		if err != nil {
			w.Close()
			return err
		}
		defer in.Close()
		if _, err := io.Copy(w, in); err != nil {
			w.Close()
			return err
		}
		return w.Close()
	}, func(err error) {
		switch {
		case err != nil:
			c.Error(err)
		case fallback != "":
			c.Success("Saved to " + fallback)
		default:
			c.Success("Saved")
		}
	})
}

// downloadsDir is the user's Downloads folder, or home.
func downloadsDir() string {
	home, _ := os.UserHomeDir()
	if st, err := os.Stat(filepath.Join(home, "Downloads")); err == nil && st.IsDir() {
		return filepath.Join(home, "Downloads")
	}
	return home
}

// ShareClip hands the WAV at path to another app through the share
// sheet on phones, or shows it in the file manager on desktops.
func (c *Ctx) ShareClip(path string) {
	c.Background(func() error {
		return c.App.Share(context.Background(), giode.ShareFile{Path: path, MIME: "audio/wav"})
	}, nil)
}

// ShareIcon is the icon for ShareClip: a share arrow where it opens
// the share sheet, a folder where it opens the file manager.
func ShareIcon() string {
	if Mobile {
		return icons.Share
	}
	return icons.Folder
}

// clipName suggests a file name for an exported clip.
func clipName(title string) string {
	clean := strings.Map(func(r rune) rune {
		if strings.ContainsRune(`<>:"/\|?*`, r) {
			return '-'
		}
		return r
	}, strings.TrimSpace(title))
	if clean == "" {
		clean = "speech"
	}
	return "Sonora - " + clean + ".wav"
}
