package utils

type Semaphore struct {
	tickets chan struct{}
}

func NewSemaphore(ticketsNum int) *Semaphore {
	return &Semaphore{make(chan struct{}, ticketsNum)}
}

func (s *Semaphore) Acquire() {
	s.tickets <- struct{}{}
}

func (s *Semaphore) Release() {
	<-s.tickets
}
