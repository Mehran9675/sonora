package onnx

import (
	"slices"
	"strings"

	"github.com/mehran9675/sonora/internal/engines"
	"github.com/mehran9675/sonora/internal/engines/onnx/sherpa"
)

const whisperTurbo = "sherpa-onnx-whisper-turbo"

// The speech recognition packages. Sizes are downloads; Whisper
// packages carry fp32 and int8 models, and only int8 is unpacked.
var (
	whisperModels = []model{
		{ID: "sherpa-onnx-whisper-tiny", Label: "Tiny · 39M", Size: "110 MB"},
		{ID: "sherpa-onnx-whisper-base", Label: "Base · 74M", Size: "197 MB"},
		{ID: "sherpa-onnx-whisper-small", Label: "Small · 244M", Size: "609 MB"},
		{ID: whisperTurbo, Label: "Large v3 Turbo · 809M", Size: "537 MB"},
	}
	parakeetModel = model{ID: "sherpa-onnx-nemo-parakeet-tdt-0.6b-v3-int8", Label: "Parakeet TDT 0.6B v3", Size: "464 MB",
		Family: sherpa.Transducer}
	senseVoiceModel = model{ID: "sherpa-onnx-sense-voice-zh-en-ja-ko-yue-int8-2025-09-09", Label: "SenseVoice Small", Size: "158 MB",
		Family: sherpa.SenseVoice}
)

func init() {
	for i := range whisperModels {
		whisperModels[i].Family, whisperModels[i].Int8 = sherpa.Whisper, true
	}
	all := []*model{&parakeetModel, &senseVoiceModel}
	for i := range whisperModels {
		all = append(all, &whisperModels[i])
	}
	for _, m := range all {
		m.Release, m.Extra = "asr-models", []string{sileroVAD}
	}
}

func whisperManifest() *engines.Manifest {
	return &engines.Manifest{
		ID: "ondevice-whisper", Name: "Whisper", Vendor: "OpenAI · on device",
		Description: "Transcribes speech in about 100 languages, offline on this device, and can translate it into English. " +
			"Bigger models are more accurate and slower; Turbo suits flagship phones.",
		Tags:    []string{"Offline", "99 languages", "Translate"},
		License: "MIT", Homepage: "https://github.com/openai/whisper",
		Size: "110–609 MB", Accent: []string{"#10b981", "#0d9488"},
		Kind: engines.KindOnDevice, Task: engines.TaskTranscribe, Translate: true,
		Options: []engines.Option{modelOption(whisperModels, "A new model downloads the first time you use it.")},
	}
}

func parakeetManifest() *engines.Manifest {
	return &engines.Manifest{
		ID: "ondevice-parakeet", Name: "Parakeet", Vendor: "NVIDIA · on device",
		Description: "NVIDIA's fast, accurate recognizer for English and 24 other European languages, with punctuation, " +
			"offline on this device. Needs about 1 GB of free memory.",
		Tags:    []string{"Offline", "25 languages", "Accurate"},
		License: "CC-BY-4.0", Homepage: "https://huggingface.co/nvidia/parakeet-tdt-0.6b-v3",
		Size: parakeetModel.Size, Accent: []string{"#76b900", "#16a34a"},
		// Parakeet v3 detects which of its 25 languages is spoken and
		// can't be told.
		Kind: engines.KindOnDevice, Task: engines.TaskTranscribe, Languages: []string{""},
	}
}

func senseVoiceManifest() *engines.Manifest {
	return &engines.Manifest{
		ID: "ondevice-sensevoice", Name: "SenseVoice", Vendor: "FunAudioLLM · on device",
		Description: "Very fast recognition of Chinese, Cantonese, English, Japanese and Korean, offline on this device.",
		Tags:        []string{"Offline", "Fast", "Chinese · English · Japanese · Korean"},
		License:     "FunASR model license", Homepage: "https://github.com/FunAudioLLM/SenseVoice",
		Size: senseVoiceModel.Size, Accent: []string{"#f59e0b", "#ef4444"},
		Kind: engines.KindOnDevice, Task: engines.TaskTranscribe, Languages: append([]string{""}, senseVoiceLanguages...),
	}
}

// whisperLanguages are the languages Whisper models know; sherpa-onnx
// ends the process when given any other. Cantonese ("yue") came with
// Large v3, so only Turbo knows it.
var whisperLanguages = strings.Fields(`en zh de es ru ko fr ja pt tr pl ca nl ar sv it id hi fi vi he uk el ms cs ro
	da hu ta no th ur hr bg lt la mi ml cy sk te fa lv bn sr az sl kn et mk br eu is hy ne mn bs kk sq sw gl mr pa si
	km sn yo so af oc ka be tg sd gu am yi lo uz fo ht ps tk nn mt sa lb my bo tl mg as tt haw ln ha ba jw su`)

// senseVoiceLanguages are the language hints SenseVoice accepts.
var senseVoiceLanguages = []string{"zh", "en", "yue", "ja", "ko"}

// supportedLanguage returns lang if model m accepts it as a hint, or ""
// to have the language detected.
func supportedLanguage(m model, lang string) string {
	var known []string
	switch m.Family {
	case sherpa.Whisper:
		known = whisperLanguages
		if m.ID == whisperTurbo {
			known = append(known, "yue")
		}
	case sherpa.SenseVoice:
		known = senseVoiceLanguages
	}
	if slices.Contains(known, lang) {
		return lang
	}
	return ""
}
