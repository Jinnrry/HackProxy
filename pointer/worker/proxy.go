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
	authPkg := dp.NewPackage(dp.DirectionP2S, dp.TypeAuth, 0, 0, 0, 0, []byte(config.Pointer2ServerAuth))
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

		switch pg.Type {
		case dp.TypeCreateConn:
			var tragetInfo *dto.TargetedInfo
			err := json.Unmarshal(pg.Data, &tragetInfo)
			if err != nil {
				pg.Type = dp.TypeCreateConnFail
				pg.Direction = dp.DirectionP2C
				err := p.Write(pg)
				if err != nil {
					log.Debug("pointer写向server失败", err)
					return
				}
			}
			acceptID, err2 := NewAccept(tragetInfo, pg.ClientID, pg.ProxyID)
			if err2 != nil {
				pg.Type = dp.TypeCreateConnFail
				pg.Direction = dp.DirectionP2C
				pg.Data = []byte(err2.Error())
				err3 := p.Write(pg)
				if err3 != nil {
					log.Debug("pointer写向server失败", err3)
					return
				}
			} else {
				pg.Type = dp.TypeCreateConnSucc
				pg.Direction = dp.DirectionP2C
				pg.AcceptID = acceptID
				err4 := p.Write(pg)
				if err4 != nil {
					log.Debug("pointer写向server失败", err4)
					return
				}
			}

		case dp.TypeData:
			accept, ok := AcceptPoolInstance.Get(pg.AcceptID)
			if !ok {
				pg.Type = dp.TypeProxyFail
				pg.Direction = dp.DirectionP2CNoReplay
				pg.Data = nil
				err3 := p.Write(pg)
				if err3 != nil {
					log.Debug("pointer写向server失败", err3)
					return
				}
			}
			err := accept.Write(pg.Data)
			if err != nil {
				pg.Type = dp.TypeProxyFail
				pg.Direction = dp.DirectionP2CNoReplay
				pg.Data = []byte(err.Error())
				err3 := p.Write(pg)
				if err3 != nil {
					log.Debug("pointer写向server失败", err3)
					return
				}
			}

		case dp.TypeCloseConn:
			accept, ok := AcceptPoolInstance.Get(pg.AcceptID)
			if ok {
				accept.Close()
			}

		default:
			log.Fatal("该类型未定义处理方法", pg.Type)
		}
	}
}
