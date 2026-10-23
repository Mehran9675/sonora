package manager

import (
	"errors"
	"strings"
)

// hints turn common low-level failures into advice a non-programmer
// can act on. The original error stays in the engine log.
var hints = []struct {
	match []string
	text  string
}{
	{[]string{"ProxyError", "ConnectionError", "Max retries exceeded", "NameResolutionError", "Temporary failure in name resolution", "LocalEntryNotFoundError"},
		"Couldn't download the model. Check your internet connection and try again."},
	{[]string{"Failed to fetch", "error sending request", "tunnel error", "dns error", "Connection refused"},
		"Couldn't download the engine's files. Check your internet connection and try again."},
	{[]string{"GatedRepoError", "401 Client Error", "Access to model"},
		"This model needs a Hugging Face account. Accept its license on huggingface.co and add your token in Settings."},
	{[]string{"CUDA out of memory", "OutOfMemoryError", "MPS backend out of memory"},
		"The GPU ran out of memory. Switch the compute device to CPU in Settings, or try a smaller model."},
	{[]string{"No module named", "ModuleNotFoundError", "ImportError"},
		"A Python package is missing. Reinstall the engine from the Engines page."},
	{[]string{"No space left on device"},
		"The disk is full. Free some space and try again."},
}

// explain prefixes err with a plain-language hint when one applies.
func explain(err error) error {
	if err == nil {
		return nil
	}
	msg := err.Error()
	for _, h := range hints {
		for _, m := range h.match {
			if strings.Contains(msg, m) {
				return errors.New(h.text + "\n" + msg)
			}
		}
	}
	return err
}
