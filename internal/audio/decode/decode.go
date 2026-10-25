// Package decode reads audio files - WAV, MP3, FLAC and Ogg Vorbis
// natively, anything else through ffmpeg where it is installed - as
// mono samples at a chosen rate. It decodes a block at a time and
// resamples as it goes, so an hour-long recording costs memory only at
// the target rate.
package decode

import (
	"bufio"
	"bytes"
	"errors"
	"io"
	"os"
	"path/filepath"

	"github.com/hajimehoshi/go-mp3"
	"github.com/jfreymuth/oggvorbis"
	"github.com/mewkiz/flac"

	"github.com/mehran9675/sonora/internal/audio"
	"github.com/mehran9675/sonora/internal/audio/convert"
	"github.com/mehran9675/sonora/internal/audio/wav"
)

// SpeechRate is the sample rate speech recognizers take.
const SpeechRate = 16000

// Extensions are the audio files worth offering in a file picker.
var Extensions = []string{"wav", "mp3", "flac", "ogg", "oga", "m4a", "aac", "opus", "webm", "mp4", "aiff", "wma"}

// ErrFormat means the file is in a format only ffmpeg can read and
// ffmpeg isn't installed (or, on a phone, can't be).
var ErrFormat = errors.New("this audio format can't be read here; use WAV, MP3, FLAC or Ogg Vorbis, or install ffmpeg")

// File decodes the audio file at path as mono samples at rate.
func File(path string, rate int) (*audio.Clip, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	r := bufio.NewReaderSize(f, 64<<10)
	head, _ := r.Peek(12)
	out := newResampler(rate)
	switch {
	case wav.IsWAV(head):
		err = wav.Stream(r, out.write)
	case bytes.HasPrefix(head, []byte("fLaC")):
		err = decodeFLAC(r, out)
	case bytes.HasPrefix(head, []byte("OggS")):
		err = decodeVorbis(r, out) // fails on Ogg Opus
	case bytes.HasPrefix(head, []byte("ID3")), len(head) >= 2 && head[0] == 0xFF && head[1]&0xE0 == 0xE0:
		err = decodeMP3(r, out) // fails on ADTS AAC, which looks alike
	default:
		return viaFFmpeg(path, rate)
	}
	if err != nil && out.empty() {
		if clip, ferr := viaFFmpeg(path, rate); ferr == nil {
			return clip, nil
		}
	}
	if err != nil {
		return nil, err
	}
	clip := out.clip()
	if len(clip.Samples) == 0 {
		return nil, errors.New("the file has no audio")
	}
	return clip, nil
}

// viaFFmpeg converts the file to WAV with ffmpeg and decodes that.
func viaFFmpeg(path string, rate int) (*audio.Clip, error) {
	if !convert.Available() {
		return nil, ErrFormat
	}
	tmp, err := os.MkdirTemp("", "sonora-decode-")
	if err != nil {
		return nil, err
	}
	defer os.RemoveAll(tmp)
	out := filepath.Join(tmp, "audio.wav")
	if err := convert.FileToWAV(path, out); err != nil {
		return nil, err
	}
	f, err := os.Open(out)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	res := newResampler(rate)
	if err := wav.Stream(bufio.NewReader(f), res.write); err != nil {
		return nil, err
	}
	return res.clip(), nil
}

func decodeMP3(r io.Reader, out *resampler) error {
	d, err := mp3.NewDecoder(r)
	if err != nil {
		return err
	}
	// go-mp3 always produces 16-bit little-endian stereo.
	buf := make([]byte, 4*4096)
	mono := make([]float32, 4096)
	for {
		n, err := io.ReadFull(d, buf)
		frames := n / 4
		for i := 0; i < frames; i++ {
			l := int16(uint16(buf[4*i]) | uint16(buf[4*i+1])<<8)
			r := int16(uint16(buf[4*i+2]) | uint16(buf[4*i+3])<<8)
			mono[i] = (float32(l) + float32(r)) / 65536
		}
		if frames > 0 {
			out.write(d.SampleRate(), mono[:frames])
		}
		if err == io.EOF || err == io.ErrUnexpectedEOF {
			return nil
		}
		if err != nil {
			return err
		}
	}
}

func decodeFLAC(r io.Reader, out *resampler) error {
	s, err := flac.New(r)
	if err != nil {
		return err
	}
	rate := int(s.Info.SampleRate)
	scale := float32(int64(1) << (s.Info.BitsPerSample - 1))
	var mono []float32
	for {
		fr, err := s.ParseNext()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}
		n := fr.Subframes[0].NSamples
		if cap(mono) < n {
			mono = make([]float32, n)
		}
		mono = mono[:n]
		clear(mono)
		for _, sub := range fr.Subframes {
			for i, v := range sub.Samples[:n] {
				mono[i] += float32(v)
			}
		}
		k := scale * float32(len(fr.Subframes))
		for i := range mono {
			mono[i] /= k
		}
		out.write(rate, mono)
	}
}

func decodeVorbis(r io.Reader, out *resampler) error {
	d, err := oggvorbis.NewReader(r)
	if err != nil {
		return err
	}
	ch := d.Channels()
	buf := make([]float32, ch*4096)
	mono := make([]float32, 4096)
	for {
		n, err := d.Read(buf)
		frames := n / ch
		for i := 0; i < frames; i++ {
			var sum float32
			for c := 0; c < ch; c++ {
				sum += buf[i*ch+c]
			}
			mono[i] = sum / float32(ch)
		}
		if frames > 0 {
			out.write(d.SampleRate(), mono[:frames])
		}
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}
	}
}
