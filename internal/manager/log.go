package manager

import (
	"sync"
	"time"
)

const maxLog = 400

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
		for _, l := range lines {
			rt.Log = appendLine(rt.Log, l)
		}
		if len(rt.Log) > maxLog {
			rt.Log = rt.Log[len(rt.Log)-maxLog:]
		}
		rt.Status, rt.Progress = tidy(lines[len(lines)-1])
	})
}

// appendLine adds l to the log, replacing the previous line when both
// are updates of the same progress bar, so a download shows as one
// line rather than hundreds.
func appendLine(log []string, l string) []string {
	if n := len(log); n > 0 {
		prev, cur := tqdmLine.FindStringSubmatch(log[n-1]), tqdmLine.FindStringSubmatch(l)
		if prev != nil && cur != nil && prev[1] == cur[1] {
			log[n-1] = l
			return log
		}
	}
	return append(log, l)
}
