package worker

import (
	"HackProxy/utils/dp"
	"HackProxy/utils/dto"
	"HackProxy/utils/log"
	"encoding/json"
	"fmt"
	"net"
	"sync"
)

type Proxy struct {
	Lock          sync.Mutex
	Enable        bool
	conn          net.Conn
	TargetAddress string
	PointerID     uint32
	AcceptID      uint64
	ID            uint64
}

func NewProxy(conn net.Conn, info *dto.TargetedInfo) {
	instance := &Proxy{
		Enable:        true,
		conn:          conn,
		ID:            ProxyPoolInstance.GenProxyID(),
		TargetAddress: fmt.Sprintf("%s:%s:%d", info.Protocol, info.IP, info.Port),
	}
	// 选取pointer
	if AcceptInstance.PickPointer == nil {
		if len(AcceptInstance.PointerInfoList) > 0 {
			AcceptInstance.PickPointer = AcceptInstance.PointerInfoList[0]
		} else {
			// 无可用的pointer
			_, _ = conn.Write([]byte{0x05, 0x01, 0x00, info.AType, 0, 0, 0, 0, 0, 0})
			_ = conn.Close()
			return
		}
	}

	instance.PointerID = AcceptInstance.PickPointer.ID
	if instance.PointerID == 0 {
		log.Fatal("错误的Pointer选择")
	}

	targetedInfo, _ := json.Marshal(info)

	// 通知远端建立连接
	err := AcceptInstance.Write(dp.NewPackage(dp.DirectionC2P, dp.TypeCreateConn,
		instance.PointerID, AcceptInstance.ClientID, 0, instance.ID, targetedInfo))
	if err != nil {
		// 通知远端失败
		_, _ = conn.Write([]byte{0x05, 0x01, 0x00, info.AType, 0, 0, 0, 0, 0, 0})
		_ = conn.Close()
		return
	}

	ProxyPoolInstance.Insert(instance)

	go instance.StartRead()
}

func (p *Proxy) Write(data []byte) error {
	p.Lock.Lock()
	defer p.Lock.Unlock()
	_, err := p.conn.Write(data)
	return err
}

func (p *Proxy) Close() {
	_ = p.conn.Close()
	ProxyPoolInstance.Remove(p.ID)
}

func (p *Proxy) StartRead() {
	for {
		size := 32 * 1024
		buf := make([]byte, size)
		n, err := p.conn.Read(buf)
		if err != nil {
			p.Close()
			// ss断连，通知pointer断连
			err := AcceptInstance.Write(dp.NewPackage(dp.DirectionC2PNoReplay, dp.TypeCloseConn, p.PointerID, AcceptInstance.ClientID, p.AcceptID, p.ID, nil))
			if err != nil {
				log.Fatal("ss断连")
				return
			}
			return
		}
		if n > 0 {
			// 转发ss数据到pointer
			err := AcceptInstance.Write(dp.NewPackage(dp.DirectionC2P, dp.TypeData, p.PointerID, AcceptInstance.ClientID, p.AcceptID, p.ID, buf[0:n]))
			if err != nil {
				log.Fatal("ss断连")
				return
			}
		}
	}
}
