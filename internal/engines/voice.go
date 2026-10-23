package engines

// Voice is a speaker an engine offers.
type Voice struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Group       string `json:"group,omitempty"`
	Description string `json:"description,omitempty"`
}

// Label returns the voice's display name.
func (v Voice) Label() string {
	if v.Name != "" {
		return v.Name
	}
	if v.ID == "" {
		return "Default"
	}
	return v.ID
}
