package engines

import "context"

// ModelStore is implemented by engines whose model is a choice the
// user makes (the manifest's model option), downloaded per model: the
// engine card asks which to download, and the main page defaults to
// one that is.
type ModelStore interface {
	// DownloadedModels reports which choices of the model option are
	// on this device.
	DownloadedModels() map[string]bool
	// InstallModel installs the engine if it isn't yet, then downloads
	// the model with the given choice value.
	InstallModel(ctx context.Context, model string, log Logger) error
}

// ModelOption returns the option that picks the engine's model: the
// one named by model_option, or a select named "model". It returns nil
// when the model isn't a choice from a list.
func (m *Manifest) ModelOption() *Option {
	id := m.ModelOptionID
	if id == "" {
		id = "model"
	}
	if o := m.Option(id); o != nil && o.Type == OptionSelect && len(o.Choices) > 0 {
		return o
	}
	return nil
}
