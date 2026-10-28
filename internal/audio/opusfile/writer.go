// Package opusfile writes and reads Ogg Opus files (RFC 7845), the
// format Sonora records in: speech at 16-64 kbps takes a few percent
// of a WAV's size. The codec is gopus, a pure-Go Opus implementation
// checked against libopus, so no C library is needed on any platform.
package opusfile

import (
	"bufio"
	"encoding/binary"
	"errors"
	"io"
	"math/rand/v2"
	"os"
	"time"

	"github.com/thesyncim/gopus"
	"github.com/thesyncim/gopus/container/ogg"
)

// frameMS is the length of each Opus frame.
const frameMS = 20

// packetsPerPage groups about a second of audio into each Ogg page, so
// the container costs little and an interrupted recording loses at most
// that much.
const packetsPerPage = 50

// Writer encodes mono audio into an Ogg Opus file as it arrives.
type Writer struct {
	f    *os.File
	w    *bufio.Writer
	enc  *gopus.Encoder
	rate int // input sample rate

	frame   []float32 // samples waiting for a full frame
	packet  []byte
	preSkip int

	serial  uint32
	seq     uint32
	granule uint64 // 48 kHz samples encoded so far, pre-skip included
	written uint64 // input samples written, for the final granule

	pageData []byte
	lacing   []byte
	packets  int
}

// Create starts a mono Ogg Opus file at path. rate is the input sample
// rate (8000, 12000, 16000, 24000 or 48000) and bitrate the target in
// bits per second.
func Create(path string, rate, bitrate int) (*Writer, error) {
	// The audio application keeps the full band, the better choice for
	// voices that will be cloned.
	enc, err := gopus.NewEncoder(gopus.EncoderConfig{SampleRate: rate, Channels: 1, Application: gopus.ApplicationAudio})
	if err != nil {
		return nil, err
	}
	if err := enc.SetBitrate(bitrate); err != nil {
		return nil, err
	}
	if err := enc.SetFrameSize(rate * frameMS / 1000); err != nil {
		return nil, err
	}
	preSkip := enc.Lookahead() * 48000 / rate
	f, err := os.Create(path)
	if err != nil {
		return nil, err
	}
	w := &Writer{f: f, w: bufio.NewWriterSize(f, 64<<10), enc: enc, rate: rate, preSkip: preSkip,
		serial: rand.Uint32(), packet: make([]byte, 4000)}
	head := opusHead(preSkip, rate)
	tags := opusTags("Sonora")
	if err := w.writePage([][]byte{head}, 0, 0x02); err != nil { // beginning of stream
		w.abort()
		return nil, err
	}
	if err := w.writePage([][]byte{tags}, 0, 0); err != nil {
		w.abort()
		return nil, err
	}
	w.granule = uint64(preSkip)
	return w, nil
}

// Write adds samples in [-1, 1] at the writer's rate.
func (w *Writer) Write(samples []float32) error {
	size := w.rate * frameMS / 1000
	for _, s := range samples {
		w.frame = append(w.frame, max(-1, min(1, s)))
		if len(w.frame) == size {
			if err := w.encode(); err != nil {
				return err
			}
		}
	}
	w.written += uint64(len(samples))
	return nil
}

func (w *Writer) encode() error {
	n, err := w.enc.Encode(w.frame, w.packet)
	if err != nil {
		return err
	}
	w.frame = w.frame[:0]
	w.granule += uint64(48000 * frameMS / 1000)
	w.addPacket(w.packet[:n])
	if w.packets >= packetsPerPage || len(w.lacing) > 200 {
		return w.flushPage(0)
	}
	return nil
}

// addPacket queues a packet for the current page.
func (w *Writer) addPacket(p []byte) {
	for n := len(p); ; n -= 255 {
		if n < 255 {
			w.lacing = append(w.lacing, byte(n))
			break
		}
		w.lacing = append(w.lacing, 255)
	}
	w.pageData = append(w.pageData, p...)
	w.packets++
}

func (w *Writer) flushPage(flags byte) error {
	if w.packets == 0 && flags == 0 {
		return nil
	}
	err := w.writeRaw(w.lacing, w.pageData, w.granule, flags)
	w.lacing, w.pageData, w.packets = w.lacing[:0], w.pageData[:0], 0
	if err != nil {
		return err
	}
	// Push each page to the file, so a crash loses at most a second.
	return w.w.Flush()
}

// writePage writes a page holding whole packets.
func (w *Writer) writePage(packets [][]byte, granule uint64, flags byte) error {
	var lacing, data []byte
	for _, p := range packets {
		for n := len(p); ; n -= 255 {
			if n < 255 {
				lacing = append(lacing, byte(n))
				break
			}
			lacing = append(lacing, 255)
		}
		data = append(data, p...)
	}
	return w.writeRaw(lacing, data, granule, flags)
}

func (w *Writer) writeRaw(lacing, data []byte, granule uint64, flags byte) error {
	h := make([]byte, 27, 27+len(lacing))
	copy(h, "OggS")
	h[5] = flags
	binary.LittleEndian.PutUint64(h[6:], granule)
	binary.LittleEndian.PutUint32(h[14:], w.serial)
	binary.LittleEndian.PutUint32(h[18:], w.seq)
	h[26] = byte(len(lacing))
	h = append(h, lacing...)
	crc := crcUpdate(crcUpdate(0, h), data)
	binary.LittleEndian.PutUint32(h[22:], crc)
	w.seq++
	if _, err := w.w.Write(h); err != nil {
		return err
	}
	_, err := w.w.Write(data)
	return err
}

// Samples returns how much audio has been written, in input samples.
func (w *Writer) Samples() uint64 { return w.written }

// Close encodes what's left (padding the last frame with silence),
// marks the end of the stream and closes the file. The last page's
// granule position trims the padding.
func (w *Writer) Close() error {
	if len(w.frame) > 0 {
		size := w.rate * frameMS / 1000
		for len(w.frame) < size {
			w.frame = append(w.frame, 0)
		}
		if err := w.encode(); err != nil {
			w.abort()
			return err
		}
	}
	// Exact end: pre-skip plus the real samples, at 48 kHz.
	w.granule = uint64(w.preSkip) + w.written*48000/uint64(w.rate)
	if w.packets == 0 {
		// An empty end-of-stream page still carries the final position.
		w.lacing = w.lacing[:0]
	}
	if err := w.flushPage(0x04); err != nil { // end of stream
		w.abort()
		return err
	}
	return w.f.Close()
}

func (w *Writer) abort() {
	w.f.Close()
	os.Remove(w.f.Name())
}

// ErrNotOpus means the file isn't Ogg Opus.
var ErrNotOpus = errors.New("not an Ogg Opus file")

// IsOpus reports whether head, the start of a file, is Ogg Opus.
func IsOpus(head []byte) bool {
	return len(head) >= 36 && string(head[:4]) == "OggS" && string(head[28:36]) == "OpusHead"
}

// Decode reads an Ogg Opus stream, calling fn with mono 48 kHz samples
// a packet at a time.
func Decode(r io.Reader, fn func(rate int, mono []float32) error) error {
	or, err := ogg.NewReader(r)
	if err != nil {
		return ErrNotOpus
	}
	ch := max(1, int(or.Channels()))
	if ch > 2 {
		return errors.New("Opus files with more than two channels aren't supported")
	}
	dec, err := gopus.NewDecoder(gopus.DefaultDecoderConfig(48000, ch))
	if err != nil {
		return err
	}
	skip := int(or.PreSkip())
	pcm := make([]float32, 5760*ch) // the longest Opus packet: 120 ms
	mono := make([]float32, 5760)
	for {
		p, _, err := or.ReadPacket()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}
		n, err := dec.Decode(p, pcm)
		if err != nil {
			return err
		}
		for i := 0; i < n; i++ {
			var sum float32
			for c := 0; c < ch; c++ {
				sum += pcm[i*ch+c]
			}
			mono[i] = sum / float32(ch)
		}
		out := mono[:n]
		if skip > 0 {
			d := min(skip, len(out))
			out, skip = out[d:], skip-d
		}
		if len(out) > 0 {
			if err := fn(48000, out); err != nil {
				return err
			}
		}
	}
}

// opusHead builds the identification header (RFC 7845 §5.1) for mono.
func opusHead(preSkip, rate int) []byte {
	b := make([]byte, 19)
	copy(b, "OpusHead")
	b[8] = 1 // version
	b[9] = 1 // channels
	binary.LittleEndian.PutUint16(b[10:], uint16(preSkip))
	binary.LittleEndian.PutUint32(b[12:], uint32(rate))
	return b // output gain 0, mapping family 0
}

// opusTags builds the comment header (RFC 7845 §5.2) with no comments.
func opusTags(vendor string) []byte {
	b := make([]byte, 8+4+len(vendor)+4)
	copy(b, "OpusTags")
	binary.LittleEndian.PutUint32(b[8:], uint32(len(vendor)))
	copy(b[12:], vendor)
	return b
}

// Duration reads an Ogg Opus file's length from its header and last
// page, without decoding it.
func Duration(path string) (time.Duration, error) {
	f, err := os.Open(path)
	if err != nil {
		return 0, err
	}
	defer f.Close()
	head := make([]byte, 64)
	if n, _ := io.ReadFull(f, head); !IsOpus(head[:n]) {
		return 0, ErrNotOpus
	}
	preSkip := int64(binary.LittleEndian.Uint16(head[28+10:]))
	st, err := f.Stat()
	if err != nil {
		return 0, err
	}
	tail := min(st.Size(), 64<<10)
	buf := make([]byte, tail)
	if _, err := f.ReadAt(buf, st.Size()-tail); err != nil && err != io.EOF {
		return 0, err
	}
	for i := len(buf) - 27; i >= 0; i-- {
		if string(buf[i:i+4]) == "OggS" && buf[i+4] == 0 {
			granule := int64(binary.LittleEndian.Uint64(buf[i+6:]))
			if granule < 0 { // -1: no packet ends on this page
				continue
			}
			return time.Duration(max(0, granule-preSkip)) * time.Second / 48000, nil
		}
	}
	return 0, errors.New("no Ogg pages found")
}
