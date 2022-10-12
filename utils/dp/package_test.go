package dp

import (
	"github.com/Jinnrry/gop"
	"testing"
)

func TestPackage_Encode(t *testing.T) {
	pg := NewPackage(DirectionP2C, 0, 1, 1, []byte("HelloWorld!"))
	bytePg := pg.Encode()
	gop.Print(bytePg)

	pg = DecodePackage(bytePg)
	gop.Print(pg)
}
