package pyenv

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
)

// Download fetches url into memory, logging progress.
func Download(ctx context.Context, url string, log Logger) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%s: %s", url, resp.Status)
	}
	var buf bytes.Buffer
	chunk := make([]byte, 256*1024)
	lastStep := -1
	for {
		n, err := resp.Body.Read(chunk)
		buf.Write(chunk[:n])
		if resp.ContentLength > 0 {
			pct := int(int64(buf.Len()) * 100 / resp.ContentLength)
			if pct/5 != lastStep {
				lastStep = pct / 5
				log(fmt.Sprintf("Downloading… %d%%", pct))
			}
		}
		if err == io.EOF {
			return buf.Bytes(), nil
		}
		if err != nil {
			return nil, err
		}
	}
}
