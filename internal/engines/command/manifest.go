// Package command runs any text-to-speech program installed on the
// machine through a command template.
package command

import "github.com/mehran9675/sonora/internal/engines"

// Template returns the manifest for a new command-line engine.
func Template(id, name string) *engines.Manifest {
	return &engines.Manifest{
		ID:          id,
		Name:        name,
		Vendor:      "Command line",
		Description: "Runs a text-to-speech program installed on this computer.",
		Tags:        []string{"Command", "Local"},
		Accent:      []string{"#64748b", "#334155"},
		Kind:        engines.KindCommand,
		Speed:       true,
		Custom:      true,
		Options: []engines.Option{
			{ID: "command", Label: "Command", Type: engines.OptionText, Scope: engines.ScopeSetup,
				Default: `piper --model {voice} --output_file {output} --length_scale {length_scale} --input_file {text_file}`,
				Help:    "Placeholders: {text} {text_file} {output} {voice} {speed} {length_scale}. Without {output}, the program's standard output is read as audio.",
				Suggestions: []string{
					`espeak-ng -w {output} -s {wpm} -v {voice} -f {text_file}`,
					`piper --model {voice} --output_file {output} --length_scale {length_scale} --input_file {text_file}`,
					`say -o {output} --file-format=WAVE --data-format=LEI16@22050 -r {wpm} -v {voice} -f {text_file}`,
				}},
			{ID: "voices", Label: "Voices", Type: engines.OptionText, Scope: engines.ScopeSetup,
				Help: "Comma-separated values for {voice}"},
		},
	}
}
