package tr

import (
	"encoding/json"
	"regexp"
	"slices"
	"sort"
	"testing"

	"github.com/mehran9675/sonora/tools/i18nkeys/keys"
)

// placeholder matches {name}.
var placeholder = regexp.MustCompile(`\{[a-z]+\}`)

// pluralForms are the forms each language's plural rules use (CLDR, as
// giode's i18n applies them).
var pluralForms = map[string][]string{
	"cs": {"one", "few", "other"}, "sk": {"one", "few", "other"},
	"pl": {"one", "few", "many"}, "ru": {"one", "few", "many"}, "uk": {"one", "few", "many"},
	"hr": {"one", "few", "many"},
	"lt": {"one", "few", "other"}, "lv": {"zero", "one", "other"},
	"sl": {"one", "two", "few", "other"}, "ro": {"one", "few", "other"},
}

func placeholders(s string) []string {
	p := placeholder.FindAllString(s, -1)
	sort.Strings(p)
	return slices.Compact(p)
}

// TestCatalogs checks every translation has every key the code uses,
// with the same placeholders, and plural forms for counted keys.
func TestCatalogs(t *testing.T) {
	ks, err := keys.Find("../..")
	if err != nil {
		t.Fatal(err)
	}
	if len(ks) < 100 {
		t.Fatalf("found only %d keys", len(ks))
	}
	for _, lang := range Languages {
		if lang.Code == "en" {
			continue
		}
		b, err := files.ReadFile("locales/" + lang.Code + ".json")
		if err != nil {
			t.Errorf("%s: %v", lang.Code, err)
			continue
		}
		var cat map[string]json.RawMessage
		if err := json.Unmarshal(b, &cat); err != nil {
			t.Errorf("%s: %v", lang.Code, err)
			continue
		}
		var missing []string
		for _, k := range ks {
			raw, ok := cat[k.Text]
			if !ok {
				missing = append(missing, k.Text)
				continue
			}
			want := placeholders(k.Text)
			var text string
			var forms map[string]string
			if json.Unmarshal(raw, &text) == nil {
				forms = map[string]string{"other": text}
				if k.Plural && slices.Contains(want, "{n}") {
					t.Errorf("%s: %q needs plural forms", lang.Code, k.Text)
				}
			} else if err := json.Unmarshal(raw, &forms); err != nil {
				t.Errorf("%s: %q: %v", lang.Code, k.Text, err)
				continue
			} else {
				need := pluralForms[lang.Code]
				if need == nil {
					need = []string{"one", "other"}
				}
				for _, f := range need {
					if forms[f] == "" {
						t.Errorf("%s: %q lacks the %q form", lang.Code, k.Text, f)
					}
				}
			}
			for form, s := range forms {
				if s == "" {
					t.Errorf("%s: %q (%s) is empty", lang.Code, k.Text, form)
				}
				got := placeholders(s)
				// A singular form may say "one" instead of {n}.
				if form == "one" || form == "zero" || form == "two" {
					got = slices.DeleteFunc(got, func(p string) bool { return p == "{n}" })
					want := slices.DeleteFunc(slices.Clone(want), func(p string) bool { return p == "{n}" })
					if !slices.Equal(got, want) {
						t.Errorf("%s: %q (%s) has placeholders %v, want %v", lang.Code, k.Text, form, got, want)
					}
					continue
				}
				if !slices.Equal(got, want) {
					t.Errorf("%s: %q (%s) has placeholders %v, want %v", lang.Code, k.Text, form, got, want)
				}
			}
		}
		if len(missing) > 0 {
			t.Errorf("%s lacks %d keys, e.g. %q (go run ./tools/i18nkeys -missing %s)", lang.Code, len(missing), missing[0], lang.Code)
		}
	}
}

// TestSwitch checks a switch reaches T and Bind.
func TestSwitch(t *testing.T) {
	defer Use("en")
	var label string
	Bind(func(s string) { label = s }, "Settings")
	Use("de")
	if got := T("Settings"); got == "Settings" || label != got {
		t.Errorf("German: T = %q, bound label = %q", got, label)
	}
	Use("en")
	if label != "Settings" {
		t.Errorf("back in English the label is %q", label)
	}
	if got := N("{n} words", 1); got != "1 word" {
		t.Errorf(`N("{n} words", 1) = %q`, got)
	}
}
