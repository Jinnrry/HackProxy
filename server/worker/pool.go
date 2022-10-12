package worker

import (
	"sync"
)

type PointerPool struct {
	Pool      sync.Map
	PointerID uint32
	Lock      sync.Mutex
}

type ClientPool struct {
	Pool     sync.Map
	ClientID uint32
	Lock     sync.Mutex
}

var ClientPoolInstance = ClientPool{
	ClientID: 1,
}

var PointerPoolInstance = PointerPool{
	PointerID: 1,
}

func (p *PointerPool) GenPointerID() uint32 {
	p.Lock.Lock()
	defer p.Lock.Unlock()

	for {
		if _, ok := p.Pool.Load(p.PointerID); !ok {
			return p.PointerID
		}
		p.PointerID++
	}

}

func (p *PointerPool) Insert(pointer *Pointer) {
	p.Pool.Store(pointer.PointerID, pointer)
}

func (p *ClientPool) GenClientID() uint32 {
	p.Lock.Lock()
	defer p.Lock.Unlock()

	for {
		if _, ok := p.Pool.Load(p.ClientID); !ok {
			return p.ClientID
		}
		p.ClientID++
	}

}

func (p *ClientPool) Insert(client *Client) {
	p.Pool.Store(client.ClientID, client)
}
