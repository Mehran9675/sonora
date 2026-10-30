package library

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/mehran9675/giode"

	"github.com/mehran9675/sonora/internal/audio/opusfile"
	"github.com/mehran9675/sonora/ui/components/chip"
	"github.com/mehran9675/sonora/ui/core"
	"github.com/mehran9675/sonora/ui/icons"
	"github.com/mehran9675/sonora/ui/theme"

	"github.com/mehran9675/sonora/ui/tr"
)

// recording is an audio file in the recordings folder.
type recording struct {
	Path     string
	Name     string
	Modified time.Time
	Size     int64
	Duration time.Duration // 0 when unknown
}

// recordingList scans the recordings folder, at most once a second.
type recordingList struct {
	c       *core.Ctx
	scanned time.Time
	dir     string
	items   []recording
	lengths map[string]time.Duration // by path and modification time
}

// audioExts are the files listed as recordings.
var audioExts = map[string]bool{".opus": true, ".ogg": true, ".wav": true, ".mp3": true, ".flac": true, ".m4a": true}

func (l *recordingList) list() []recording {
	dir := l.c.RecordingsDir()
	if dir == l.dir && time.Since(l.scanned) < time.Second {
		return l.items
	}
	if l.lengths == nil {
		l.lengths = map[string]time.Duration{}
	}
	l.dir, l.scanned, l.items = dir, time.Now(), nil
	entries, _ := os.ReadDir(dir)
	for _, e := range entries {
		ext := strings.ToLower(filepath.Ext(e.Name()))
		info, err := e.Info()
		if e.IsDir() || !audioExts[ext] || err != nil {
			continue
		}
		r := recording{Path: filepath.Join(dir, e.Name()), Name: strings.TrimSuffix(e.Name(), filepath.Ext(e.Name())),
			Modified: info.ModTime(), Size: info.Size()}
		key := r.Path + "|" + info.ModTime().String()
		if d, ok := l.lengths[key]; ok {
			r.Duration = d
		} else if ext == ".opus" || ext == ".ogg" {
			r.Duration, _ = opusfile.Duration(r.Path)
			l.lengths[key] = r.Duration
		}
		l.items = append(l.items, r)
	}
	sort.Slice(l.items, func(i, j int) bool { return l.items[i].Modified.After(l.items[j].Modified) })
	return l.items
}

// refresh rescans on the next frame.
func (l *recordingList) refresh() { l.scanned = time.Time{} }

// recRow is one recording in the list.
type recRow struct {
	c          *core.Ctx
	r          recording
	play       *giode.PressableElement
	transcribe *giode.ButtonElement
	export     *giode.ButtonElement
	remove     *giode.ButtonElement
}

func newRecRow(c *core.Ctx, l *recordingList, r recording) *recRow {
	row := &recRow{c: c, r: r}
	row.play = playButton(c, r.Path, r.Name, tr.T("Recording"))
	ghost := giode.Styles{Color: "theme.muted", Background: "theme.text", Padding: giode.UniformInset(8), BorderRadius: 8}
	row.transcribe = giode.Button("", ghost).Variant(giode.VariantGhost).Icon(icons.Icon(icons.Captions, 16, "")).
		OnClick(func() { c.Transcribe(r.Path, r.Name) })
	row.export = giode.Button("", ghost).Variant(giode.VariantGhost).Icon(icons.Icon(icons.Share, 16, "")).
		OnClick(func() { c.ExportFile(r.Path, r.Name) })
	row.remove = giode.Button("", ghost).Variant(giode.VariantGhost).Icon(icons.Icon(icons.Trash, 16, "")).OnClick(func() {
		c.Toasts.ShowAction(giode.ToastWarning, tr.T("Delete {name}?", "name", r.Name), tr.T("Delete"), func() {
			giode.AfterLayout(func() {
				if c.Player.Loaded() == r.Path {
					c.Player.Unload()
					c.Now = nil
				}
				if err := os.Remove(r.Path); err != nil {
					c.Error(err)
				}
				l.refresh()
			})
		})
	})
	return row
}

func (row *recRow) element() giode.Element {
	r := row.r
	card := theme.Card()
	card.FlexDirection, card.Align, card.Gap, card.Width = giode.FlexDirectionRow, giode.AlignCenter, 16, -1
	card.Padding = giode.SymmetricInset(16, 14)
	card.Wrap = giode.WrapNoWrap
	meta := []giode.Element{chip.Chip(strings.ToUpper(strings.TrimPrefix(filepath.Ext(r.Path), ".")))}
	info := sizeLabel(r.Size) + "  ·  " + core.Ago(r.Modified)
	if r.Duration > 0 {
		info = core.Clock(r.Duration) + "  ·  " + info
	}
	meta = append(meta, giode.Text(info, theme.Caption))
	return giode.Row(card,
		row.play,
		giode.Stack(giode.Styles{FlexGrow: 1, Gap: 8},
			giode.Text(r.Name, giode.Styles{Color: "theme.text", FontSize: 14, FontWeight: giode.FontWeightMedium, TextOverflow: giode.TextOverflowEllipsis}),
			giode.Row(giode.Styles{Gap: 8, Align: giode.AlignCenter}, meta...),
		),
		giode.Row(giode.Styles{Gap: 2, Wrap: giode.WrapNoWrap}, row.transcribe, row.export, row.remove),
	)
}

// sizeLabel formats a file size.
func sizeLabel(n int64) string {
	switch {
	case n >= 1<<30:
		return tr.T("{size} GB", "size", fmt.Sprintf("%.1f", float64(n)/(1<<30)))
	case n >= 1<<20:
		return tr.T("{size} MB", "size", fmt.Sprintf("%.1f", float64(n)/(1<<20)))
	}
	return tr.T("{size} KB", "size", (n+1023)>>10)
}
