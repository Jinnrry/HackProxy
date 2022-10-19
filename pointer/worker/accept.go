package worker

import (
	"HackProxy/utils/dp"
	"HackProxy/utils/dto"
	"HackProxy/utils/log"
	"errors"
	"fmt"
	"net"
	"time"
)

type Accept struct {
	conn          net.Conn
	ID            uint64
	Enabled       bool
	ClientID      uint32
	ProxyID       uint64
	TargetAddress string
}

func NewAccept(targetedInfo *dto.TargetedInfo, clientID uint32, proxyID uint64) (uint64, error) {
	instance := &Accept{
		TargetAddress: fmt.Sprintf("%s:%s:%d", targetedInfo.Protocol, targetedInfo.IP, targetedInfo.Port),
	}

	// 1秒内没建立连接就放弃
	d := net.Dialer{Timeout: 1 * time.Second}
	destAddrPort := fmt.Sprintf("%s:%d", targetedInfo.IP, targetedInfo.Port)
	dest, err := d.Dial("tcp", destAddrPort)
	if err != nil {
		log.Info("ProxyID:", proxyID, "创建连接失败", err)
		return 0, err
	}
	log.Info("ProxyID:", proxyID, "创建连接成功")

	instance.conn = dest
	instance.ID = AcceptPoolInstance.GenAcceptID()

	AcceptPoolInstance.Insert(instance)

	instance.Enabled = true
	instance.ClientID = clientID
	instance.ProxyID = proxyID

	go instance.StartRead()

	return instance.ID, nil
}

func (p *Accept) Close() {
	p.Enabled = false
	_ = p.conn.Close()
	AcceptPoolInstance.Remove(p.ID)
}

func (p *Accept) Write(data []byte) error {
	if p == nil {
		log.Error("指针为空")
		return errors.New("connect err")
	}

	if p.conn == nil {
		log.Error("连接为空")
		return errors.New("connect err")
	}

	_, err := p.conn.Write(data)

	return err

}

func (p *Accept) StartRead() {
	for {
		if p.Enabled {
			size := 32 * 1024
			buf := make([]byte, size)
			n, err := p.conn.Read(buf)
			if err != nil {
				p.Close()

				// 通知client断连
				err := ProxyIntance.Write(dp.NewPackage(dp.DirectionP2CNoReplay, dp.TypeCloseConn, ProxyIntance.PointerID, p.ClientID, p.ID, p.ProxyID, []byte(err.Error())))
				if err != nil {
					log.Fatal("与server断连")
				}
				return
			}
			if n > 0 {
				err := ProxyIntance.Write(dp.NewPackage(dp.DirectionP2C, dp.TypeData, ProxyIntance.PointerID, p.ClientID, p.ID, p.ProxyID, buf[0:n]))
				if err != nil {
					log.Fatal("与server断连")
				}
			}

		}
	}
}
