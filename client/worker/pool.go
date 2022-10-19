package worker

import (
	"sync"
)

type ProxyPool struct {
	Pool    sync.Map
	ProxyID uint64
	Lock    sync.Mutex
}

var ProxyPoolInstance = ProxyPool{
	ProxyID: 1,
}

func (p *ProxyPool) GenProxyID() uint64 {
	p.Lock.Lock()
	defer p.Lock.Unlock()

	for {
		p.ProxyID++
		if _, ok := p.Pool.Load(p.ProxyID); !ok {
			return p.ProxyID
		}

	}
}

func (p *ProxyPool) Insert(proxy *Proxy) {
	p.Pool.Store(proxy.ID, proxy)
}

func (p *ProxyPool) Get(proxyID uint64) (*Proxy, bool) {
	v, ok := p.Pool.Load(proxyID)
	if v == nil {
		return nil, false
	}

	return v.(*Proxy), ok
}

func (p *ProxyPool) Remove(proxyID uint64) {
	p.Pool.Delete(proxyID)
}
