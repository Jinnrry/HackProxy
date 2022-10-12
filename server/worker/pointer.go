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
		err := instace.Write(dp.NewPackage(dp.DirectionS2P, dp.TypeAuth, instace.PointerID, 0, 0, []byte{}))
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
	return p.WebSocketConn.WriteMessage(websocket.BinaryMessage, pg.Encode())
}

func (p *Pointer) Close() {
	p.Enabled = false
	_ = p.WebSocketConn.Close()
	PointerPoolInstance.Pool.Delete(p.PointerID)
}

func (p *Pointer) StartRead() {
	for {
		_, data, err := p.WebSocketConn.ReadMessage()
		if err != nil {
			log.Error("读取客户端数据失败,关闭连接", err)
			p.Close()
			return
		}
		_ = data
		//todo
	}
}
