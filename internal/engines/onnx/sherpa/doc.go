// Package sherpa is a minimal binding to sherpa-onnx's C API
// (https://github.com/k2-fsa/sherpa-onnx, Apache-2.0) for offline text
// to speech: VITS/Piper, Kitten, Kokoro and ZipVoice models running on
// the CPU.
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
	// ZipVoice clones the voice of a reference recording.
	ZipVoice Family = "zipvoice"
)

// Config locates a model's files.
type Config struct {
	Family  Family
	Model   string // .onnx file
	Voices  string // voices.bin (Kitten, Kokoro)
	Tokens  string // tokens.txt
	DataDir string // espeak-ng-data
	Lexicon string // comma-separated lexicons (Kokoro, ZipVoice)
	DictDir string // dict (Kokoro)
	// Encoder, Decoder and Vocoder are ZipVoice's three models.
	Encoder, Decoder, Vocoder string
	Threads                   int
}

// Params are per-request settings.
type Params struct {
	Speaker int // speaker id of multi-speaker models
	Speed   float32
	// Reference is the recording a cloning model copies (mono samples at
	// ReferenceRate), and ReferenceText what is said in it.
	Reference     []float32
	ReferenceRate int
	ReferenceText string
	// Steps is the number of flow-matching steps (ZipVoice); 0 uses the
	// model's default.
	Steps int
}

// Progress receives generation progress in [0, 1]; returning false
// stops generation.
type Progress func(fraction float32) bool
