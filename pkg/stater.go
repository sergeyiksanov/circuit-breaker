package pkg

type state int

const (
	Close    state = iota
	Open     state = iota
	HalfOpen state = iota
)

type stater interface {
	state() state
}

func (s state) state() state {
	return s
}
