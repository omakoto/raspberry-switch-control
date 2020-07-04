package utils

import "sync"

type Synchronized struct {
	mu *sync.Mutex
}

func NewSynchronized() *Synchronized {
	return &Synchronized{&sync.Mutex{}}
}

func (s *Synchronized) Run(f func()) {
	s.mu.Lock()
	defer s.mu.Unlock()

	f()
}
