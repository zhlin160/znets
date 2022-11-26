package znets

type Handler struct {
	Middlewares  []HandlerFunc   //中间件集合
	abort        bool            //中间件执行中是否有中断
	workpoolSize uint32          //工作池
	tasks        []chan IRequest //收到请求任务通道

	before      HandlerFunc //前置操作
	after       HandlerFunc //后置操作
	eventHandle IEvent      //操作接收主体

}

func NewHandler() *Handler {
	return &Handler{
		Middlewares:  make([]HandlerFunc, 0),
		abort:        false,
		workpoolSize: 10,
	}
}

//开始处理请求
func (h *Handler) RunHandler(request IRequest) {
	if h.eventHandle == nil {
		Log.Error("You must set IEvent obj")
		return
	}
	//如果有中间件需要执行
	for k := range h.Middlewares {
		h.Middlewares[k](request)
		if h.abort {
			h.abort = false //有中断执行
			return
		}
	}
	h.start(request)
}

//设置前置处理钩子
func (h *Handler) Before(rf HandlerFunc) {
	h.before = rf
	Log.Info("Add Before hook")
}

//设置后置处理钩子
func (h *Handler) After(rf HandlerFunc) {
	h.after = rf
	Log.Info("Add After hook")
}

//设置中间件
func (h *Handler) Use(rf HandlerFunc) {
	h.Middlewares = append(h.Middlewares, rf)
}

//终断请求
func (h *Handler) Abort() {
	h.abort = true
}

//设置工作池数量
func (h *Handler) SetWorkPoolSize(size uint32) {
	h.workpoolSize = size
}

//启动工作池
func (h *Handler) RunWorkPool() {
	if h.eventHandle != nil {
		h.eventHandle.OnWorkerStart()
	}
	h.tasks = make([]chan IRequest, h.workpoolSize)
	for i := 0; i < int(h.workpoolSize); i++ {
		h.tasks[i] = make(chan IRequest)
		go h.runWork(h.tasks[i])
	}
	Log.Info("[WorkPool] %d workpools are Running...", h.workpoolSize)
}

//协程中启动监听请求到来
func (h *Handler) runWork(tr chan IRequest) {
	for {
		select {
		case rq := <-tr:
			h.RunHandler(rq)
			rid := rq.getRid()
			*rid-- //全局请求数-1
		}
	}
}

//轮询获取工作池处理任务
func (h *Handler) SendToTasks(rq IRequest) {
	id := *(rq.getRid()) % h.workpoolSize
	rq.SetWorkId(id)
	h.tasks[id] <- rq
}

//设置事件处理类
func (h *Handler) SetEventHandle(event IEvent) {
	h.eventHandle = event
}

//开始处理接收信息
func (h *Handler) start(request IRequest) {
	if h.before != nil {
		h.before(request)
	}
	h.eventHandle.OnMessage(request)
	if h.after != nil {
		h.after(request)
	}
}
