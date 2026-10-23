package manager

import (
	"regexp"
	"strconv"
	"sync"
	"time"
)

const maxLog = 400

var percent = regexp.MustCompile(`(\d{1,3}(?:\.\d+)?)\s?%`)

// logSink batches log lines from background goroutines and applies
// them to the runtime on the UI goroutine at most ten times a second,
// so a chatty pip install doesn't flood the render loop.
type logSink struct {
	mu      sync.Mutex
	pending []string
	timer   *time.Timer
	post    func(func())
	rt      *Runtime
}

func (m *Manager) sink(rt *Runtime) *logSink {
	return &logSink{post: m.post, rt: rt}
}

func (s *logSink) Log(line string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.pending = append(s.pending, line)
	if s.timer == nil {
		s.timer = time.AfterFunc(100*time.Millisecond, s.flush)
	}
}

func (s *logSink) flush() {
	s.mu.Lock()
	lines := s.pending
	s.pending, s.timer = nil, nil
	s.mu.Unlock()
	s.post(func() {
		rt := s.rt
		rt.Log = append(rt.Log, lines...)
		if len(rt.Log) > maxLog {
			rt.Log = rt.Log[len(rt.Log)-maxLog:]
		}
		last := lines[len(lines)-1]
		rt.Status = last
		rt.Progress = -1
		if m := percent.FindStringSubmatch(last); m != nil {
			if v, err := strconv.ParseFloat(m[1], 64); err == nil && v <= 100 {
				rt.Progress = float32(v / 100)
			}
		}
	})
}
