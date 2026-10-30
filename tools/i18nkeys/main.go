// Command i18nkeys prints the text Sonora translates as JSON, for
// translators; with -missing CODE, only the keys locale CODE lacks.
//
//	go run ./tools/i18nkeys -missing de
package main

import (
	"encoding/json"
	"flag"
	"log"
	"os"
	"path/filepath"

	"github.com/mehran9675/sonora/tools/i18nkeys/keys"
)

func main() {
	missing := flag.String("missing", "", "print only the keys this locale's file lacks")
	flag.Parse()
	ks, err := keys.Find(".")
	if err != nil {
		log.Fatal(err)
	}
	if *missing != "" {
		have := map[string]any{}
		if b, err := os.ReadFile(filepath.Join("ui", "tr", "locales", *missing+".json")); err == nil {
			if err := json.Unmarshal(b, &have); err != nil {
				log.Fatal(err)
			}
		}
		var out []keys.Key
		for _, k := range ks {
			if _, ok := have[k.Text]; !ok {
				out = append(out, k)
			}
		}
		ks = out
	}
	enc := json.NewEncoder(os.Stdout)
	enc.SetEscapeHTML(false)
	enc.SetIndent("", "  ")
	if err := enc.Encode(ks); err != nil {
		log.Fatal(err)
	}
}
