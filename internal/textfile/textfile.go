// Package textfile reads documents as plain text to be spoken: text
// and Markdown, subtitles (SRT, WebVTT), HTML, Word (DOCX),
// OpenDocument (ODT), EPUB, PDF and RTF.
package textfile

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"path/filepath"
	"regexp"
	"strings"
	"unicode/utf16"
	"unicode/utf8"
)

// Extensions are the files worth offering in a file picker.
var Extensions = []string{"txt", "md", "markdown", "srt", "vtt", "html", "htm", "xhtml", "docx", "odt", "epub", "pdf", "rtf"}

// maxSize bounds how much of a file is read.
const maxSize = 64 << 20

// Read returns the text of the document in r. name (which may be
// empty, as on Android) picks the format by extension; without one the
// format is sniffed from the content.
func Read(name string, r io.Reader) (string, error) {
	data, err := io.ReadAll(io.LimitReader(r, maxSize))
	if err != nil {
		return "", err
	}
	ext := strings.TrimPrefix(strings.ToLower(filepath.Ext(name)), ".")
	if ext == "" {
		ext = sniff(data)
	}
	var text string
	switch ext {
	case "docx":
		text, err = docx(data)
	case "odt":
		text, err = odt(data)
	case "epub":
		text, err = epub(data)
	case "pdf":
		text, err = pdf(data)
	case "rtf":
		text = rtf(decodeText(data))
	case "html", "htm", "xhtml":
		text = html(data)
	case "srt", "vtt":
		text = subtitles(decodeText(data))
	case "md", "markdown":
		text = markdown(decodeText(data))
	default:
		if !looksLikeText(data) {
			return "", fmt.Errorf("%s doesn't look like a text document", orDefault(name, "the file"))
		}
		text = decodeText(data)
	}
	if err != nil {
		return "", err
	}
	text = tidy(text)
	if text == "" {
		return "", errors.New("no text found in " + orDefault(name, "the file"))
	}
	return text, nil
}

// sniff guesses the format of an unnamed file.
func sniff(data []byte) string {
	switch {
	case bytes.HasPrefix(data, []byte("%PDF")):
		return "pdf"
	case bytes.HasPrefix(data, []byte(`{\rtf`)):
		return "rtf"
	case bytes.HasPrefix(data, []byte("PK\x03\x04")):
		return zipKind(data)
	case bytes.HasPrefix(data, []byte("WEBVTT")):
		return "vtt"
	}
	head := strings.ToLower(string(data[:min(len(data), 512)]))
	if strings.Contains(head, "<html") || strings.Contains(head, "<!doctype html") {
		return "html"
	}
	if srtTiming.MatchString(head) {
		return "srt"
	}
	return "txt"
}

// decodeText turns bytes into a string, honoring UTF-16 byte-order
// marks and falling back to Latin-1 for invalid UTF-8.
func decodeText(data []byte) string {
	switch {
	case bytes.HasPrefix(data, []byte{0xEF, 0xBB, 0xBF}):
		data = data[3:]
	case bytes.HasPrefix(data, []byte{0xFF, 0xFE}), bytes.HasPrefix(data, []byte{0xFE, 0xFF}):
		big := data[0] == 0xFE
		data = data[2:]
		u := make([]uint16, len(data)/2)
		for i := range u {
			if big {
				u[i] = uint16(data[2*i])<<8 | uint16(data[2*i+1])
			} else {
				u[i] = uint16(data[2*i+1])<<8 | uint16(data[2*i])
			}
		}
		return string(utf16.Decode(u))
	}
	if utf8.Valid(data) {
		return string(data)
	}
	r := make([]rune, len(data))
	for i, b := range data {
		r[i] = rune(b)
	}
	return string(r)
}

// looksLikeText rejects binary files: text has no NUL bytes outside
// UTF-16.
func looksLikeText(data []byte) bool {
	if bytes.HasPrefix(data, []byte{0xFF, 0xFE}) || bytes.HasPrefix(data, []byte{0xFE, 0xFF}) {
		return true
	}
	return !bytes.Contains(data[:min(len(data), 8192)], []byte{0})
}

var (
	blankLines = regexp.MustCompile(`\n{3,}`)
	lineSpaces = regexp.MustCompile(`[ \t\f\v\x{00a0}]+`)
	srtTiming  = regexp.MustCompile(`\d{1,2}:\d{2}:\d{2}[,.]\d{3}\s*-->`)
)

// tidy normalizes line endings and whitespace, keeping paragraphs.
func tidy(s string) string {
	s = strings.ReplaceAll(s, "\r\n", "\n")
	s = strings.ReplaceAll(s, "\r", "\n")
	lines := strings.Split(s, "\n")
	for i, l := range lines {
		lines[i] = strings.TrimSpace(lineSpaces.ReplaceAllString(l, " "))
	}
	s = strings.Join(lines, "\n")
	return strings.TrimSpace(blankLines.ReplaceAllString(s, "\n\n"))
}

func orDefault(s, def string) string {
	if s == "" {
		return def
	}
	return filepath.Base(s)
}
