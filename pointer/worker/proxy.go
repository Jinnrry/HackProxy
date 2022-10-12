package worker

import (
	"HackProxy/config"
	"HackProxy/utils/dp"
	"HackProxy/utils/log"
	"github.com/gorilla/websocket"
	"net/url"
	"sync"
)

type Proxy struct {
	conn      *websocket.Conn
	PointerID uint32
	Lock      sync.Mutex
}

var ProxyIntance *Proxy

func init() {
	ProxyIntance = &Proxy{}
}

func (p *Proxy) Start() {
	// 和server节点建立连接
	u := url.URL{Scheme: "ws", Host: config.ServerAddress, Path: "/pointer"}
	log.Info("connecting to", u.String())

	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		log.Fatal("dial:", err)
		return
	}
	// 写入权鉴数据
	authPkg := dp.NewPackage(dp.DirectionP2S, dp.TypeAuth, 0, 0, 0, []byte(config.Pointer2ServerAuth))
	err = c.WriteMessage(websocket.BinaryMessage, authPkg.Encode())
	if err != nil {
		log.Fatal("发送权鉴包失败", err)
		return
	}

	// 接收权鉴响应包
	pg, err := dp.ReadPkg(c)
	if err != nil {
		log.Fatal("读取权鉴响应包失败", err)
	}
	if pg.Direction == dp.DirectionS2P && pg.Type == dp.TypeAuth {
		p.PointerID = pg.PointerID
		p.conn = c
		log.Info("连接server成功，pointer id :", pg.PointerID)
	} else {
		log.Fatal("未收到权鉴响应包")
		return
	}
	p.StartRead()
}

func (p *Proxy) Write(pg *dp.Package) error {
	p.Lock.Lock()
	defer p.Lock.Unlock()
	return p.conn.WriteMessage(websocket.BinaryMessage, pg.Encode())
}

func (p *Proxy) StartRead() {
	for {
		pg, err := dp.ReadPkg(p.conn)
		if err != nil {
			log.Error("读取server数据失败", err)
			return
		}
		log.Debug("收到服务端数据", pg)
	}
}
