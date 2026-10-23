package optionfield

import (
	"github.com/mehran9675/sonora/internal/engines"
	"github.com/mehran9675/sonora/ui/components/toggle"
)

func newBool(opt engines.Option, st Store) control {
	on, _ := valueOr(st.Get(opt.ID), opt.Default).(bool)
	return toggle.New(opt.Label, on, func(v bool) { st.Set(opt.ID, v) })
}
