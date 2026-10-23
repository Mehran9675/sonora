package pyenv

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/gzip"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// Extract unpacks a .zip or .tar.gz archive. dest maps each archive
// entry to its output path; "" skips the entry.
func Extract(data []byte, ext string, dest func(name string) string) error {
	if ext == ".zip" {
		return extractZip(data, dest)
	}
	return extractTarGz(data, dest)
}

// StripTo returns a dest function for Extract that drops the archive's
// top-level folder and places everything under root.
func StripTo(root string) func(string) string {
	return func(name string) string {
		_, rest, ok := strings.Cut(filepath.ToSlash(name), "/")
		if !ok || rest == "" {
			return ""
		}
		return safeJoin(root, rest)
	}
}

func extractZip(data []byte, dest func(string) string) error {
	zr, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return err
	}
	for _, f := range zr.File {
		if f.FileInfo().IsDir() {
			continue
		}
		rc, err := f.Open()
		if err != nil {
			return err
		}
		err = writeEntry(dest(f.Name), f.Mode(), rc)
		rc.Close()
		if err != nil {
			return err
		}
	}
	return nil
}

func extractTarGz(data []byte, dest func(string) string) error {
	gz, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		return err
	}
	tr := tar.NewReader(gz)
	for {
		h, err := tr.Next()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}
		if h.Typeflag != tar.TypeReg {
			continue
		}
		if err := writeEntry(dest(h.Name), os.FileMode(h.Mode), tr); err != nil {
			return err
		}
	}
}

func writeEntry(out string, mode os.FileMode, r io.Reader) error {
	if out == "" {
		return nil
	}
	if err := os.MkdirAll(filepath.Dir(out), 0o755); err != nil {
		return err
	}
	f, err := os.OpenFile(out, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, mode.Perm()|0o644)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = io.Copy(f, r)
	return err
}

// safeJoin joins an archive entry name under root, refusing entries
// that would escape it.
func safeJoin(root, name string) string {
	p := filepath.Join(root, filepath.FromSlash(name))
	if !strings.HasPrefix(p, filepath.Clean(root)+string(os.PathSeparator)) {
		return ""
	}
	return p
}
