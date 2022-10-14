package worker

import (
	"HackProxy/utils/dp"
	"HackProxy/utils/dto"
	"HackProxy/utils/log"
	"fmt"
	"io"
	"net"
)

type Accept struct {
	conn     net.Conn
	ID       uint64
	Enabled  bool
	ClientID uint32
	ProxyID  uint64
}

func NewAccept(targetedInfo *dto.TargetedInfo, clientID uint32, proxyID uint64) (uint64, error) {
	instance := &Accept{}

	destAddrPort := fmt.Sprintf("%s:%d", targetedInfo.IP, targetedInfo.Port)
	dest, err := net.Dial("tcp", destAddrPort)
	if err != nil {
		return 0, err
	}
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
	_, err := p.conn.Write(data)

	return err

}

func (p *Accept) StartRead() {
	for {
		if p.Enabled {
			size := 32 * 1024
			buf := make([]byte, size)
			n, err := p.conn.Read(buf)
			if err != nil && err != io.EOF {
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
