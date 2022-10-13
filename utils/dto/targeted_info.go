package dto

type TargetedInfo struct {
	Protocol string
	IP       string
	Port     uint16
	AType    byte //0x01表示IPv4地址，DST.ADDR为4个字节
	//0x03表示域名，DST.ADDR是一个可变长度的域名
	//0x04表示IPv6地址，DST.ADDR为16个字节长度
}
