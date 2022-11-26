package znets

type HandlerFunc func(request IRequest)

type IHandler interface {
	RunHandler(request IRequest)
	Before(HandlerFunc)
	After(HandlerFunc)
	Use(HandlerFunc)
	Abort()
	RunWorkPool()
	SendToTasks(rq IRequest)
	SetWorkPoolSize(size uint32)

	SetEventHandle(IEvent)
}
