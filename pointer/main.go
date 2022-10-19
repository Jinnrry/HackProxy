package main

import (
	"HackProxy/config"
	"HackProxy/pointer/worker"
	"HackProxy/utils/log"
	"flag"
	"fmt"
	"net/http"
	"time"
)

func showStatus(w http.ResponseWriter, r *http.Request) {
	info := `
<html>
<body>
<h3>Pointer Status</h3>
Accept Info:<br>
<table>
<tr><td>AcceptID</td><td>ClientID</td><td>ProxyID</td><td>TargetAddress</td></tr>
`
	worker.AcceptPoolInstance.Pool.Range(func(key, value any) bool {
		info += fmt.Sprintf("<tr><td>%d</td><td>%d</td><td>%d</td><td>%s</td></tr>", value.(*worker.Accept).ID, value.(*worker.Accept).ClientID, value.(*worker.Accept).ProxyID, value.(*worker.Accept).TargetAddress)
		return true
	})
	info += `
</table>
</body>
</html>
`

	w.Write([]byte(info))
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

	if logLevel == "Debug" {
		go func() {
			// 开启status展示
			http.HandleFunc("/status", showStatus)
			log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", config.ServerPort+1), nil))
		}()
	}

	for {
		worker.ProxyIntance.Start()
		log.Error("与服务端断连，1分钟后重连")
		time.Sleep(1 * time.Minute)
	}
}
