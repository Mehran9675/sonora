package transcripts

import (
	"testing"
	"time"

	"github.com/mehran9675/sonora/internal/engines"
)

func TestSRT(t *testing.T) {
	e := Entry{Segments: []engines.Segment{{Start: 0.5, End: 3723.25, Text: " Hello "}, {Start: 3724, End: 3725, Text: "there"}}}
	want := "1\n00:00:00,500 --> 01:02:03,250\nHello\n\n2\n01:02:04,000 --> 01:02:05,000\nthere\n\n"
	if got := SRT(e); got != want {
		t.Fatalf("got %q", got)
	}
	if got := VTT(Entry{Text: "Hi", Duration: 2 * time.Second}); got != "WEBVTT\n\n00:00:00.000 --> 00:00:02.000\nHi\n\n" {
		t.Fatalf("got %q", got)
	}
}

func TestStore(t *testing.T) {
	dir := t.TempDir()
	s := Open(dir)
	id, _ := s.NewFile()
	s.Add(Entry{ID: id, Text: "one", Audio: id + ".wav", Created: time.Now()})
	s.SetText(id, "edited")
	s2 := Open(dir)
	if e, ok := s2.Get(id); !ok || e.Text != "edited" {
		t.Fatalf("reloaded %+v", e)
	}
	s2.Remove(id)
	if len(Open(dir).Entries()) != 0 {
		t.Fatal("not removed")
	}
}
