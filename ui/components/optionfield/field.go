// Package optionfield renders an engine option (from its manifest) as
// the matching control, so any engine's settings get a UI for free.
package optionfield

import (
	"github.com/mehran9675/giode"

	"github.com/mehran9675/sonora/internal/engines"
	"github.com/mehran9675/sonora/ui/theme"
)

// Store reads and writes option values.
type Store struct {
	Get func(id string) any
	Set func(id string, v any)
	// Detail, when set, describes a choice of a select option (such as
	// whether a model is downloaded); "" adds nothing.
	Detail func(id, value string) string
}

func (st Store) detail(id, value string) string {
	if st.Detail == nil {
		return ""
	}
	return st.Detail(id, value)
}

// control is one rendered option.
type control interface {
	Element() giode.Element
}

// Field is an option's label, control and help text. Create it once
// per engine option.
type Field struct {
	opt  engines.Option
	ctrl control
}

// New builds the control for opt.
func New(opt engines.Option, st Store) *Field {
	f := &Field{opt: opt}
	switch opt.Type {
	case engines.OptionSelect:
		f.ctrl = newSelect(opt, st)
	case engines.OptionNumber:
		f.ctrl = newNumber(opt, st)
	case engines.OptionBool:
		f.ctrl = newBool(opt, st)
	default:
		f.ctrl = newText(opt, st)
	}
	return f
}

// ID returns the option's id.
func (f *Field) ID() string { return f.opt.ID }

// Element returns the labelled control.
func (f *Field) Element() giode.Element {
	if f.opt.Type == engines.OptionBool {
		return giode.Stack(giode.Styles{Width: -1, Gap: 4}, f.ctrl.Element(), f.help())
	}
	head := giode.Row(giode.Styles{Width: -1, Align: giode.AlignCenter, Wrap: giode.WrapNoWrap},
		giode.Box(giode.Styles{FlexGrow: 1}, giode.Text(f.opt.Label, theme.Overline)))
	if n, ok := f.ctrl.(*number); ok {
		head = giode.Row(giode.Styles{Width: -1, Align: giode.AlignCenter, Wrap: giode.WrapNoWrap},
			giode.Box(giode.Styles{FlexGrow: 1}, giode.Text(f.opt.Label, theme.Overline)),
			giode.Text(n.display(), giode.Styles{Color: "theme.text", FontSize: 12, FontWeight: giode.FontWeightBold}))
	}
	return giode.Stack(giode.Styles{Width: -1, Gap: 8}, head, f.ctrl.Element(), f.help())
}

func (f *Field) help() giode.Element {
	if f.opt.Help == "" {
		return giode.Box(giode.Styles{})
	}
	return giode.Text(f.opt.Help, theme.Caption)
}
