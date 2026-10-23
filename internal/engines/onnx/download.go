package onnx

import (
	"archive/tar"
	"compress/bzip2"
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/mehran9675/sonora/internal/engines"
)

// fetch downloads and unpacks a model package into dir/<id>, unless it
// is already there. Unpacking goes to a temporary folder that is
// renamed when complete, so an interrupted download never looks
// installed.
func fetch(ctx context.Context, dir string, m model, log engines.Logger) (string, error) {
	dst := filepath.Join(dir, m.ID)
	if _, err := os.Stat(dst); err == nil {
		return dst, nil
	}
	tmp := dst + ".partial"
	os.RemoveAll(tmp)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, m.url(), nil)
	if err != nil {
		return "", err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("download %s: %s", m.ID, resp.Status)
	}
	body := &progressReader{r: resp.Body, total: resp.ContentLength, label: "Downloading " + m.Label, log: log}
	// Download and unpack in one pass; the bzip2 stream is decoded as it
	// arrives.
	tr := tar.NewReader(bzip2.NewReader(body))
	for {
		h, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			os.RemoveAll(tmp)
			if ctx.Err() != nil {
				return "", ctx.Err()
			}
			return "", fmt.Errorf("unpack %s: %w", m.ID, err)
		}
		// Archives hold <id>/...; drop that top folder.
		_, rest, ok := strings.Cut(filepath.ToSlash(h.Name), "/")
		if !ok || rest == "" {
			continue
		}
		out := filepath.Join(tmp, filepath.FromSlash(rest))
		if !strings.HasPrefix(out, tmp+string(os.PathSeparator)) {
			continue // outside the folder
		}
		switch h.Typeflag {
		case tar.TypeDir:
			err = os.MkdirAll(out, 0o755)
		case tar.TypeReg:
			err = writeFile(out, tr)
		}
		if err != nil {
			os.RemoveAll(tmp)
			return "", err
		}
	}
	if err := os.Rename(tmp, dst); err != nil {
		os.RemoveAll(tmp)
		return "", err
	}
	log(m.Label + " is ready.")
	return dst, nil
}

func writeFile(path string, r io.Reader) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	if _, err := io.Copy(f, r); err != nil {
		f.Close()
		return err
	}
	return f.Close()
}

// progressReader logs "label NN%" as bytes arrive, in 2% steps.
type progressReader struct {
	r     io.Reader
	total int64
	read  int64
	step  int64
	label string
	log   engines.Logger
}

func (p *progressReader) Read(b []byte) (int, error) {
	n, err := p.r.Read(b)
	p.read += int64(n)
	if p.total > 0 {
		if s := p.read * 50 / p.total; s != p.step {
			p.step = s
			p.log(fmt.Sprintf("%s… %d%%", p.label, p.read*100/p.total))
		}
	}
	return n, err
}
