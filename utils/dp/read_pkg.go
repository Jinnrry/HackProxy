package dp

import "github.com/gorilla/websocket"

func ReadPkg(conn *websocket.Conn) (*Package, error) {
	_, data, err := conn.ReadMessage()
	if err != nil {
		return nil, err
	}

	return DecodePackage(data), nil

}
