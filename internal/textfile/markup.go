package textfile

import (
	"bytes"
	"regexp"
	"strconv"
	"strings"

	nethtml "golang.org/x/net/html"
)

var (
	mdImage    = regexp.MustCompile(`!\[[^\]]*\]\([^)]*\)`)
	mdLink     = regexp.MustCompile(`\[([^\]]*)\]\([^)]*\)`)
	mdHeading  = regexp.MustCompile(`^\s{0,3}#{1,6}\s+`)
	mdBullet   = regexp.MustCompile(`^\s*(?:[-*+]|\d+[.)])\s+`)
	mdQuote    = regexp.MustCompile(`^\s*>\s?`)
	mdRule     = regexp.MustCompile(`^\s*(?:[-*_]\s*){3,}$`)
	mdEmphasis = regexp.MustCompile("(\\*\\*|__|\\*|_|~~|`)")
	htmlTag    = regexp.MustCompile(`<[^>]+>`)
)

// markdown keeps what a reader would say: no markup, links as their
// text, code blocks dropped.
func markdown(s string) string {
	var out []string
	fence := false
	for _, l := range strings.Split(strings.ReplaceAll(s, "\r\n", "\n"), "\n") {
		if strings.HasPrefix(strings.TrimSpace(l), "```") || strings.HasPrefix(strings.TrimSpace(l), "~~~") {
			fence = !fence
			continue
		}
		if fence || mdRule.MatchString(l) {
			continue
		}
		heading := mdHeading.MatchString(l)
		l = mdHeading.ReplaceAllString(l, "")
		l = mdBullet.ReplaceAllString(l, "")
		l = mdQuote.ReplaceAllString(l, "")
		l = mdImage.ReplaceAllString(l, "")
		l = mdLink.ReplaceAllString(l, "$1")
		l = htmlTag.ReplaceAllString(l, "")
		l = mdEmphasis.ReplaceAllString(l, "")
		if heading && l != "" {
			// A heading is its own sentence.
			if !strings.ContainsAny(l[len(l)-1:], ".!?:") {
				l += "."
			}
			out = append(out, "", l, "")
			continue
		}
		out = append(out, l)
	}
	return strings.Join(out, "\n")
}

// subtitles keeps the spoken lines of SRT and WebVTT files.
func subtitles(s string) string {
	var out []string
	skip := false // inside a WebVTT NOTE or STYLE block
	for _, l := range strings.Split(strings.ReplaceAll(s, "\r\n", "\n"), "\n") {
		t := strings.TrimSpace(l)
		switch {
		case t == "":
			skip = false
			out = append(out, "")
			continue
		case skip:
			continue
		case strings.HasPrefix(t, "WEBVTT"):
			continue
		case strings.HasPrefix(t, "NOTE") || t == "STYLE" || t == "REGION":
			skip = true
			continue
		case strings.Contains(t, "-->"):
			continue
		}
		if _, err := strconv.Atoi(t); err == nil {
			continue // cue number
		}
		out = append(out, htmlTag.ReplaceAllString(t, ""))
	}
	// Cues are separated by blank lines; join each cue's lines.
	return joinCues(out)
}

func joinCues(lines []string) string {
	var paras []string
	var cur []string
	for _, l := range append(lines, "") {
		if l == "" {
			if len(cur) > 0 {
				paras = append(paras, strings.Join(cur, " "))
				cur = nil
			}
			continue
		}
		cur = append(cur, l)
	}
	return strings.Join(paras, "\n")
}

// block are the HTML elements that start a new line.
var block = map[string]bool{
	"p": true, "div": true, "br": true, "li": true, "h1": true, "h2": true, "h3": true, "h4": true, "h5": true,
	"h6": true, "tr": true, "section": true, "article": true, "blockquote": true, "pre": true, "dd": true, "dt": true,
	"header": true, "footer": true, "figcaption": true, "td": true, "th": true, "hr": true, "title": true,
}

// html returns an HTML (or XHTML) document's visible text.
func html(data []byte) string {
	z := nethtml.NewTokenizer(bytes.NewReader(data))
	var b strings.Builder
	skip := 0
	for {
		switch z.Next() {
		case nethtml.ErrorToken:
			return b.String()
		case nethtml.StartTagToken, nethtml.SelfClosingTagToken:
			name, _ := z.TagName()
			switch string(name) {
			case "script", "style", "head", "noscript", "svg":
				if t := z.Token(); t.Type == nethtml.StartTagToken {
					skip++
				}
			}
			if block[string(name)] {
				b.WriteString("\n")
			}
		case nethtml.EndTagToken:
			name, _ := z.TagName()
			switch string(name) {
			case "script", "style", "head", "noscript", "svg":
				if skip > 0 {
					skip--
				}
			}
			if block[string(name)] {
				b.WriteString("\n")
			}
		case nethtml.TextToken:
			if skip == 0 {
				b.WriteString(strings.ReplaceAll(string(z.Text()), "\n", " "))
			}
		}
	}
}

// rtf strips RTF control words and groups that hold no text.
func rtf(s string) string {
	var b strings.Builder
	// Groups whose content is not text (font tables, pictures…).
	ignore := map[string]bool{"fonttbl": true, "colortbl": true, "stylesheet": true, "info": true, "pict": true,
		"header": true, "footer": true, "listtable": true, "listoverridetable": true, "rsidtbl": true, "generator": true,
		"xmlnstbl": true, "themedata": true, "colorschememapping": true, "datastore": true, "latentstyles": true}
	type group struct{ skip bool }
	stack := []group{{}}
	skipNext := 0 // characters to skip after \uN
	for i := 0; i < len(s); i++ {
		c := s[i]
		top := &stack[len(stack)-1]
		switch c {
		case '{':
			stack = append(stack, group{skip: top.skip})
			if strings.HasPrefix(s[i:], `{\*`) {
				stack[len(stack)-1].skip = true
			}
		case '}':
			if len(stack) > 1 {
				stack = stack[:len(stack)-1]
			}
		case '\\':
			if i+1 >= len(s) {
				break
			}
			n := s[i+1]
			switch {
			case n == '\\' || n == '{' || n == '}':
				if !top.skip {
					b.WriteByte(n)
				}
				i++
			case n == '\'' && i+3 < len(s):
				if v, err := strconv.ParseUint(s[i+2:i+4], 16, 8); err == nil && !top.skip && skipNext == 0 {
					b.WriteRune(rune(v)) // Windows-1252, close enough to Latin-1
				}
				if skipNext > 0 {
					skipNext--
				}
				i += 3
			case n == '~':
				if !top.skip {
					b.WriteByte(' ')
				}
				i++
			case (n >= 'a' && n <= 'z') || (n >= 'A' && n <= 'Z'):
				j := i + 1
				for j < len(s) && ((s[j] >= 'a' && s[j] <= 'z') || (s[j] >= 'A' && s[j] <= 'Z')) {
					j++
				}
				word := s[i+1 : j]
				k := j
				if k < len(s) && (s[k] == '-' || (s[k] >= '0' && s[k] <= '9')) {
					k++
					for k < len(s) && s[k] >= '0' && s[k] <= '9' {
						k++
					}
				}
				arg := s[j:k]
				if k < len(s) && s[k] == ' ' {
					k++
				}
				i = k - 1
				switch {
				case ignore[word]:
					top.skip = true
				case top.skip:
				case word == "par" || word == "line" || word == "sect" || word == "page":
					b.WriteByte('\n')
				case word == "tab" || word == "cell":
					b.WriteByte(' ')
				case word == "u":
					if v, err := strconv.Atoi(arg); err == nil {
						if v < 0 {
							v += 65536
						}
						b.WriteRune(rune(v))
						skipNext = 1
					}
				}
			default:
				i++
			}
		case '\r', '\n':
		default:
			if skipNext > 0 {
				skipNext--
				continue
			}
			if !top.skip {
				b.WriteByte(c)
			}
		}
	}
	return b.String()
}
