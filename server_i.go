package znets

import (
	"net"
)

type hookHandler func(c IConnection)
type overloadHandler func(c *net.TCPConn)

type IServer interface {
	Run()
	start()
	Stop()

	Before(HandlerFunc)

	After(HandlerFunc)
	Use(HandlerFunc)
	Abort()

	SetWorkPoolSize(uint32)
	GetRid() *uint32

	GetManager() IManager
	OverLoad(overloadHandler)
	SetMaxCon(uint32)

	OnStart(hookHandler)
	runOnStart(IConnection)

	OnStop(hookHandler)
	runOnStop(IConnection)

	SetEventHandle(IEvent)
}
