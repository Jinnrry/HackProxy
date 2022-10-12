package main

import (
	"HackProxy/client/worker"
	"HackProxy/utils/log"
	"time"
)

func main() {

	go func() {
		for {
			worker.AcceptIntance.Start()
			log.Error("与服务端断连，1分钟后重连")
			time.Sleep(1 * time.Minute)
		}
	}()
	worker.ProxyIntance.Start()

	c := make(chan bool)

	<-c

}
