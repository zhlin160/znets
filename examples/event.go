package main

import "github.com/zhlin160/znets"

type Event struct {
}

func (e *Event) OnMessage(request znets.IRequest) {
	znets.Log.Info("收到到收据：%s", string(request.GetData()))

	znets.SendToClient(request, request.GetClientId(), []byte("hello,znets"))
}

func (e *Event) OnConnect(conn znets.IConnection, clientId string) {
	znets.Log.Info("有连接请求, clientId：%s", clientId)
}

func (e *Event) OnClose(conn znets.IConnection, clientId string) {
	znets.Log.Info("有连接关闭, clientId：%s", clientId)
}

func (e *Event) OnWorkerStart() {
	znets.Log.Info("进程开始...")
}
