package safemap

import "sync"

type SafeMap struct {
	m map[string]any
	sync.RWMutex
}

func NewSafeMap() *SafeMap {
	return &SafeMap{
		m: make(map[string]any),
	}
}

func (s *SafeMap) Get(key string) any {
	s.RLock()
	defer s.RUnlock()
	return s.m[key]
}

func (s *SafeMap) Set(key string, value any) {
	s.Lock()
	defer s.Unlock()
	s.m[key] = value
}
