package wav

import (
	"bytes"
	"encoding/binary"
	"os"

	"github.com/mehran9675/sonora/internal/audio"
)

// Write saves the clip as 16-bit mono PCM.
func Write(path string, c *audio.Clip) error {
	var buf bytes.Buffer
	le := binary.LittleEndian
	dataLen := uint32(len(c.Samples) * 2)
	buf.WriteString("RIFF")
	binary.Write(&buf, le, 36+dataLen)
	buf.WriteString("WAVEfmt ")
	binary.Write(&buf, le, uint32(16))
	binary.Write(&buf, le, uint16(1)) // PCM
	binary.Write(&buf, le, uint16(1)) // mono
	binary.Write(&buf, le, uint32(c.SampleRate))
	binary.Write(&buf, le, uint32(c.SampleRate*2))
	binary.Write(&buf, le, uint16(2))
	binary.Write(&buf, le, uint16(16))
	buf.WriteString("data")
	binary.Write(&buf, le, dataLen)
	for _, s := range c.Samples {
		binary.Write(&buf, le, int16(max(-1, min(1, s))*32767))
	}
	return os.WriteFile(path, buf.Bytes(), 0o644)
}
