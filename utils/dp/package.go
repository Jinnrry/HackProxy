package dp

import "HackProxy/utils"

const (
	DirectionC2P = 1
	DirectionP2C = 2
	DirectionC2S = 3
	DirectionP2S = 4
	DirectionS2P = 5
	DirectionS2C = 6
)

const (
	TypeAuth = 1
)

type Package struct {
	Direction uint8 // 1表示从client到pointer的包，2表示从pointer到client
	Type      uint8
	PointerID uint32
	ClientID  uint32
	AcceptID  uint32
	Data      []byte
}

func NewPackage(direction, ptype uint8, pointerID, clientID, acceptID uint32, data []byte) *Package {
	return &Package{
		Direction: direction,
		Type:      ptype,
		PointerID: pointerID,
		AcceptID:  acceptID,
		ClientID:  clientID,
		Data:      data,
	}
}

func (p *Package) Encode() []byte {
	// 前10个字节为header内容
	ret := []byte{p.Direction, p.Type}

	ret = append(ret, utils.Unt32ToBytes(p.PointerID)...)
	ret = append(ret, utils.Unt32ToBytes(p.ClientID)...)
	ret = append(ret, utils.Unt32ToBytes(p.AcceptID)...)
	ret = append(ret, p.Data...)

	return ret
}

func DecodePackage(data []byte) *Package {
	ret := &Package{
		Direction: data[0],
		Type:      data[1],
		PointerID: utils.BytesToUint32(data[2:6]),
		ClientID:  utils.BytesToUint32(data[6:10]),
		AcceptID:  utils.BytesToUint32(data[10:14]),
		Data:      data[14:],
	}

	return ret
}
