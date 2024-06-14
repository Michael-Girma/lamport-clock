package node

import "sync"

type Clock struct {
	Value int
	Mutex *sync.Mutex
}

func (clock *Clock) UpdateClock(newValue int) {
	clock.Mutex.Lock()
	defer clock.Mutex.Unlock()
	if clock.Value < newValue {
		clock.Value = newValue
	}
	clock.Value++
}

func (clock *Clock) Increment() {
	clock.Value++
}

func NewClock() *Clock {
	return &Clock{
		Value: 0,
		Mutex: &sync.Mutex{},
	}
}
