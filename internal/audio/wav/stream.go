package wav

import (
	"encoding/binary"
	"errors"
	"io"
)

// Stream decodes WAV audio from r a block at a time, without holding
// the file in memory, calling fn with mono samples at rate.
func Stream(r io.Reader, fn func(rate int, mono []float32) error) error {
	var head [12]byte
	if _, err := io.ReadFull(r, head[:]); err != nil || !IsWAV(head[:]) {
		return errors.New("not a WAV file")
	}
	var f format
	haveFmt := false
	for {
		var h [8]byte
		if _, err := io.ReadFull(r, h[:]); err != nil {
			if haveFmt {
				return nil
			}
			return errors.New("WAV file has no audio")
		}
		id := string(h[0:4])
		size := int64(binary.LittleEndian.Uint32(h[4:8]))
		switch id {
		case "fmt ":
			body := make([]byte, size)
			if _, err := io.ReadFull(r, body); err != nil {
				return err
			}
			var err error
			if f, err = parseFmt(body); err != nil {
				return err
			}
			haveFmt = true
		case "data":
			if !haveFmt {
				return errors.New("bad WAV header")
			}
			// Streamed WAVs leave the size at 0 or 0xFFFFFFFF: read to
			// the end.
			if size == 0 || size == 0xFFFFFFFF {
				size = -1
			}
			return streamData(r, f, size, fn)
		default:
			if _, err := io.CopyN(io.Discard, r, size); err != nil {
				return err
			}
		}
		if size%2 == 1 {
			io.CopyN(io.Discard, r, 1)
		}
	}
}

// streamData decodes size bytes of samples (to the end when size < 0).
func streamData(r io.Reader, f format, size int64, fn func(int, []float32) error) error {
	bps := int(f.bits) / 8
	frame := bps * int(f.channels)
	if frame == 0 || f.rate == 0 {
		return errors.New("bad WAV header")
	}
	if size >= 0 {
		r = io.LimitReader(r, size)
	}
	buf := make([]byte, frame*4096)
	mono := make([]float32, 4096)
	for {
		n, err := io.ReadFull(r, buf)
		frames := n / frame
		for i := 0; i < frames; i++ {
			var sum float32
			for c := 0; c < int(f.channels); c++ {
				off := i*frame + c*bps
				v, err := sample(f, buf[off:off+bps])
				if err != nil {
					return err
				}
				sum += v
			}
			mono[i] = sum / float32(f.channels)
		}
		if frames > 0 {
			if err := fn(int(f.rate), mono[:frames]); err != nil {
				return err
			}
		}
		if err == io.EOF || err == io.ErrUnexpectedEOF {
			return nil
		}
		if err != nil {
			return err
		}
	}
}
