package persistent_msg_relay_hdl

import (
	"sync"
)

type syncSlice struct {
	sl []string
	mu sync.Mutex
}

func (s *syncSlice) Append(v ...string) {
	if len(v) == 0 {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.sl = append(s.sl, v...)
}

func (s *syncSlice) Values() []string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.sl
}
