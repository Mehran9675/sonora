// Package sherpa is a minimal binding to sherpa-onnx's C API
// (https://github.com/k2-fsa/sherpa-onnx, Apache-2.0) for offline text
// to speech: VITS/Piper, Kitten and Kokoro models running on the CPU.
//
// The binding is compiled only with the "sherpa" build tag, and links
// against libsherpa-onnx-c-api.so from third_party/sherpa-onnx (see
// tools/fetchsherpa). Without the tag, Available is false and New
// fails, so normal desktop builds need no native library. c-api.h is
// the header of the release named in tools/fetchsherpa.
package sherpa

// Family selects the model architecture.
type Family string

const (
	VITS   Family = "vits" // Piper voices
	Kitten Family = "kitten"
	Kokoro Family = "kokoro"
)

// Config locates a model's files.
type Config struct {
	Family  Family
	Model   string // .onnx file
	Voices  string // voices.bin (Kitten, Kokoro)
	Tokens  string // tokens.txt
	DataDir string // espeak-ng-data
	Lexicon string // comma-separated lexicons (Kokoro)
	DictDir string // dict (Kokoro)
	Threads int
}

// Progress receives generation progress in [0, 1]; returning false
// stops generation.
type Progress func(fraction float32) bool
