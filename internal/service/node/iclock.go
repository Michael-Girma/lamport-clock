package node

type IClock interface {
	UpdateClock(value int)
	Increment()
}
