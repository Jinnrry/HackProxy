package worker

import (
	"fmt"
	"testing"
)

func TestProxyPool_GenProxyID(t *testing.T) {

	for i := 0; i <= 20; i++ {
		p := ProxyPoolInstance.GenProxyID()
		fmt.Println(p)
	}

}
