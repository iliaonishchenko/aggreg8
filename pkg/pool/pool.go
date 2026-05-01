package pool

import "sync"

// generate:reset
type Pool[T Resetable] struct {
	values []T
	m      sync.Mutex
}

type Resetable interface {
	Reset()
}

func New[T Resetable]() *Pool[T] {
	return &Pool[T]{
		values: make([]T, 0),
	}
}
func (p *Pool[T]) Get() T {
	p.m.Lock()
	defer p.m.Unlock()

	var zero T
	n := len(p.values)
	if n == 0 {
		return zero
	}
	last := p.values[n-1]
	p.values = p.values[:n-1]
	return last

}

func (p *Pool[T]) Put(t T) {
	t.Reset()
	p.m.Lock()
	defer p.m.Unlock()
	p.values = append(p.values, t)
}
