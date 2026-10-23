package engines

// Request is one synthesis job.
type Request struct {
	Text          string  `json:"text"`
	Voice         string  `json:"voice,omitempty"`
	Speed         float64 `json:"speed,omitempty"`
	Reference     string  `json:"reference,omitempty"`
	ReferenceText string  `json:"reference_text,omitempty"`
	Options       Values  `json:"options,omitempty"`
	// Output is the WAV file the engine must write.
	Output string `json:"output"`
}

// Environment carries app-wide settings engines may need.
type Environment struct {
	// Device is "auto", "cpu" or "gpu".
	Device string
	// HFToken authenticates Hugging Face downloads of gated models.
	HFToken string
}
