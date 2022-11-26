package znets

import "net"

type IConnection interface {
	Start()
	Stop()
	GetConn() *net.TCPConn
	GetID() uint32
	RemoteAddr() net.Addr
	Send(data []byte) error
	SetProperty(key string, val interface{})
	GetProperty(key string) (interface{}, error)
	DelProperty(key string)

	SetProtoPack(IPack)
	GetServer() IServer
}
