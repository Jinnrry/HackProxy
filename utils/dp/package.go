package dp

import (
	"HackProxy/utils"
	"HackProxy/utils/log"
	"fmt"
)

const (
	DirectionC2P = 1
	DirectionP2C = 2
	DirectionC2S = 3
	DirectionP2S = 4
	DirectionS2P = 5
	DirectionS2C = 6

	// 不需要回复
	DirectionC2PNoReplay = 7
	DirectionP2CNoReplay = 8
	DirectionC2SNoReplay = 9
	DirectionP2SNoReplay = 10
	DirectionS2PNoReplay = 11
	DirectionS2CNoReplay = 12
)

const (
	TypeAuth           = 1
	TypePointerInfo    = 2
	TypeCreateConn     = 3
	TypeCreateConnFail = 4
	TypeCreateConnSucc = 5
	TypeProxyFail      = 6
	TypeCloseConn      = 7
	TypeData           = 8
)

type Package struct {
	Direction uint8 // 1表示从client到pointer的包，2表示从pointer到client
	Type      uint8
	PointerID uint32
	ClientID  uint32
	AcceptID  uint64 // 远端id
	ProxyID   uint64 // 本地端id
	Data      []byte
}

func NewPackage(direction, ptype uint8, pointerID, clientID uint32, acceptID, proxyID uint64, data []byte) *Package {
	return &Package{
		Direction: direction,
		Type:      ptype,
		PointerID: pointerID,
		AcceptID:  acceptID,
		ProxyID:   proxyID,
		ClientID:  clientID,
		Data:      data,
	}
}

func (p *Package) Encode() []byte {
	// 前10个字节为header内容
	ret := []byte{p.Direction, p.Type}

	ret = append(ret, utils.Unt32ToBytes(p.PointerID)...)
	ret = append(ret, utils.Unt32ToBytes(p.ClientID)...)
	ret = append(ret, utils.Unt64ToBytes(p.AcceptID)...)
	ret = append(ret, utils.Unt64ToBytes(p.ProxyID)...)
	ret = append(ret, p.Data...)

	return ret
}

func DecodePackage(data []byte) *Package {
	ret := &Package{
		Direction: data[0],
		Type:      data[1],
		PointerID: utils.BytesToUint32(data[2:6]),
		ClientID:  utils.BytesToUint32(data[6:10]),
		AcceptID:  utils.BytesToUint64(data[10:18]),
		ProxyID:   utils.BytesToUint64(data[18:26]),
		Data:      data[26:],
	}

	return ret
}

func (p *Package) Debug() {
	if log.GetLogLevel() != log.LevelDebug {
		return
	}

	DirMap := map[uint8]string{
		DirectionC2P:         "c2p",
		DirectionP2C:         "p2c",
		DirectionC2S:         "c2s",
		DirectionP2S:         "p2s",
		DirectionS2P:         "s2p",
		DirectionS2C:         "s2c",
		DirectionC2PNoReplay: "C2PNoReplay",
		DirectionP2CNoReplay: "P2CNoReplay",
		DirectionC2SNoReplay: "C2SNoReplay",
		DirectionP2SNoReplay: "P2SNoReplay",
		DirectionS2PNoReplay: "S2PNoReplay",
		DirectionS2CNoReplay: "S2CNoReplay",
	}

	TypeMap := map[uint8]string{
		TypeAuth:           "权鉴",
		TypePointerInfo:    "Pointer信息",
		TypeCreateConn:     "创建连接",
		TypeCreateConnFail: "建连失败",
		TypeCreateConnSucc: "建连成功",
		TypeProxyFail:      "代理失败",
		TypeCloseConn:      "关闭连接",
		TypeData:           "数据包",
	}

	//if p.Type == TypeData {
	//
	//	fmt.Printf("方向：%s,类型:%s,PointerID:%d,ClientID:%d,AcceptID:%d,ProxyID:%d \n",
	//		DirMap[p.Direction], TypeMap[p.Type], p.PointerID, p.ClientID, p.AcceptID, p.ProxyID,
	//	)
	//} else {
	//	if utf8.ValidString(string(p.Data)) {
	fmt.Printf("方向：%s,类型:%s,PointerID:%d,ClientID:%d,AcceptID:%d,ProxyID:%d,Data:%s \n",
		DirMap[p.Direction], TypeMap[p.Type], p.PointerID, p.ClientID, p.AcceptID, p.ProxyID, string(p.Data),
	)
	//} else {
	//	fmt.Printf("方向：%s,类型:%s,PointerID:%d,ClientID:%d,AcceptID:%d,ProxyID:%d \n",
	//		DirMap[p.Direction], TypeMap[p.Type], p.PointerID, p.ClientID, p.AcceptID, p.ProxyID,
	//	)
	//}
	//}

}
