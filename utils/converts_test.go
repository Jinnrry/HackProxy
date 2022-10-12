package utils

import (
	"github.com/Jinnrry/gop"
	"testing"
)

func TestUnt32ToBytes(t *testing.T) {
	res := Unt32ToBytes(17)

	gop.Print(res)
}

func TestBytesToUint32(t *testing.T) {
	res := Unt32ToBytes(17)

	res2 := BytesToUint32(res)
	gop.Print(res2)
}
