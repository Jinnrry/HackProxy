package worker

import (
	"HackProxy/config"
	"HackProxy/utils/dp"
	"HackProxy/utils/log"
	"github.com/gorilla/websocket"
	"sync"
	"time"
)

type Pointer struct {
	WebSocketConn *websocket.Conn
	PointerID     uint32
	RemoteIP      string
	Lock          sync.Mutex
	Enabled       bool
}

func NewPointer(conn *websocket.Conn) {
	instace := &Pointer{
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

		if pg.Direction == dp.DirectionP2S && string(pg.Data) == config.Pointer2ServerAuth {
			c1 <- true
			return
		}

		log.Info("pointer端秘钥错误，提供秘钥：", string(pg.Data))
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
		log.Debug("pointer端权鉴包超时")
		conn.Close()
	}

	if authSucc {
		// 生成pointer id
		instace.PointerID = PointerPoolInstance.GenPointerID()
		// 写回pointid
		err := instace.Write(dp.NewPackage(dp.DirectionS2P, dp.TypeAuth, instace.PointerID, 0, 0, 0, []byte{}))
		if err != nil {
			log.Error("写回pointer id失败", err)
			return
		}
		instace.Enabled = true
		// 插入pointer pool
		PointerPoolInstance.Insert(instace)
		// 启一个协程读取数据
		go instace.StartRead()
	}
}

func (p *Pointer) Write(pg *dp.Package) error {
	p.Lock.Lock()
	defer p.Lock.Unlock()
	pg.Debug()
	return p.WebSocketConn.WriteMessage(websocket.BinaryMessage, pg.Encode())
}

func (p *Pointer) Close() {
	p.Enabled = false
	_ = p.WebSocketConn.Close()
	PointerPoolInstance.Pool.Delete(p.PointerID)
	for i, info := range PointerPoolInstance.PointerList {
		if info.ID == p.PointerID {
			PointerPoolInstance.PointerList = append(PointerPoolInstance.PointerList[:i], PointerPoolInstance.PointerList[i+1:]...)
			break
		}
	}

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
}

func (p *Pointer) StartRead() {
	for {
		if p.Enabled {
			pg, err := dp.ReadPkg(p.WebSocketConn)
			if err != nil {
				log.Error("读取客户端数据失败,关闭连接", err)
				p.Close()
				return
			}
			pg.Debug()
			if pg.Direction == dp.DirectionP2C {
				c, ok := ClientPoolInstance.Get(pg.ClientID)
				if !ok {
					pg.Type = dp.TypeProxyFail
					log.Error("代理失败，找不到client")
					err := p.Write(pg)
					if err != nil {
						p.Close()
					}
				}
				err := c.Write(pg)
				if err != nil {
					c.Close()
					pg.Type = dp.TypeProxyFail
					log.Error("代理失败，向client写数据失败")
					err := p.Write(pg)
					if err != nil {
						p.Close()
					}
				}
			} else if pg.Direction == dp.DirectionP2CNoReplay {
				c, ok := ClientPoolInstance.Get(pg.ClientID)
				if ok {
					err := c.Write(pg)
					if err != nil {
						c.Close()
					}
				}
			}
		}

	}
}
