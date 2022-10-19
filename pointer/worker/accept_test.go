package worker

import (
	"HackProxy/utils/dto"
	"testing"
)

func TestNewAccept(t *testing.T) {
	NewAccept(&dto.TargetedInfo{Port: 443, IP: "104.16.248.249"}, 1, 14)
}
