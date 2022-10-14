package main

import (
	"HackProxy/pointer/worker"
	"HackProxy/utils/log"
	"time"
)

func main() {
	log.SetLogLevel(log.LevelInfo)
	for {
		worker.ProxyIntance.Start()
		log.Error("与服务端断连，1分钟后重连")
		time.Sleep(1 * time.Minute)
	}
}
