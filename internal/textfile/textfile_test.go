package textfile

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// The fixtures were made from sample.md with pandoc (and LibreOffice
// for the PDF).
func TestRead(t *testing.T) {
	for _, name := range []string{"sample.md", "sample.docx", "sample.odt", "sample.epub", "sample.html", "sample.rtf", "sample.pdf", "sample.srt"} {
		data, err := os.ReadFile(filepath.Join("testdata", name))
		if err != nil {
			t.Fatal(err)
		}
		// Named, and unnamed as the Android picker hands files over.
		for _, n := range []string{name, ""} {
			if n == "" && (name == "sample.md" || name == "sample.html") {
				continue // plain text without a name is read as is
			}
			text, err := Read(n, bytes.NewReader(data))
			if err != nil {
				t.Fatalf("%s (%q): %v", name, n, err)
			}
			flat := strings.Join(strings.Fields(text), " ")
			for _, want := range []string{"The keeper climbed the stairs at dusk", "Light the lamp"} {
				if !strings.Contains(flat, want) {
					t.Errorf("%s (%q): missing %q in %q", name, n, want, flat)
				}
			}
			if strings.ContainsAny(flat, "*#{}<>") {
				t.Errorf("%s (%q): markup left in %q", name, n, flat)
			}
		}
	}
}

func TestMarkdown(t *testing.T) {
	got := tidy(markdown("# Title\n\nSome *bold* [link](http://x).\n\n```\nskip me\n```\n- item"))
	want := "Title.\n\nSome bold link.\n\nitem"
	if got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestBinaryRejected(t *testing.T) {
	if _, err := Read("song.bin", bytes.NewReader([]byte{0, 1, 2, 0, 5})); err == nil {
		t.Fatal("binary file accepted")
	}
}
