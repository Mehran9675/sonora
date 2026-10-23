// Package picker is a searchable list in a modal dialog, used for
// voices, engines and long option lists.
package picker

import "strings"

// Item is one choice.
type Item struct {
	ID     string
	Label  string
	Group  string
	Detail string
}

func (it Item) matches(q string) bool {
	if q == "" {
		return true
	}
	for _, s := range []string{it.Label, it.ID, it.Group, it.Detail} {
		if strings.Contains(strings.ToLower(s), q) {
			return true
		}
	}
	return false
}
