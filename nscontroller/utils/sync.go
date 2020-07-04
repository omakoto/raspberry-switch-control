package utils

import "sync"

type Synchronized struct {
	Mutex *sync.Mutex
}

func NewSynchronized() *Synchronized {
	return &Synchronized{&sync.Mutex{}}
}

func (s *Synchronized) Run(f func()) {
	s.Mutex.Lock()
	defer s.Mutex.Unlock()

	f()
}

func (s *Synchronized) RunForValue(f func() interface{}) interface{} {
	s.Mutex.Lock()
	defer s.Mutex.Unlock()

	return f()
}
