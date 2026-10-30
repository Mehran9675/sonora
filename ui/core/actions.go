package core

import (
	"github.com/mehran9675/sonora/internal/transcripts"
	"github.com/mehran9675/sonora/ui/components/actionsheet"
	"github.com/mehran9675/sonora/ui/icons"

	"github.com/mehran9675/sonora/ui/tr"
)

// ExportFile opens the sheet to save or share the audio file at path
// as it is (a generated WAV, an Opus recording).
func (c *Ctx) ExportFile(path, title string) {
	share := actionsheet.Action{Icon: icons.Share, Label: tr.T("Share…"), Detail: tr.T("Send the file to another app"), Do: func() { c.ShareClip(path) }}
	if !Mobile {
		share = actionsheet.Action{Icon: icons.Folder, Label: tr.T("Show in folder"), Do: func() { c.ShareClip(path) }}
	}
	c.Sheet.Show(title,
		actionsheet.Action{Icon: icons.Download, Label: tr.T("Save a copy…"), Detail: tr.T("Choose where to keep the file"), Do: func() { c.SaveClip(path, title) }},
		share,
	)
}

// ExportTranscript opens the sheet to save or share a transcript as
// text or as SRT subtitles. text is the transcript as shown, edits
// included; the subtitles follow the recognized segments.
func (c *Ctx) ExportTranscript(e transcripts.Entry, text string) {
	srt := transcripts.SRT(e)
	actions := []actionsheet.Action{
		{Icon: icons.FileText, Label: tr.T("Save as text…"), Detail: ".txt", Do: func() { c.SaveText(FileName(e.Title, ".txt"), text) }},
		{Icon: icons.Captions, Label: tr.T("Save as subtitles…"), Detail: tr.T(".srt, with timings"), Do: func() { c.SaveText(FileName(e.Title, ".srt"), srt) }},
	}
	// Desktops have Copy for the text and no share sheet.
	if Mobile {
		actions = append(actions,
			actionsheet.Action{Icon: icons.Share, Label: tr.T("Share as text…"), Do: func() { c.ShareText(text, e.Title) }},
			actionsheet.Action{Icon: icons.Share, Label: tr.T("Share subtitles…"), Detail: tr.T("Sends an .srt file"), Do: func() { c.ShareTextFile(FileName(e.Title, ".srt"), srt) }},
		)
	}
	c.Sheet.Show(e.Title, actions...)
}
