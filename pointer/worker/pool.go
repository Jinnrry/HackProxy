package worker

import (
	"sync"
)

type AcceptPool struct {
	Pool     sync.Map
	AcceptID uint64
	Lock     sync.Mutex
}

var AcceptPoolInstance = AcceptPool{
	AcceptID: 1,
}

func (p *AcceptPool) GenAcceptID() uint64 {
	p.Lock.Lock()
	defer p.Lock.Unlock()

	for {
		if _, ok := p.Pool.Load(p.AcceptID); !ok {
			return p.AcceptID
		}
		p.AcceptID++
	}
}

func (p *AcceptPool) Insert(Accept *Accept) {
	p.Pool.Store(Accept.ID, Accept)
}

func (p *AcceptPool) Remove(acceptID uint64) {
	p.Pool.Delete(acceptID)
}

func (p *AcceptPool) Get(acceptID uint64) (*Accept, bool) {
	v, ok := p.Pool.Load(acceptID)
	if ok {
		if v == nil {
			return nil, false
		}
		return v.(*Accept), ok
	}
	return nil, false
}
