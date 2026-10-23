package manager

import (
	"regexp"
	"strconv"
	"strings"
)

var (
	// tqdmLine matches progress bars printed by Python tools, e.g.
	// "model.safetensors:  45%|████▌     | 450M/1.0G [00:10<00:12, 45MB/s]".
	tqdmLine = regexp.MustCompile(`^(.*?):?\s*(\d{1,3})%\s*\|[^|]*\|\s*(.*)$`)
	count    = regexp.MustCompile(`^(\S+/\S+)`)
	eta      = regexp.MustCompile(`<([\d:]+)`)
	percent  = regexp.MustCompile(`(\d{1,3}(?:\.\d+)?)\s?%`)
)

// tidy turns a raw log line into a short status for the UI and the
// progress it reports (-1 when none).
func tidy(line string) (string, float32) {
	line = strings.Map(func(r rune) rune {
		switch {
		case r >= 0x2580 && r <= 0x259F: // block elements drawn by text progress bars
			return -1
		case r < 0x20 && r != '\t':
			return -1
		}
		return r
	}, line)
	line = strings.TrimSpace(line)

	if m := tqdmLine.FindStringSubmatch(line); m != nil {
		pct, _ := strconv.Atoi(m[2])
		desc := strings.TrimSpace(m[1])
		parts := []string{}
		switch {
		case desc == "":
			parts = append(parts, "Working")
		case strings.Contains(desc, "."):
			parts = append(parts, "Downloading "+desc) // a file name
		default:
			parts = append(parts, desc)
		}
		parts = append(parts, strconv.Itoa(pct)+"%")
		if c := count.FindStringSubmatch(m[3]); c != nil {
			parts = append(parts, c[1])
		}
		if e := eta.FindStringSubmatch(m[3]); e != nil && e[1] != "00:00" && e[1] != "?" {
			parts = append(parts, e[1]+" left")
		}
		return strings.Join(parts, " · "), float32(min(pct, 100)) / 100
	}

	progress := float32(-1)
	if m := percent.FindStringSubmatch(line); m != nil {
		if v, err := strconv.ParseFloat(m[1], 64); err == nil && v <= 100 {
			progress = float32(v / 100)
		}
	}
	if r := []rune(line); len(r) > 140 {
		line = string(r[:140]) + "…"
	}
	return line, progress
}
