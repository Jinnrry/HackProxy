package main

import (
	"HackProxy/client/worker"
	"HackProxy/utils/dto"
	"HackProxy/utils/log"
	"encoding/binary"
	"errors"
	"flag"
	"fmt"
	"io"
	"net"
	"time"
)

func main() {

	log.SetLogLevel(log.LevelInfo)

	var ss5prot string
	flag.StringVar(&ss5prot, "p", "1080", "-p 设置ss5代理端口")
	go func() {
		for {
			worker.AcceptInstance = &worker.Accept{}
			worker.AcceptInstance.Start()
			log.Error("与服务端断连，1分钟后重连")
			time.Sleep(1 * time.Minute)
		}
	}()

	server, err := net.Listen("tcp", fmt.Sprintf(":%s", ss5prot))
	if err != nil {
		log.Debug("Listen failed: %v\n", err)
		return
	}

	for {
		conn, err := server.Accept()
		if err != nil {
			log.Debug("Accept failed: %v", err)
			continue
		}
		go process(conn)
	}

}

func process(conn net.Conn) {
	if err := socks5Auth(conn); err != nil {
		log.Debug("auth error:", err)
		conn.Close()
		return
	}

	targetedInfo, err := getAddress(conn) // 解析出需要连接的地址
	if err != nil {
		log.Debug("connect error:", err)
		conn.Close()
		return
	}

	worker.NewProxy(conn, targetedInfo)
}

func getAddress(client net.Conn) (*dto.TargetedInfo, error) {
	buf := make([]byte, 256)

	ret := &dto.TargetedInfo{}

	n, err := io.ReadFull(client, buf[:4])
	if n != 4 {
		return ret, errors.New("read header: " + err.Error())
	}

	ver, cmd, _, atyp := buf[0], buf[1], buf[2], buf[3]
	if ver != 5 || cmd != 1 {
		return ret, errors.New("invalid ver/cmd")
	}

	addr := ""
	switch atyp {
	case 1:
		n, err = io.ReadFull(client, buf[:4])
		if n != 4 {
			return ret, errors.New("invalid IPv4: " + err.Error())
		}
		addr = fmt.Sprintf("%d.%d.%d.%d", buf[0], buf[1], buf[2], buf[3])

	case 3:
		n, err = io.ReadFull(client, buf[:1])
		if n != 1 {
			return ret, errors.New("invalid hostname: " + err.Error())
		}
		addrLen := int(buf[0])

		n, err = io.ReadFull(client, buf[:addrLen])
		if n != addrLen {
			return ret, errors.New("invalid hostname: " + err.Error())
		}
		addr = string(buf[:addrLen])

	case 4:
		return ret, errors.New("IPv6: no supported yet")

	default:
		return ret, errors.New("invalid atyp")
	}

	n, err = io.ReadFull(client, buf[:2])
	if n != 2 {
		return ret, errors.New("read port: " + err.Error())
	}
	port := binary.BigEndian.Uint16(buf[:2])

	//n, err = client.Write([]byte{0x05, 0x00, 0x00, 0x01, 0, 0, 0, 0, 0, 0})
	//if err != nil {
	//	return ret, errors.New("write rsp: " + err.Error())
	//}

	ret.Port = port
	ret.IP = addr
	ret.AType = atyp

	return ret, nil
}

func socks5Auth(client net.Conn) (err error) {
	buf := make([]byte, 256)

	// 读取 VER 和 NMETHODS
	n, err := io.ReadFull(client, buf[:2])
	if n != 2 {
		return errors.New("reading header: " + err.Error())
	}

	ver, nMethods := int(buf[0]), int(buf[1])
	if ver != 5 {
		return errors.New("invalid version")
	}

	// 读取 METHODS 列表
	n, err = io.ReadFull(client, buf[:nMethods])
	if n != nMethods {
		return errors.New("reading methods: " + err.Error())
	}

	//无需认证
	n, err = client.Write([]byte{0x05, 0x00})
	if n != 2 || err != nil {
		return errors.New("write rsp: " + err.Error())
	}

	return nil
}
