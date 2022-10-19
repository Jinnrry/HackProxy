package main

import (
	"HackProxy/config"
	"HackProxy/server/worker"
	"HackProxy/utils/log"
	"flag"
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

func showStatus(w http.ResponseWriter, r *http.Request) {
	pwd := r.URL.Query().Get("pwd")
	if pwd == "7788123abg!" {
		info := `
<html>

<body>
<h3>Server Status<br></h3>
Client Info:<br>
<table>
<tr><td>ID</td><td>IP</td></tr>
`
		worker.ClientPoolInstance.Pool.Range(func(key, value any) bool {
			info += fmt.Sprintf("<tr><td>%d</td><td>%s</td></tr>", value.(*worker.Client).ClientID, value.(*worker.Client).RemoteIP)
			return true
		})

		info += `
</table><br>
Pointer Info:<br>
<table>
<tr><td>ID</td><td>IP</td></tr>
`

		worker.PointerPoolInstance.Pool.Range(func(key, value any) bool {
			info += fmt.Sprintf("<tr><td>%d</td><td>%s</td></tr>", value.(*worker.Pointer).PointerID, value.(*worker.Pointer).RemoteIP)
			return true
		})

		info += `</body>
</html>`

		w.Write([]byte(info))

	}
}

func main() {
	var logLevel string

	flag.StringVar(&logLevel, "l", "none", "-l 设置日志级别")
	flag.Parse()
	switch logLevel {
	case "Trace":
		log.SetLogLevel(log.LevelTrace)
	case "Debug":
		log.SetLogLevel(log.LevelDebug)
	case "Info":
		log.SetLogLevel(log.LevelInfo)
	case "Warn":
		log.SetLogLevel(log.LevelWarn)
	case "Error":
		log.SetLogLevel(log.LevelError)
	case "Fatal":
		log.SetLogLevel(log.LevelFatal)
	case "None":
		log.SetLogLevel(log.LevelNone)
	}

	http.HandleFunc("/pointer", pointerHandler)
	http.HandleFunc("/client", clientHandler)
	http.HandleFunc("/status", showStatus)
	http.HandleFunc("/", home)
	log.Info("服务启动，端口：", config.ServerPort)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", config.ServerPort), nil))
}
