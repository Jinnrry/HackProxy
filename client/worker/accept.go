package worker

import (
	"HackProxy/config"
	"HackProxy/utils/dp"
	"HackProxy/utils/dto"
	"HackProxy/utils/log"
	"encoding/json"
	"github.com/gorilla/websocket"
	"net/url"
	"sync"
)

type Accept struct {
	conn            *websocket.Conn
	ClientID        uint32
	PointerInfoList []*dto.PointerInfo
	PickPointer     *dto.PointerInfo
	Lock            sync.Mutex
}

var AcceptInstance *Accept

func (a *Accept) Start() {
	// 和server节点建立连接
	u := url.URL{Scheme: "ws", Host: config.ServerAddress, Path: "/client"}
	log.Info("connecting to", u.String())

	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		log.Fatal("dial:", err)
		return
	}
	// 写入权鉴数据
	authPkg := dp.NewPackage(dp.DirectionC2S, dp.TypeAuth, 0, 0, 0, 0, []byte(config.Client2ServerAuth))
	err = c.WriteMessage(websocket.BinaryMessage, authPkg.Encode())
	if err != nil {
		log.Fatal("发送权鉴包失败", err)
		return
	}

	// 接收权鉴响应包
	pg, err := dp.ReadPkg(c)
	if err != nil {
		log.Fatal("读取权鉴响应包失败", err)
		return
	}
	if pg.Direction == dp.DirectionS2C && pg.Type == dp.TypeAuth {
		a.ClientID = pg.ClientID
		a.conn = c
		log.Info("连接server成功，client id :", pg.ClientID)
	} else {
		log.Fatal("未收到权鉴响应包")
		return
	}
	a.StartRead()
}

func (a *Accept) Write(pg *dp.Package) error {
	a.Lock.Lock()
	defer a.Lock.Unlock()
	pg.Debug()
	return a.conn.WriteMessage(websocket.BinaryMessage, pg.Encode())
}

func (a *Accept) StartRead() {
	for {
		pg, err := dp.ReadPkg(a.conn)
		if err != nil {
			log.Error("读取server数据失败", err)
			return
		}
		pg.Debug()
		switch pg.Type {
		case dp.TypePointerInfo:
			_ = json.Unmarshal(pg.Data, &a.PointerInfoList)
		case dp.TypeCreateConnSucc:
			proxy, ok := ProxyPoolInstance.Get(pg.ProxyID)
			if !ok {
				pg.Type = dp.TypeCloseConn
				pg.Direction = dp.DirectionC2PNoReplay
				err := a.Write(pg)
				if err != nil {
					log.Fatal("写server失败", err)
					return
				}
			} else {
				var tragetedInfo dto.TargetedInfo
				proxy.AcceptID = pg.AcceptID
				err := json.Unmarshal(pg.Data, &tragetedInfo)
				if err != nil {
					err = proxy.Write([]byte{0x05, 0x00, 0x00, 0x01, 0, 0, 0, 0, 0, 0})
				} else {
					err = proxy.Write([]byte{0x05, 0x00, 0x00, tragetedInfo.AType, 0, 0, 0, 0, 0, 0})
				}
				if err != nil {
					proxy.Close()
					pg.Type = dp.TypeCloseConn
					pg.Direction = dp.DirectionC2PNoReplay
					err := a.Write(pg)
					if err != nil {
						log.Fatal("写server失败", err)
						return
					}
				}
			}

		case dp.TypeData:
			proxy, ok := ProxyPoolInstance.Get(pg.ProxyID)
			if !ok {
				// 找不到proxy，通知pointer断连
				pg.Type = dp.TypeCloseConn
				pg.Direction = dp.DirectionC2PNoReplay
				log.Error("代理失败，找不到proxy")
				err := a.Write(pg)
				if err != nil {
					log.Fatal("写server失败", err)
					return
				}
			} else {
				err := proxy.Write(pg.Data)
				if err != nil {
					proxy.Close()
					pg.Type = dp.TypeCloseConn
					pg.Direction = dp.DirectionC2PNoReplay
					err := a.Write(pg)
					if err != nil {
						log.Fatal("写ss5失败", err)
						return
					}
				} else {
					log.Debug("写ss5成功")
				}
			}
		case dp.TypeCloseConn:
			proxy, ok := ProxyPoolInstance.Get(pg.ProxyID)
			if ok {
				proxy.Close()
			}

		case dp.TypeProxyFail:
			proxy, ok := ProxyPoolInstance.Get(pg.ProxyID)
			if ok {
				proxy.Close()
			}

		case dp.TypeCreateConnFail:
			proxy, ok := ProxyPoolInstance.Get(pg.ProxyID)
			if ok {
				proxy.Close()
			}

		default:
			log.Fatal("该类型未定义处理方法", pg.Type)
		}
	}
}
