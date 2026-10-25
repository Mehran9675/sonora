package textfile

import (
	"archive/zip"
	"bytes"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"path"
	"strings"
	"unicode"
	"unicode/utf8"

	pdflib "github.com/ledongthuc/pdf"
)

// zipKind tells DOCX, ODT and EPUB apart by their contents.
func zipKind(data []byte) string {
	z, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return "txt"
	}
	for _, f := range z.File {
		switch f.Name {
		case "word/document.xml":
			return "docx"
		case "content.xml":
			return "odt"
		case "META-INF/container.xml":
			return "epub"
		}
	}
	return "zip"
}

func openZip(data []byte) (*zip.Reader, error) {
	z, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return nil, errors.New("the document is damaged or not what its name says")
	}
	return z, nil
}

func zipFile(z *zip.Reader, name string) ([]byte, error) {
	f, err := z.Open(name)
	if err != nil {
		return nil, fmt.Errorf("the document has no %s", name)
	}
	defer f.Close()
	return io.ReadAll(io.LimitReader(f, maxSize))
}

// xmlText walks an XML document: text from character data inside
// elements named textEl, a line break at the end of elements in
// breakEl, and replacements for empty elements in emptyEl.
func xmlText(data []byte, textEl string, breakEl map[string]bool, emptyEl map[string]string) string {
	d := xml.NewDecoder(bytes.NewReader(data))
	d.Strict = false
	var b strings.Builder
	depth := 0 // inside textEl (any text when textEl is "")
	for {
		tok, err := d.Token()
		if err != nil {
			return b.String()
		}
		switch t := tok.(type) {
		case xml.StartElement:
			if t.Name.Local == textEl {
				depth++
			}
			if r, ok := emptyEl[t.Name.Local]; ok {
				b.WriteString(r)
			}
		case xml.EndElement:
			if t.Name.Local == textEl && depth > 0 {
				depth--
			}
			if breakEl[t.Name.Local] {
				b.WriteString("\n")
			}
		case xml.CharData:
			if textEl == "" || depth > 0 {
				b.Write(t)
			}
		}
	}
}

func docx(data []byte) (string, error) {
	z, err := openZip(data)
	if err != nil {
		return "", err
	}
	doc, err := zipFile(z, "word/document.xml")
	if err != nil {
		return "", err
	}
	return xmlText(doc, "t", map[string]bool{"p": true}, map[string]string{"tab": " ", "br": "\n", "cr": "\n"}), nil
}

func odt(data []byte) (string, error) {
	z, err := openZip(data)
	if err != nil {
		return "", err
	}
	doc, err := zipFile(z, "content.xml")
	if err != nil {
		return "", err
	}
	// Only the body: styles and fonts sit in office:automatic-styles,
	// which holds no character data, so all text is the document's.
	return xmlText(doc, "", map[string]bool{"p": true, "h": true}, map[string]string{"s": " ", "tab": " ", "line-break": "\n"}), nil
}

// epub reads the book's chapters in reading order.
func epub(data []byte) (string, error) {
	z, err := openZip(data)
	if err != nil {
		return "", err
	}
	container, err := zipFile(z, "META-INF/container.xml")
	if err != nil {
		return "", err
	}
	var c struct {
		Rootfiles []struct {
			Path string `xml:"full-path,attr"`
		} `xml:"rootfiles>rootfile"`
	}
	if xml.Unmarshal(container, &c) != nil || len(c.Rootfiles) == 0 {
		return "", errors.New("the book has no table of contents")
	}
	opfPath := c.Rootfiles[0].Path
	opfData, err := zipFile(z, opfPath)
	if err != nil {
		return "", err
	}
	var opf struct {
		Items []struct {
			ID   string `xml:"id,attr"`
			Href string `xml:"href,attr"`
		} `xml:"manifest>item"`
		Spine []struct {
			IDRef string `xml:"idref,attr"`
		} `xml:"spine>itemref"`
	}
	if err := xml.Unmarshal(opfData, &opf); err != nil {
		return "", errors.New("the book's contents can't be read")
	}
	hrefs := map[string]string{}
	for _, it := range opf.Items {
		hrefs[it.ID] = it.Href
	}
	var parts []string
	for _, ref := range opf.Spine {
		href, ok := hrefs[ref.IDRef]
		if !ok {
			continue
		}
		name := path.Join(path.Dir(opfPath), href)
		if i := strings.IndexByte(name, '#'); i >= 0 {
			name = name[:i]
		}
		chapter, err := zipFile(z, name)
		if err != nil {
			continue
		}
		parts = append(parts, html(chapter))
	}
	return strings.Join(parts, "\n\n"), nil
}

// pdf extracts the text layer page by page. Scanned PDFs have none.
func pdf(data []byte) (text string, err error) {
	defer func() {
		// The PDF parser panics on some malformed files.
		if recover() != nil {
			text, err = "", errors.New("this PDF can't be read")
		}
	}()
	r, err := pdflib.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return "", errors.New("this PDF can't be read: " + err.Error())
	}
	var b strings.Builder
	for i := 1; i <= r.NumPage(); i++ {
		p := r.Page(i)
		if p.V.IsNull() {
			continue
		}
		t, err := p.GetPlainText(nil)
		if err != nil {
			continue
		}
		b.WriteString(t)
		b.WriteString("\n\n")
	}
	if strings.TrimSpace(b.String()) == "" {
		return "", errors.New("this PDF has no text layer (it may be scanned images)")
	}
	return joinPDFLines(b.String()), nil
}

// joinPDFLines undoes the line breaks a PDF puts inside sentences: a
// line that starts in lower case or with punctuation continues the one
// before, and lone bullets are dropped.
func joinPDFLines(s string) string {
	var out []string
	for _, l := range strings.Split(s, "\n") {
		t := strings.TrimSpace(l)
		if t == "•" || t == "·" || t == "◦" || t == "▪" {
			continue
		}
		if n := len(out); n > 0 && t != "" && out[n-1] != "" {
			r, _ := utf8.DecodeRuneInString(t)
			switch {
			case strings.ContainsRune(".,;:!?)]’”", r):
				out[n-1] += t
				continue
			case unicode.IsLower(r):
				out[n-1] += " " + t
				continue
			}
		}
		out = append(out, t)
	}
	return strings.Join(out, "\n")
}
