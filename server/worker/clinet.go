package worker

import "C"
import (
	"HackProxy/config"
	"HackProxy/utils/dp"
	"HackProxy/utils/log"
	"encoding/json"
	"github.com/gorilla/websocket"
	"sync"
	"time"
)

type Client struct {
	WebSocketConn *websocket.Conn
	RemoteIP      string
	ClientID      uint32
	Enabled       bool
	Lock          sync.Mutex
}

func NewClient(conn *websocket.Conn) {

	instace := &Client{
		WebSocketConn: conn,
		RemoteIP:      conn.RemoteAddr().String(),
	}
	c1 := make(chan bool, 1)

	go func() {
		pg, err := dp.ReadPkg(conn)
		if err != nil {
			c1 <- false
			return
		}
		if pg.Type != dp.TypeAuth {
			c1 <- false
			return
		}

		if pg.Direction == dp.DirectionC2S && string(pg.Data) == config.Client2ServerAuth {
			c1 <- true
			return
		}

		log.Info("client端秘钥错误，提供秘钥：", string(pg.Data))
		c1 <- false
	}()

	authSucc := false

	select {
	case res := <-c1:
		if res {
			authSucc = true
			break
		}
		conn.Close()
	case <-time.After(time.Second * config.AuthWaitTime):
		log.Debug("client端权鉴包超时")
		conn.Close()
	}

	if authSucc {
		// 生成client id
		instace.ClientID = ClientPoolInstance.GenClientID()
		// 写回id
		err := instace.Write(dp.NewPackage(dp.DirectionS2C, dp.TypeAuth, 0, instace.ClientID, 0, 0, []byte{}))
		if err != nil {
			log.Error("写回pointer id失败", err)
			return
		}
		instace.Enabled = true
		// 插入pointer pool
		ClientPoolInstance.Insert(instace)
		// 推送pointer信息
		err = instace.PushPointerInfo()
		if err != nil {
			instace.Close()
			return
		}

		// 启一个协程读取数据
		go instace.StartRead()
	}

}

func (p *Client) Write(pg *dp.Package) error {
	p.Lock.Lock()
	defer p.Lock.Unlock()
	pg.Debug()
	return p.WebSocketConn.WriteMessage(websocket.BinaryMessage, pg.Encode())
}

func (p *Client) PushPointerInfo() error {
	infoList := PointerPoolInstance.GetPointerList()
	byteData, _ := json.Marshal(infoList)
	pg := dp.NewPackage(dp.DirectionS2C, dp.TypePointerInfo, 0, p.ClientID, 0, 0, byteData)
	return p.Write(pg)
}

func (p *Client) Close() {
	p.Enabled = false
	_ = p.WebSocketConn.Close()
	ClientPoolInstance.Pool.Delete(p.ClientID)
}

func (p *Client) StartRead() {
	for {
		if p.Enabled {
			pg, err := dp.ReadPkg(p.WebSocketConn)
			if err != nil {
				log.Error("客户端读取错误", err)
				log.Error("client id:", p.ClientID, "断开连接")
				p.Close()
				return
			}
			pg.Debug()
			if pg.Direction == dp.DirectionC2P {
				pointer, ok := PointerPoolInstance.Get(pg.PointerID)
				if !ok {
					pg.Type = dp.TypeProxyFail
					pg.Direction = dp.DirectionS2C
					log.Error("代理失败，找pointer失败", pg.PointerID)
					err := p.Write(pg)
					if err != nil {
						p.Close()
					}
				} else {
					err := pointer.Write(pg)
					if err != nil {
						pointer.Close()
						pg.Type = dp.TypeProxyFail
						pg.Direction = dp.DirectionS2C
						log.Error("代理失败，写pointer失败", pg.PointerID)
						err := p.Write(pg)
						if err != nil {
							p.Close()
						}
					}
				}
			} else if pg.Direction == dp.DirectionC2PNoReplay {
				pointer, ok := PointerPoolInstance.Get(pg.PointerID)
				if ok {
					err := pointer.Write(pg)
					if err != nil {
						pointer.Close()
					}
				}
			}
		}
	}
}
