// Package sysdlg shows the platform's native open/save file dialogs by
// driving the tools every desktop already has (zenity/kdialog,
// AppleScript, PowerShell), so no extra native dependencies are needed.
package sysdlg

import "errors"

var (
	// ErrCancelled means the user closed the dialog without choosing.
	ErrCancelled = errors.New("cancelled")
	// ErrUnavailable means no dialog tool exists on this system.
	ErrUnavailable = errors.New("no file dialog available")
)

// Filter restricts the files shown, e.g. {"Audio", []string{"wav", "mp3"}}.
type Filter struct {
	Name       string
	Extensions []string
}
