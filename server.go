package znets

import (
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/spf13/cast"
	"github.com/spf13/viper"
)

const (
	DEV        = "dev"
	PRODUCTION = "production"
)

type Options struct {
	IP             string
	Port           int
	MaxConnections uint32
	Model          string //运行模式 dev|production
	MaxConnNum     uint32
	WorkPool       uint32
	PidFilePath    string //pid保存路径
}

type Server struct {
	IP             string
	Port           int
	Conn           *Listener
	IPVersion      string
	Handles        *Handler
	Version        string
	cid            uint32  //当前连接数
	rids           *uint32 //当前请求数
	maxConnections uint32  //最大连接数
	manager        IManager
	overload       overloadHandler
	onStart        hookHandler
	onStop         hookHandler

	protoPack IPack

	config   *viper.Viper //配置文件对象
	runModel string       //运行模式 dev|production

	pidFilePath string //pid保存路径

	isExit bool //循环监听中是否需要退出
}

var version string = "v1.0.2"
var Log *HLog

//通过配置文件构建默认server
func NewServer() *Server {
	config := parseConfigFile()
	ip := config.GetString("Server.Ip")
	port := config.GetInt("Server.Port")
	model := config.GetString("Server.Model")
	maxConnNum := config.GetUint32("Server.MaxConnNum")
	workPool := config.GetUint32("Server.WorkPoll")
	pidFilePath := config.GetString("Server.PidFilePath")

	return buildServ(ip, port, model, maxConnNum, workPool, pidFilePath, config)
}

//使用Options字段构建server
func NewServerWithOptions(options *Options) *Server {
	ip := options.IP
	port := options.Port
	model := options.Model
	maxConnNum := options.MaxConnNum
	workPool := options.WorkPool
	pidFilePath := options.PidFilePath

	return buildServ(ip, port, model, maxConnNum, workPool, pidFilePath, nil)
}

func buildServ(ip string, port int, model string, maxConnNum, workPool uint32, pidFilePath string, config *viper.Viper) *Server {
	if ip == "" {
		ip = "0.0.0.0"
	}
	if port == 0 {
		port = 9503
	}
	if model == "" {
		model = DEV
	}
	if maxConnNum == 0 {
		maxConnNum = 10240
	}
	if workPool == 0 {
		workPool = 10
	}

	if len(pidFilePath) == 0 {
		path, _ := os.Getwd()
		pidFilePath = path + "/pid"
	}
	Log = NewLogWithModel(model)
	s := &Server{
		IP:             ip,
		Port:           port,
		IPVersion:      "tcp4",
		Version:        version,
		cid:            0,
		rids:           new(uint32),
		maxConnections: maxConnNum,
		Handles:        NewHandler(),
		manager:        NewManager(),
		runModel:       model,
		pidFilePath:    pidFilePath,
		isExit:         false,
	}
	if config != nil {
		s.SetConfig(config)
	}
	s.SetWorkPoolSize(workPool)
	return s
}

//解析配置文件
func parseConfigFile() *viper.Viper {
	path, _ := os.Getwd()
	config := viper.New()
	config.SetConfigName("config")
	config.SetConfigType("yaml")
	config.AddConfigPath(path)

	if err := config.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			panic("配置文件 " + path + "/config.yaml 未找到")
		} else {
			panic(err)
		}
	}
	return config
}

func (s *Server) SetConfig(conf *viper.Viper) {
	s.config = conf
}

func (s *Server) start() {
	if s.Handles.eventHandle == nil {
		Log.Error("You must set eventHandel")
		return
	}

	op, err := s.checkOp()
	if err != nil {
		Log.Error(err.Error())
		return
	}
	//不是开始服务到这一步就结束了
	if op != "start" {
		return
	}

	//获取TCP地址
	addr, err := net.ResolveTCPAddr(s.IPVersion, fmt.Sprintf("%s:%d", s.IP, s.Port))
	if err != nil {
		Log.Error("Resolve tcp addr err:" + err.Error())
		return
	}

	//监听服务器地址
	s.Conn, err = ListenTCP(s.IPVersion, addr, s)
	if err != nil {
		Log.Error("Listen tcp err:" + err.Error())
		return
	}

	s.writePid(os.Getpid()) //写入进程id

	//监听成功输出
	Log.Info("Start server success..., listen on %s:%d", s.IP, s.Port)
	//开启工作池
	s.Handles.RunWorkPool()
	//循环接受用户连接
	for {
		if s.isExit {
			s.Conn.Close()
			break
		}
		con, err := s.Conn.Accept()
		if err != nil {
			Log.Error("Accept err:%s", err)
			if strings.Contains(err.Error(), " use of closed network connection") {
				Log.Info("连接已关闭")
				break
			}
			continue
		}

		if s.manager.Num() >= int(s.maxConnections) {
			if s.overload != nil {
				s.overload(con)
			}
			con.Close()
			continue
		}

		dealCon := NewConnection(s, con, s.cid, s.Handles, s.Conn.wg)
		dealCon.SetProtoPack(s.protoPack)
		s.cid++
		go dealCon.Start()
	}

}

//运行服务器
func (s *Server) Run() {
	s.start()
}

//停止服务器
func (s *Server) Stop() {
	s.Conn.Close()
	s.manager.Clear()
}

//添加全局中间件
func (s *Server) Use(rf HandlerFunc) {
	s.Handles.Use(rf)
}

//终断请求
func (s *Server) Abort() {
	s.Handles.Abort()
}

//设置前置处理钩子
func (s *Server) Before(rf HandlerFunc) {
	s.Handles.Before(rf)
}

//设置后置处理钩子
func (s *Server) After(rf HandlerFunc) {
	s.Handles.After(rf)
}

//设置工作池大小
func (s *Server) SetWorkPoolSize(size uint32) {
	s.Handles.SetWorkPoolSize(size)
}

func (s *Server) GetManager() IManager {
	return s.manager
}

func (s *Server) GetRid() *uint32 {
	return s.rids
}

func (s *Server) OverLoad(o overloadHandler) {
	s.overload = o
}

func (s *Server) SetMaxCon(size uint32) {
	s.maxConnections = size
}

//连接创建hook
func (s *Server) OnStart(hook hookHandler) {
	s.onStart = hook
}
func (s *Server) runOnStart(c IConnection) {
	if s.onStart != nil {
		s.onStart(c)
	}
	if s.Handles.eventHandle.OnConnect != nil {
		s.Handles.eventHandle.OnConnect(c, AddressToClientId(c))
	}
}

//连接断开hook
func (s *Server) OnStop(c hookHandler) {
	s.onStop = c
}
func (s *Server) runOnStop(c IConnection) {
	if s.onStop != nil {
		s.onStop(c)
	}
	if s.Handles.eventHandle.OnClose != nil {
		s.Handles.eventHandle.OnClose(c, AddressToClientId(c))
	}
}

func (s *Server) SetEventHandle(event IEvent) {
	s.Handles.SetEventHandle(event)
}

func (s *Server) SetProtoPack(proto IPack) {
	s.protoPack = proto
}

//返回配置对象
func (s Server) GetConfig() *viper.Viper {
	return s.config
}

//写入pid文件
func (s *Server) writePid(pid int) {
	fmt.Println("路径：" + s.pidFilePath)
	ioutil.WriteFile(s.pidFilePath, []byte(strconv.Itoa(pid)), 0777)
}

//获取pid文件
func (s *Server) getPid() (int, error) {
	res, err := ioutil.ReadFile(s.pidFilePath)
	if err != nil {
		//Log.Info("获取pid文件内容失败：%s", err.Error())
		return 0, err
	}
	pid, _ := strconv.Atoi(string(res))
	return pid, nil
}

//检查启动命令 options
func (s *Server) checkOp() (string, error) {
	args := os.Args

	if len(args) < 2 {
		return "", fmt.Errorf("must set up operation parameters start | restart | stop")
	}

	op := args[1]

	if op != "start" && op != "restart" && op != "stop" {
		return "", fmt.Errorf("the operation parameters can be to start | restart | stop")
	}

	pid, _ := s.getPid()
	switch op {
	case "start":
		isGracefulEvn := "IS_GRACEFUL=1"
		isGraceful := false

		for _, value := range os.Environ() {
			if value == isGracefulEvn {
				isGraceful = true
			}
		}
		if pid != 0 && !isGraceful {
			return "", fmt.Errorf("the Server is running")
		}

		return "start", nil

	case "restart":
		if pid == 0 {
			return "", fmt.Errorf("the Server is not running")
		}

		err := syscall.Kill(pid, syscall.SIGUSR2)
		if err != nil {
			return "", fmt.Errorf("restart server err:%s", err.Error())
		}
		return "reload", nil

	case "stop":
		if pid == 0 {
			Log.Error("The Server is not running!!!")
			return "", fmt.Errorf("the Server is not running")
		}
		err := syscall.Kill(pid, syscall.SIGTERM)
		if err != nil {
			Log.Error("stop server err:%s", err.Error())
			return "", fmt.Errorf("stop server err:%s", err.Error())
		}
		return "stop", nil
	}
	return "", nil
}

//连接端封装程clientId
func AddressToClientId(connection IConnection) string {
	address := connection.GetConn().RemoteAddr().String()
	str := hex.EncodeToString([]byte(address + ":" + strconv.Itoa(int(connection.GetID()))))
	return str
}

//通过clientId 解析成连接对象
func clientIdToAddress(clientId string) (ip string, port uint32, connId uint32) {
	hex_data, _ := hex.DecodeString(clientId)
	connInfo := strings.Split(string(hex_data), ":")
	ip = connInfo[0]
	port = cast.ToUint32(connInfo[1])
	connId = cast.ToUint32(connInfo[2])
	return
}

//给给定的客户端发送消息
func SendToClient(request IRequest, clientId string, data []byte) error {
	_, _, connId := clientIdToAddress(clientId)
	c, err := request.GetConnection().GetServer().GetManager().Get(connId)
	if err != nil {
		Log.Error("获取链接信息失败:%s", err.Error())
		return err
	}
	return c.Send(data)
}

//踢掉一个连接并发送消息
func CloseClient(request IRequest, clientId string, data []byte) error {
	var connId uint32
	if request.GetClientId() == clientId {
		_, _, connId = clientIdToAddress(request.GetClientId())
	} else {
		_, _, connId = clientIdToAddress(clientId)
	}
	c, err := request.GetConnection().GetServer().GetManager().Get(connId)
	if err != nil {
		Log.Error("获取链接信息失败:%s", err.Error())
		return err
	}
	err = c.Send(data)
	if err != nil {
		Log.Error("发送消息失败:%s", err.Error())
		return err
	}

	time.Sleep(2 * time.Second)
	return c.Send(nil)
}

//是否有连接记录，是否在线
func IsOnLine(request IRequest, clientId string) bool {
	_, _, connId := clientIdToAddress(clientId)
	_, err := request.GetConnection().GetServer().GetManager().Get(connId)
	if err != nil {
		Log.Error("查询是否在线记录失败:%s", err.Error())
		return false
	}
	return true
}
