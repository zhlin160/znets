# znets
基于Golang简单轻量级TCP并发服务器框架

***
### 安装
```go
go get github.com/zhlin160/znets
```

### 快速开始
* 创建服务句柄
```go
znets.NewServer() //默认方式创建句柄
znets.NewServerWithOptions(options *Options) //以options方式传参创建
```
```go
type Options struct {
	IP             string 
	Port           int
	Model          string  //运行模式 dev|production
	MaxConnNum     uint32  //最大连接数
	WorkPool       uint32  //工作池大小
	PidFilePath    string  //pid文件保存路径，默认启动目录
}
```
* 设置消息响应回调对象，只需实现IEvent接口
```go
type IEvent interface {
	OnMessage(IRequest)            //收到消息
	OnConnect(IConnection, string) //有连接接入，参数：连接对象, clientId
	OnClose(IConnection, string)   //有连接关闭
	OnWorkerStart()                // 服务启动回调
}
```
* 设置协议解析,需实现IPack 接口
```go
type IPack interface {
	Input(string) int      //返回长度位置
	Pack([]byte) []byte    //打包
	UnPack([]byte) []byte  //解包
}
```
* 启动服务
```go
func (s *Server) Run()
```
> 程序执行启动参数 start：启动， restart：优雅重启，stop：优雅停止

### IRequest方法
```go
type IRequest interface {
    GetConnection() IConnection  //获取连接对象
    GetData() []byte             //获取数据
    GetWorkId() uint32           //获取工作池工作id
    GetClientId() string         //获取客户端连接id,封装的地址及连接信息字符串
}
```
### Server全局方法
```go
//给某个连接发送消息
func SendToClient(request IRequest, clientId string, data []byte) error 

//关闭某个连接
func CloseClient(request IRequest, clientId string, data []byte) error

//是否在线
func IsOnLine(request IRequest, clientId string) bool
```

### Example
***
```go
package main

import "github.com/zhlin160/znets"

func main() {
	//srv := znets.NewServer()
	srv := znets.NewServerWithOptions(&znets.Options{
		IP:       "127.0.0.1",
		Port:     9898,
		WorkPool: 15,
	})

	srv.SetEventHandle(&Event{})
	srv.SetProtoPack(NewMessagePack())

	srv.Run()
}
```