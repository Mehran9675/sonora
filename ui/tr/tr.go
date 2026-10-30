// Package tr translates Sonora's text. Messages are keyed by their
// English text, so English needs no catalog except for plural forms:
//
//	tr.T("Recording")                         // "Aufnahme" in German
//	tr.T("Downloading {file}", "file", name)  // placeholders by name
//	tr.N("{n} words", count)                  // plural forms; {n} is count
//
// Text laid out every frame calls T where it's drawn. Elements built
// once (buttons, inputs) take their text through Bind, which sets it
// again whenever the language changes.
package tr

import (
	"embed"
	"log"
	"strings"
	"sync"

	"github.com/mehran9675/giode/i18n"
)

//go:embed locales/*.json
var files embed.FS

// Language is a language Sonora is translated into.
type Language struct {
	Code string // "de", "pt", "nb"
	Name string // its own name for itself: "Deutsch"
}

// Languages are the translations, English first, then by name.
var Languages = []Language{
	{"en", "English"},
	{"bg", "Български"},
	{"ca", "Català"},
	{"cs", "Čeština"},
	{"da", "Dansk"},
	{"de", "Deutsch"},
	{"et", "Eesti"},
	{"el", "Ελληνικά"},
	{"es", "Español"},
	{"fr", "Français"},
	{"hr", "Hrvatski"},
	{"it", "Italiano"},
	{"lv", "Latviešu"},
	{"lt", "Lietuvių"},
	{"hu", "Magyar"},
	{"nl", "Nederlands"},
	{"nb", "Norsk bokmål"},
	{"pl", "Polski"},
	{"pt", "Português"},
	{"ro", "Română"},
	{"ru", "Русский"},
	{"sk", "Slovenčina"},
	{"sl", "Slovenščina"},
	{"fi", "Suomi"},
	{"sv", "Svenska"},
	{"tr", "Türkçe"},
	{"uk", "Українська"},
}

func init() {
	if err := i18n.Load(files, "locales"); err != nil {
		log.Printf("translations: %v", err)
	}
}

// Use switches to the language code, or the system's when code is "".
// A system language Sonora isn't translated into reads in English.
func Use(code string) {
	if code == "" {
		code = System()
	}
	i18n.SetLocale(code)
}

// System returns the system's language as one of Languages' codes, or
// "en".
func System() string {
	loc := i18n.SystemLocale()
	lang, _, _ := strings.Cut(loc, "-")
	switch lang {
	case "no", "nn": // Norwegian: Sonora has Bokmål
		lang = "nb"
	}
	for _, l := range Languages {
		if l.Code == lang {
			return lang
		}
	}
	return "en"
}

// Current returns the language in use.
func Current() string {
	lang, _, _ := strings.Cut(i18n.Locale(), "-")
	return lang
}

// Name returns a language's own name.
func Name(code string) string {
	for _, l := range Languages {
		if l.Code == code {
			return l.Name
		}
	}
	return code
}

// T returns key's translation with its {placeholders} filled from
// name, value pairs.
func T(key string, args ...any) string { return i18n.T(key, args...) }

// N returns the plural form of key for n; {n} is n.
func N(key string, n int, args ...any) string { return i18n.N(key, n, args...) }

var (
	mu    sync.Mutex
	binds []func()
	once  sync.Once
)

// Bind sets an element's text to T(key, args...) now and again after
// every language change. Use it for elements built once; text drawn
// every frame calls T instead.
//
//	tr.Bind(func(s string) { b.Label(s) }, "Record")
func Bind(set func(string), key string, args ...any) {
	once.Do(func() { i18n.OnChange(rebind) })
	f := func() { set(T(key, args...)) }
	f()
	mu.Lock()
	binds = append(binds, f)
	mu.Unlock()
}

func rebind() {
	mu.Lock()
	fs := append([]func(){}, binds...)
	mu.Unlock()
	for _, f := range fs {
		f()
	}
}

// Msg marks text that's translated later, where it's shown (an error
// message, a status): it returns key unchanged, and the catalog check
// finds it.
func Msg(key string) string { return key }
