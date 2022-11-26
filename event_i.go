package znets

type IEvent interface {
	OnMessage(IRequest)
	OnConnect(IConnection, string) //连接对象, clientId
	OnClose(IConnection, string)
	OnWorkerStart()
}
