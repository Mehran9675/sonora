package onnx

import (
	"strconv"
	"strings"

	"github.com/mehran9675/sonora/internal/engines"
)

func kittenManifest() *engines.Manifest {
	var choices []engines.Choice
	for _, m := range kittenModels {
		choices = append(choices, engines.Choice{Value: m.ID, Label: m.Label})
	}
	var voices []engines.Voice
	for i, v := range kittenVoices {
		voices = append(voices, engines.Voice{ID: strconv.Itoa(i), Name: v.Name, Group: "English", Description: v.Gender})
	}
	return &engines.Manifest{
		ID: "ondevice-kitten", Name: "Kitten TTS Lite", Vendor: "KittenML · on device",
		Description: "Tiny expressive English voices that run right on this device, offline. Pick a bigger model for more natural speech.",
		Tags:        []string{"Offline", "Tiny", "English"},
		License:     "Apache-2.0", Homepage: "https://github.com/KittenML/KittenTTS",
		Size: "31–68 MB", Accent: []string{"#facc15", "#f97316"},
		Kind: engines.KindOnDevice, Speed: true, Voices: voices, DefaultVoice: "1",
		Options: []engines.Option{{ID: "model", Label: "Model", Type: engines.OptionSelect, Scope: engines.ScopeLoad,
			Default: kittenModels[0].ID, Choices: choices, Help: "A new model downloads the first time you use it."}},
	}
}

func kokoroManifest() *engines.Manifest {
	var voices []engines.Voice
	for i, id := range kokoroVoices {
		group := "American English"
		if id[0] == 'b' {
			group = "British English"
		}
		gender := "Female"
		if id[1] == 'm' {
			gender = "Male"
		}
		name := strings.ToUpper(id[3:4]) + id[4:]
		voices = append(voices, engines.Voice{ID: strconv.Itoa(i), Name: name, Group: group, Description: gender})
	}
	return &engines.Manifest{
		ID: "ondevice-kokoro", Name: "Kokoro", Vendor: "hexgrad · on device",
		Description: "Natural, clear American and British English voices from the acclaimed 82M Kokoro model, running offline on this device.",
		Tags:        []string{"Offline", "28 voices", "English"},
		License:     "Apache-2.0", Homepage: "https://github.com/hexgrad/kokoro",
		Size: kokoroModel.Size, Accent: []string{"#34d399", "#0ea5e9"},
		Kind: engines.KindOnDevice, Speed: true, Voices: voices, DefaultVoice: "3", // af_heart
	}
}

func piperManifest() *engines.Manifest {
	var voices []engines.Voice
	for _, v := range piperVoices {
		voices = append(voices, engines.Voice{ID: v.ID, Name: v.Name, Group: v.Language, Description: v.Quality + " quality · 67 MB"})
	}
	return &engines.Manifest{
		ID: "ondevice-piper", Name: "Piper", Vendor: "Rhasspy / OHF · on device",
		Description: "Fast neural voices in many languages that run offline on this device. Each voice downloads the first time you use it.",
		Tags:        []string{"Offline", "18 languages", "Fast"},
		License:     "GPL-3.0 · voices vary", Homepage: "https://github.com/OHF-Voice/piper1-gpl",
		Size: "67 MB per voice", Accent: []string{"#a3e635", "#10b981"},
		Kind: engines.KindOnDevice, Speed: true, Voices: voices, DefaultVoice: piperVoices[0].ID,
	}
}
