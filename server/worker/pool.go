package worker

import (
	"HackProxy/utils/dto"
	"HackProxy/utils/log"
	"sync"
)

type PointerPool struct {
	Pool        sync.Map
	PointerList []*dto.PointerInfo
	PointerID   uint32
	Lock        sync.Mutex
	Length      int
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
		p.PointerID++

		if _, ok := p.Pool.Load(p.PointerID); !ok {
			return p.PointerID
		}
	}

}

func (p *PointerPool) Insert(pointer *Pointer) {
	p.Pool.Store(pointer.PointerID, pointer)
	p.PointerList = append(p.PointerList, &dto.PointerInfo{
		ID: pointer.PointerID,
		IP: pointer.RemoteIP,
	})

	// 向所有client推送pointer信息
	ClientPoolInstance.Pool.Range(func(key, value any) bool {
		go func() {
			err := value.(*Client).PushPointerInfo()
			if err != nil {
				log.Error("推送pointer信息失败，断开连接", err)
				value.(*Client).Close()
			}

		}()
		return true
	})

	p.Length++
}

func (p *ClientPool) GenClientID() uint32 {
	p.Lock.Lock()
	defer p.Lock.Unlock()

	for {
		p.ClientID++

		if _, ok := p.Pool.Load(p.ClientID); !ok {
			return p.ClientID
		}
	}
}

func (p *ClientPool) Insert(client *Client) {
	p.Pool.Store(client.ClientID, client)
}

func (p *ClientPool) Get(clientID uint32) (*Client, bool) {
	v, ok := p.Pool.Load(clientID)
	if ok {
		return v.(*Client), ok
	}
	return nil, false
}

func (p *PointerPool) GetPointerList() []*dto.PointerInfo {
	return p.PointerList
}

func (p *PointerPool) Get(pointerID uint32) (*Pointer, bool) {
	v, ok := p.Pool.Load(pointerID)
	if ok {
		return v.(*Pointer), ok

	}
	return nil, false
}
