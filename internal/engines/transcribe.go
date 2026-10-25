package engines

import "context"

// Task is what an engine does.
type Task string

const (
	// TaskSpeak turns text into speech (the default).
	TaskSpeak Task = ""
	// TaskTranscribe turns speech into text.
	TaskTranscribe Task = "transcribe"
)

// TranscribeRequest is one speech-to-text job.
type TranscribeRequest struct {
	// Audio is a 16 kHz mono 16-bit WAV file.
	Audio string `json:"audio"`
	// Language is an ISO 639-1 code, or empty to detect it.
	Language string `json:"language,omitempty"`
	// Translate asks for an English translation instead of a
	// transcript, on engines that can.
	Translate bool   `json:"translate,omitempty"`
	Options   Values `json:"options,omitempty"`
}

// Segment is a timed stretch of a transcript.
type Segment struct {
	Start float64 `json:"start"` // seconds
	End   float64 `json:"end"`
	Text  string  `json:"text"`
}

// Transcript is what a transcription returns.
type Transcript struct {
	Text     string    `json:"text"`
	Language string    `json:"language,omitempty"`
	Segments []Segment `json:"segments,omitempty"`
}

// Transcriber is implemented by engines whose Task is TaskTranscribe.
type Transcriber interface {
	Transcribe(ctx context.Context, req TranscribeRequest, log Logger) (*Transcript, error)
}

// Languages are the languages offered for transcription, by ISO code;
// an engine may support only some of them.
var Languages = []Choice{
	{Value: "", Label: "Detect automatically"},
	{Value: "en", Label: "English"}, {Value: "zh", Label: "Chinese"}, {Value: "es", Label: "Spanish"},
	{Value: "fr", Label: "French"}, {Value: "de", Label: "German"}, {Value: "it", Label: "Italian"},
	{Value: "pt", Label: "Portuguese"}, {Value: "nl", Label: "Dutch"}, {Value: "ru", Label: "Russian"},
	{Value: "uk", Label: "Ukrainian"}, {Value: "pl", Label: "Polish"}, {Value: "cs", Label: "Czech"},
	{Value: "sv", Label: "Swedish"}, {Value: "da", Label: "Danish"}, {Value: "fi", Label: "Finnish"},
	{Value: "no", Label: "Norwegian"}, {Value: "el", Label: "Greek"}, {Value: "hu", Label: "Hungarian"},
	{Value: "ro", Label: "Romanian"}, {Value: "tr", Label: "Turkish"}, {Value: "ar", Label: "Arabic"},
	{Value: "fa", Label: "Persian"}, {Value: "he", Label: "Hebrew"}, {Value: "hi", Label: "Hindi"},
	{Value: "bn", Label: "Bengali"}, {Value: "ur", Label: "Urdu"}, {Value: "ja", Label: "Japanese"},
	{Value: "ko", Label: "Korean"}, {Value: "vi", Label: "Vietnamese"}, {Value: "th", Label: "Thai"},
	{Value: "id", Label: "Indonesian"}, {Value: "ms", Label: "Malay"}, {Value: "yue", Label: "Cantonese"},
}
