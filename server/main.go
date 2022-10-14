package main

import (
	"HackProxy/config"
	"HackProxy/server/worker"
	"HackProxy/utils/log"
	"fmt"
	"github.com/gorilla/websocket"
	"net/http"
)

var upgrader = websocket.Upgrader{} // use default options

func pointerHandler(w http.ResponseWriter, r *http.Request) {
	// Upgrade our raw HTTP connection to a websocket based one
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Error("Error during connection upgradation:", err)
		return
	}
	worker.NewPointer(conn)
}

func clientHandler(w http.ResponseWriter, r *http.Request) {
	// Upgrade our raw HTTP connection to a websocket based one
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Error("Error during connection upgradation:", err)
		return
	}
	worker.NewClient(conn)
}

func home(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hello World!")
}

func main() {
	log.SetLogLevel(log.LevelInfo)

	http.HandleFunc("/pointer", pointerHandler)
	http.HandleFunc("/client", clientHandler)
	http.HandleFunc("/", home)
	log.Info("服务启动，端口：", config.ServerPort)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", config.ServerPort), nil))
}
