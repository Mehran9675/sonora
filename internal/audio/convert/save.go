package convert

import (
	"os"

	"github.com/mehran9675/sonora/internal/audio/wav"
)

// SaveAsWAV writes audio bytes to path as WAV, converting them first
// when they are in another format.
func SaveAsWAV(path string, data []byte) error {
	if !wav.IsWAV(data) {
		converted, err := ToWAV(data)
		if err != nil {
			return err
		}
		data = converted
	}
	return os.WriteFile(path, data, 0o644)
}
