package core

import (
	"fmt"
	"strings"
	"time"

	"github.com/mehran9675/sonora/ui/tr"
)

// Clock formats a duration as m:ss, or h:mm:ss from an hour.
func Clock(d time.Duration) string {
	s := int(d.Round(time.Second).Seconds())
	if s >= 3600 {
		return fmt.Sprintf("%d:%02d:%02d", s/3600, s/60%60, s%60)
	}
	return fmt.Sprintf("%d:%02d", s/60, s%60)
}

// months are the short month names Ago shows, translated.
var months = [...]string{tr.Msg("Jan"), tr.Msg("Feb"), tr.Msg("Mar"), tr.Msg("Apr"), tr.Msg("May"), tr.Msg("Jun"),
	tr.Msg("Jul"), tr.Msg("Aug"), tr.Msg("Sep"), tr.Msg("Oct"), tr.Msg("Nov"), tr.Msg("Dec")}

// Ago formats a timestamp relative to now.
func Ago(t time.Time) string {
	d := time.Since(t)
	switch {
	case d < time.Minute:
		return tr.T("just now")
	case d < time.Hour:
		return tr.N("{n} min ago", int(d.Minutes()))
	case d < 24*time.Hour:
		return tr.N("{n} h ago", int(d.Hours()))
	case d < 7*24*time.Hour:
		return tr.N("{n} d ago", int(d.Hours()/24))
	}
	return tr.T("{month} {day}, {year}", "month", tr.T(months[t.Month()-1]), "day", t.Day(), "year", t.Year())
}

// onDevice notes an engine that runs in the app ("hexgrad · on device").
var onDevice = tr.Msg("on device")

// Parts translates each part of a " · "-separated line from an engine
// ("hexgrad · on device", "High quality · 110 MB"); names and sizes
// have no translation and stay as they are.
func Parts(s string) string {
	parts := strings.Split(s, " · ")
	for i, p := range parts {
		parts[i] = tr.T(p)
	}
	return strings.Join(parts, " · ")
}
