// Package wav reads and writes WAV files.
package wav

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"math"
	"os"

	"github.com/mehran9675/sonora/internal/audio"
)

type format struct {
	tag, channels, bits uint16
	rate                uint32
}

// Read decodes a WAV file (PCM 8/16/24/32-bit or IEEE float), mixing
// every channel down to mono.
func Read(path string) (*audio.Clip, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return Decode(data)
}

// IsWAV reports whether data starts with a RIFF/WAVE header.
func IsWAV(data []byte) bool {
	return len(data) >= 12 && string(data[0:4]) == "RIFF" && string(data[8:12]) == "WAVE"
}

// Decode decodes WAV bytes; see Read.
func Decode(data []byte) (*audio.Clip, error) {
	if !IsWAV(data) {
		return nil, errors.New("not a WAV file")
	}
	f, pcm, err := chunks(bytes.NewReader(data[12:]))
	if err != nil {
		return nil, err
	}
	bps := int(f.bits) / 8
	frame := bps * int(f.channels)
	if frame == 0 || f.rate == 0 {
		return nil, errors.New("bad WAV header")
	}
	n := len(pcm) / frame
	out := make([]float32, n)
	for i := 0; i < n; i++ {
		var sum float32
		for c := 0; c < int(f.channels); c++ {
			off := i*frame + c*bps
			v, err := sample(f, pcm[off:off+bps])
			if err != nil {
				return nil, err
			}
			sum += v
		}
		out[i] = sum / float32(f.channels)
	}
	return &audio.Clip{Samples: out, SampleRate: int(f.rate)}, nil
}

// chunks walks the RIFF chunks and returns the format and sample data.
func chunks(r *bytes.Reader) (format, []byte, error) {
	var f format
	var pcm []byte
	var err error
	haveFmt := false
	for {
		var h [8]byte
		if _, err := io.ReadFull(r, h[:]); err != nil {
			break
		}
		id := string(h[0:4])
		size := binary.LittleEndian.Uint32(h[4:8])
		// Streamed WAVs leave the data size at 0 or 0xFFFFFFFF: take
		// whatever is left.
		if rem := uint32(r.Len()); size > rem || (id == "data" && size == 0) {
			size = rem
		}
		body := make([]byte, size)
		if _, err := io.ReadFull(r, body); err != nil {
			return f, nil, err
		}
		if size%2 == 1 {
			r.ReadByte()
		}
		switch id {
		case "fmt ":
			if f, err = parseFmt(body); err != nil {
				return f, nil, err
			}
			haveFmt = true
		case "data":
			pcm = body
		}
	}
	if !haveFmt || pcm == nil {
		return f, nil, errors.New("WAV file has no audio")
	}
	return f, pcm, nil
}

// parseFmt reads a "fmt " chunk.
func parseFmt(body []byte) (format, error) {
	var f format
	if len(body) < 16 {
		return f, errors.New("bad fmt chunk")
	}
	f.tag = binary.LittleEndian.Uint16(body[0:2])
	f.channels = binary.LittleEndian.Uint16(body[2:4])
	f.rate = binary.LittleEndian.Uint32(body[4:8])
	f.bits = binary.LittleEndian.Uint16(body[14:16])
	if f.tag == 0xFFFE && len(body) >= 26 { // WAVE_FORMAT_EXTENSIBLE
		f.tag = binary.LittleEndian.Uint16(body[24:26])
	}
	return f, nil
}

func sample(f format, b []byte) (float32, error) {
	switch {
	case f.tag == 3 && f.bits == 32:
		return math.Float32frombits(binary.LittleEndian.Uint32(b)), nil
	case f.tag == 3 && f.bits == 64:
		return float32(math.Float64frombits(binary.LittleEndian.Uint64(b))), nil
	case f.tag == 1 && f.bits == 8:
		return (float32(b[0]) - 128) / 128, nil
	case f.tag == 1 && f.bits == 16:
		return float32(int16(binary.LittleEndian.Uint16(b))) / 32768, nil
	case f.tag == 1 && f.bits == 24:
		return float32(int32(b[0])|int32(b[1])<<8|int32(int8(b[2]))<<16) / 8388608, nil
	case f.tag == 1 && f.bits == 32:
		return float32(int32(binary.LittleEndian.Uint32(b))) / 2147483648, nil
	}
	return 0, fmt.Errorf("unsupported WAV encoding (format %d, %d-bit)", f.tag, f.bits)
}
