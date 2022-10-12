package worker

import (
	"HackProxy/config"
	"HackProxy/utils/dp"
	"HackProxy/utils/log"
	"github.com/gorilla/websocket"
	"net/url"
)

type Accept struct {
	conn     *websocket.Conn
	ClientID uint32
}

var AcceptIntance *Accept

func init() {
	AcceptIntance = &Accept{}
}

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
	authPkg := dp.NewPackage(dp.DirectionC2S, dp.TypeAuth, 0, 0, 0, []byte(config.Client2ServerAuth))
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

func (a *Accept) Write() {

}

func (a *Accept) StartRead() {
	for {
		pg, err := dp.ReadPkg(a.conn)
		if err != nil {
			log.Error("读取server数据失败", err)
			return
		}
		log.Debug("收到服务端数据", pg)
	}
}
