package playback

import (
	"encoding/binary"
	"io"
	"math"
	"sync"
)

// source streams mono samples as interleaved stereo float32 bytes and
// supports seeking, which oto's Player forwards to it.
type source struct {
	mu      sync.Mutex
	samples []float32
	pos     int64 // in bytes
}

func (s *source) total() int64 { return int64(len(s.samples)) * frameBytes }

func (s *source) Read(p []byte) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.pos >= s.total() {
		return 0, io.EOF
	}
	n := 0
	for n+frameBytes <= len(p) && s.pos < s.total() {
		v := math.Float32bits(s.samples[s.pos/frameBytes])
		binary.LittleEndian.PutUint32(p[n:], v)
		binary.LittleEndian.PutUint32(p[n+4:], v)
		n += frameBytes
		s.pos += frameBytes
	}
	return n, nil
}

func (s *source) Seek(off int64, whence int) (int64, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	switch whence {
	case io.SeekCurrent:
		off += s.pos
	case io.SeekEnd:
		off += s.total()
	}
	s.pos = max(0, min(s.total(), off/frameBytes*frameBytes))
	return s.pos, nil
}

func (s *source) position() int64 {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.pos
}
