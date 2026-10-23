package optionfield

import (
	"fmt"

	"github.com/mehran9675/giode"

	"github.com/mehran9675/sonora/internal/engines"
	"github.com/mehran9675/sonora/ui/components/picker"
	"github.com/mehran9675/sonora/ui/components/segmented"
)

// newSelect uses segments for up to three choices and a picker beyond.
func newSelect(opt engines.Option, st Store) control {
	current := func() string { return fmt.Sprint(valueOr(st.Get(opt.ID), opt.Default)) }
	if len(opt.Choices) <= 3 {
		var segs []segmented.Option
		for _, c := range opt.Choices {
			segs = append(segs, segmented.Option{Value: c.Value, Label: c.Label})
		}
		return segmented.New(segs, current, func(v string) { st.Set(opt.ID, v) })
	}
	items := func() []picker.Item {
		var out []picker.Item
		for _, c := range opt.Choices {
			out = append(out, picker.Item{ID: c.Value, Label: c.Label})
		}
		return out
	}
	pk := picker.New(opt.Label, items, current, func(it picker.Item) { st.Set(opt.ID, it.ID) })
	return picker.NewTrigger(opt.Label, func() (string, string) {
		cur := current()
		for _, c := range opt.Choices {
			if c.Value == cur {
				return c.Label, ""
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
