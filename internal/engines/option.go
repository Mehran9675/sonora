package engines

import (
	"fmt"
	"strconv"
	"strings"
)

// OptionType selects the control an option is edited with.
type OptionType string

const (
	OptionSelect OptionType = "select"
	OptionNumber OptionType = "number"
	OptionText   OptionType = "text"
	OptionSecret OptionType = "secret"
	OptionBool   OptionType = "bool"
)

// Scope says where an option is edited and what changing it costs.
type Scope string

const (
	// ScopeSynth options apply per request (the default).
	ScopeSynth Scope = ""
	// ScopeLoad options require reloading the model.
	ScopeLoad Scope = "load"
	// ScopeSetup options configure the engine itself (an endpoint, a
	// command) and are edited on the Engines page.
	ScopeSetup Scope = "setup"
)

// Option is one engine setting, rendered generically by the UI.
type Option struct {
	ID          string     `json:"id"`
	Label       string     `json:"label"`
	Type        OptionType `json:"type"`
	Scope       Scope      `json:"scope,omitempty"`
	Default     any        `json:"default,omitempty"`
	Choices     []Choice   `json:"choices,omitempty"`
	Suggestions []string   `json:"suggestions,omitempty"`
	Min         float64    `json:"min,omitempty"`
	Max         float64    `json:"max,omitempty"`
	Step        float64    `json:"step,omitempty"`
	Help        string     `json:"help,omitempty"`
	Placeholder string     `json:"placeholder,omitempty"`
}

// Choice is one value of a select option.
type Choice struct {
	Value string `json:"value"`
	Label string `json:"label"`
}

// Values holds option values by id, as stored in settings and sent to
// adapters.
type Values map[string]any

// Resolve fills in defaults for every option missing from v.
func Resolve(opts []Option, v Values) Values {
	out := Values{}
	for _, o := range opts {
		if val, ok := v[o.ID]; ok {
			out[o.ID] = val
		} else if o.Default != nil {
			out[o.ID] = o.Default
		}
	}
	return out
}

// String returns a value as text.
func (v Values) String(id string) string {
	switch x := v[id].(type) {
	case nil:
		return ""
	case string:
		return x
	case float64:
		return strconv.FormatFloat(x, 'f', -1, 64)
	default:
		return fmt.Sprint(x)
	}
}

// Float returns a numeric value, or def.
func (v Values) Float(id string, def float64) float64 {
	switch x := v[id].(type) {
	case float64:
		return x
	case int:
		return float64(x)
	case string:
		if f, err := strconv.ParseFloat(strings.TrimSpace(x), 64); err == nil {
			return f
		}
	}
	return def
}

// Bool returns a boolean value.
func (v Values) Bool(id string) bool {
	b, _ := v[id].(bool)
	return b
}

// Subset returns the values of options in scope.
func Subset(opts []Option, v Values, scope Scope) Values {
	out := Values{}
	for _, o := range opts {
		if o.Scope == scope {
			if val, ok := v[o.ID]; ok {
				out[o.ID] = val
			}
		}
	}
	return out
}
