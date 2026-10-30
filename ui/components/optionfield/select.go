package optionfield

import (
	"fmt"
	"unicode/utf8"

	"github.com/mehran9675/giode"

	"github.com/mehran9675/sonora/internal/engines"
	"github.com/mehran9675/sonora/ui/components/picker"
	"github.com/mehran9675/sonora/ui/components/segmented"
	"github.com/mehran9675/sonora/ui/tr"
)

// segmentChars is the most label text, in characters, that fits
// segments across a phone's width.
const segmentChars = 30

// newSelect uses segments for up to three short choices and a picker
// otherwise, or whenever the store describes choices (Store.Detail).
// Segments, the picker's title and the trigger's name translate the
// labels they keep at render; the items translate here.
func newSelect(opt engines.Option, st Store) control {
	current := func() string { return fmt.Sprint(valueOr(st.Get(opt.ID), opt.Default)) }
	chars, described := 0, false
	for _, c := range opt.Choices {
		chars += utf8.RuneCountInString(c.Label)
		described = described || st.detail(opt.ID, c.Value) != ""
	}
	if len(opt.Choices) <= 3 && chars <= segmentChars && !described {
		var segs []segmented.Option
		for _, c := range opt.Choices {
			segs = append(segs, segmented.Option{Value: c.Value, Label: c.Label})
		}
		return segmented.New(segs, current, func(v string) { st.Set(opt.ID, v) })
	}
	items := func() []picker.Item {
		var out []picker.Item
		for _, c := range opt.Choices {
			out = append(out, picker.Item{ID: c.Value, Label: tr.T(c.Label), Detail: st.detail(opt.ID, c.Value)})
		}
		return out
	}
	pk := picker.New(opt.Label, items, current, func(it picker.Item) { st.Set(opt.ID, it.ID) })
	return picker.NewTrigger(opt.Label, func() (string, string) {
		cur := current()
		for _, c := range opt.Choices {
			if c.Value == cur {
				return tr.T(c.Label), st.detail(opt.ID, cur)
			}
		}
		return cur, ""
	}, func() giode.Element { return nil }, pk.Open)
}

func valueOr(v, def any) any {
	if v == nil {
		return def
	}
	return v
}
