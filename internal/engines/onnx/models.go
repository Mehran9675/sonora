// Package onnx runs text-to-speech models inside the app with
// sherpa-onnx: downloaded Kitten, Kokoro, Piper and ZipVoice models on
// the CPU, with no Python. It's what Android uses; desktop builds include it
// when built with the "sherpa" tag.
package onnx

import "github.com/mehran9675/sonora/internal/engines/onnx/sherpa"

// release is where sherpa-onnx publishes its ready-to-run model packages.
const release = "https://github.com/k2-fsa/sherpa-onnx/releases/download/tts-models/"

// vocos is the vocoder ZipVoice models need, a separate download.
const vocos = "https://github.com/k2-fsa/sherpa-onnx/releases/download/vocoder-models/vocos_24khz.onnx"

// model is one downloadable package.
type model struct {
	ID     string // package name; the archive is release+ID+".tar.bz2"
	Label  string
	Size   string
	Family sherpa.Family
	// Lexicon lists lexicon files inside the package (Kokoro).
	Lexicon string
	// Dict is the dictionary folder inside the package (Kokoro).
	Dict string
	// Steps is the number of flow-matching steps (ZipVoice).
	Steps int
}

func (m model) url() string { return release + m.ID + ".tar.bz2" }

var (
	kittenModels = []model{
		{ID: "kitten-nano-en-v0_8-int8", Label: "Nano · 15M", Size: "31 MB", Family: sherpa.Kitten},
		{ID: "kitten-micro-en-v0_8", Label: "Micro · 40M", Size: "44 MB", Family: sherpa.Kitten},
		{ID: "kitten-mini-en-v0_8", Label: "Mini · 80M", Size: "68 MB", Family: sherpa.Kitten},
	}
	// kokoroModels share one voice list; the full-precision model
	// sounds a little cleaner and needs about 1 GB of free memory.
	kokoroModels = []model{
		{ID: "kokoro-int8-multi-lang-v1_0", Label: "Compact · int8", Size: "132 MB",
			Family: sherpa.Kokoro, Lexicon: "lexicon-us-en.txt", Dict: "dict"},
		{ID: "kokoro-multi-lang-v1_0", Label: "Full quality · fp32", Size: "333 MB",
			Family: sherpa.Kokoro, Lexicon: "lexicon-us-en.txt", Dict: "dict"},
	}
	// zipvoiceModels clone a voice from a recording and its transcript.
	// Both are the distilled model, which needs only 4 steps; sizes
	// include the shared 52 MB vocoder.
	zipvoiceModels = []model{
		{ID: "sherpa-onnx-zipvoice-distill-int8-zh-en-emilia", Label: "Compact · int8", Size: "156 MB",
			Family: sherpa.ZipVoice, Lexicon: "lexicon.txt", Steps: 4},
		{ID: "sherpa-onnx-zipvoice-distill-fp32-zh-en-emilia", Label: "Full quality · fp32", Size: "508 MB",
			Family: sherpa.ZipVoice, Lexicon: "lexicon.txt", Steps: 4},
	}
)

// piperVoices are the Piper voices sherpa-onnx packages, one download
// each.
var piperVoices = []struct {
	ID, Name, Language, Quality string
}{
	{"en_US-lessac-medium", "Lessac", "English (US)", "Medium"},
	{"en_US-lessac-high", "Lessac", "English (US)", "High"},
	{"en_US-amy-low", "Amy", "English (US)", "Low"},
	{"en_US-ryan-medium", "Ryan", "English (US)", "Medium"},
	{"en_US-ryan-high", "Ryan", "English (US)", "High"},
	{"en_US-ljspeech-high", "LJSpeech", "English (US)", "High"},
	{"en_GB-alba-medium", "Alba", "English (UK)", "Medium"},
	{"en_GB-cori-high", "Cori", "English (UK)", "High"},
	{"en_GB-northern_english_male-medium", "Northern English Male", "English (UK)", "Medium"},
	{"de_DE-thorsten-medium", "Thorsten", "German", "Medium"},
	{"de_DE-thorsten-high", "Thorsten", "German", "High"},
	{"fr_FR-siwis-medium", "Siwis", "French", "Medium"},
	{"es_ES-davefx-medium", "Davefx", "Spanish", "Medium"},
	{"it_IT-paola-medium", "Paola", "Italian", "Medium"},
	{"pt_BR-faber-medium", "Faber", "Portuguese (Brazil)", "Medium"},
	{"pl_PL-gosia-medium", "Gosia", "Polish", "Medium"},
	{"ru_RU-irina-medium", "Irina", "Russian", "Medium"},
	{"uk_UA-ukrainian_tts-medium", "Ukrainian TTS", "Ukrainian", "Medium"},
	{"tr_TR-dfki-medium", "DFKI", "Turkish", "Medium"},
	{"ar_JO-kareem-medium", "Kareem", "Arabic", "Medium"},
	{"fa_IR-amir-medium", "Amir", "Persian", "Medium"},
	{"hi_IN-pratham-medium", "Pratham", "Hindi", "Medium"},
	{"zh_CN-huayan-medium", "Huayan", "Chinese", "Medium"},
}

// piperSize is the download size of a voice by quality.
func piperSize(quality string) string {
	if quality == "High" {
		return "110 MB"
	}
	return "67 MB"
}

func piperModel(voice string) model {
	quality := "Medium"
	for _, v := range piperVoices {
		if v.ID == voice {
			quality = v.Quality
		}
	}
	return model{ID: "vits-piper-" + voice, Label: voice, Size: piperSize(quality), Family: sherpa.VITS}
}

// kittenVoices are speaker ids 0-7 in sherpa-onnx's voices.bin
// (expr-voice-2-m, -2-f, -3-m, -3-f, …), with KittenML's names.
var kittenVoices = []struct{ Name, Gender string }{
	{"Jasper", "Male"}, {"Bella", "Female"}, {"Bruno", "Male"}, {"Luna", "Female"},
	{"Hugo", "Male"}, {"Rosie", "Female"}, {"Leo", "Male"}, {"Kiki", "Female"},
}

// kokoroVoices are Kokoro v1.0's English speakers; their position is
// the speaker id in sherpa-onnx's voices.bin.
var kokoroVoices = []string{
	"af_alloy", "af_aoede", "af_bella", "af_heart", "af_jessica", "af_kore", "af_nicole", "af_nova",
	"af_river", "af_sarah", "af_sky", "am_adam", "am_echo", "am_eric", "am_fenrir", "am_liam",
	"am_michael", "am_onyx", "am_puck", "am_santa", "bf_alice", "bf_emma", "bf_isabella", "bf_lily",
	"bm_daniel", "bm_fable", "bm_george", "bm_lewis",
}
