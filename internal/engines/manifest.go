// Package engines defines what a text-to-speech engine is: its
// manifest (name, install recipe, capabilities, options) and the
// Engine interface every backend implements.
package engines

// Kind selects the backend that runs an engine.
type Kind string

const (
	// KindPython runs an adapter.py in a managed Python environment.
	KindPython Kind = "python"
	// KindOnDevice runs downloaded ONNX models inside the app (sherpa-onnx);
	// it's how phones run models.
	KindOnDevice Kind = "ondevice"
	// KindOpenAI calls an OpenAI-compatible /v1/audio/speech endpoint.
	KindOpenAI Kind = "openai"
	// KindCommand runs a command-line program.
	KindCommand Kind = "command"
	// KindSystem uses the operating system's built-in voices.
	KindSystem Kind = "system"
)

// Clone describes whether an engine accepts a reference recording.
type Clone string

const (
	CloneNone     Clone = ""
	CloneOptional Clone = "optional"
	CloneRequired Clone = "required"
)

// Manifest describes an engine. Built-in Python engines ship one as
// engine.json; plugins in the data folder's plugins/ use the same
// format; user-added API and command engines are stored in settings.
type Manifest struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Vendor      string   `json:"vendor,omitempty"`
	Description string   `json:"description,omitempty"`
	Tags        []string `json:"tags,omitempty"`
	License     string   `json:"license,omitempty"`
	Homepage    string   `json:"homepage,omitempty"`
	Size        string   `json:"size,omitempty"`
	// Accent is a two-color gradient for the engine's badge.
	Accent []string `json:"accent,omitempty"`
	Kind   Kind     `json:"kind"`

	// Python engines.
	Python  string            `json:"python,omitempty"`
	Install [][]string        `json:"install,omitempty"`
	Source  *Source           `json:"source,omitempty"`
	Env     map[string]string `json:"env,omitempty"`

	// Clone says whether a reference recording can be given;
	// Transcript whether its transcript is optional or required.
	Clone      Clone `json:"clone,omitempty"`
	Transcript Clone `json:"transcript,omitempty"`
	// Speed is true when the engine honors a speaking-rate multiplier.
	Speed bool `json:"speed,omitempty"`
	// Voices is a fixed voice list; empty means the engine reports
	// its voices once loaded.
	Voices       []Voice  `json:"voices,omitempty"`
	DefaultVoice string   `json:"default_voice,omitempty"`
	Hint         string   `json:"hint,omitempty"`
	Options      []Option `json:"options,omitempty"`

	// Dir is the folder holding engine.json and adapter.py; Custom
	// marks engines the user created (they can be deleted).
	Dir    string `json:"-"`
	Custom bool   `json:"custom,omitempty"`
}

// Source is an archive of code an engine needs, unpacked into the
// engine's src/ folder at install time ({src} in Install arguments).
type Source struct {
	URL string `json:"url"`
}

// NeedsInstall reports whether the engine has an install step.
func (m *Manifest) NeedsInstall() bool { return m.Kind == KindPython || m.Kind == KindOnDevice }

// Option returns the option with the given id, or nil.
func (m *Manifest) Option(id string) *Option {
	for i := range m.Options {
		if m.Options[i].ID == id {
			return &m.Options[i]
		}
	}
	return nil
}
