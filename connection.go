package znets

import (
	"bytes"
	"errors"
	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/transform"
	"io/ioutil"
	"net"
	"sync"
)

type HandleFunc func(*net.TCPConn, []byte, int) error

type Connection struct {
	//连接的套接字
	Conn *net.TCPConn
	//连接的ID
	ConnID uint32
	//连接状态
	isClosed bool
	//写通道退出状态的channel
	ExitChan chan bool
	//当前连接的处理方法
	Handles IHandler
	//无缓冲的读写channel
	dataChan chan []byte
	//父server
	server IServer
	//保护属性
	property map[string]interface{}
	//保护锁
	propertyLock sync.RWMutex

	packProto IPack           //协议解析
	connWg    *sync.WaitGroup //进程中协程连接同步等待，用于在需要结束进程时等待处理未完成连接
}

func NewConnection(server IServer, conn *net.TCPConn, id uint32, handler IHandler, wg *sync.WaitGroup) IConnection {
	c := &Connection{
		Conn:     conn,
		ConnID:   id,
		isClosed: false,
		ExitChan: make(chan bool, 1),
		Handles:  handler,
		dataChan: make(chan []byte),
		server:   server,
		property: make(map[string]interface{}),

		connWg: wg,
	}

	c.server.GetManager().Add(c)
	return c
}

func (c *Connection) SetPackProto(proto IPack) {
	c.packProto = proto
}

//主动踢掉连接
func (c *Connection) closeConn() {
	err := c.Conn.SetLinger(-1)
	if err != nil {
		Log.Error("Error when setting linger:%s", err.Error())
		return
	}

	//<-time.After(2 * time.Second)
	c.Stop()
}

//发送数据处理
func (c *Connection) StartWriter() {
	for {
		select {
		case data := <-c.dataChan:
			//收到nil则关闭掉连接
			if data == nil {
				c.closeConn()
				return
			}
			//读写channel有数据时
			if _, err := c.Conn.Write(data); err != nil {
				Log.Error("Send data err:%s", err.Error())
				return
			}

		case <-c.ExitChan:
			return
		}
	}
}

//收到数据处理
func (c *Connection) StartReader() {
	defer c.Stop()

	var recvBuff string
	var currentPackageLength int

	for {
		var buff [65535]byte
		n, err := c.Conn.Read(buff[:])
		if err != nil {
			Log.Error("read msg data err:%s", err.Error())
			break
		}
		gBbuff, _ := GbToUtf8(buff[:n]) //通讯中有中文简单处理
		recvBuff += string(gBbuff)
		recvBuffLen := len(recvBuff)

		if c.packProto != nil {
			for recvBuff != "" {
				var message string
				currentPackageLength = c.packProto.Input(recvBuff) //解析数据输入分割
				if currentPackageLength > 0 {
					if currentPackageLength == recvBuffLen {
						message = recvBuff
						recvBuff = ""
						recvBuffLen = 0
					} else {
						message = recvBuff[:currentPackageLength]
						recvBuff = recvBuff[currentPackageLength:]
						recvBuffLen -= currentPackageLength
					}

					message = string(c.packProto.UnPack([]byte(message)))
					msg := &Message{
						Data:   []byte(message),
						Length: uint32(currentPackageLength),
					}
					currentPackageLength = 0

					//调用通知处理
					rid := c.server.GetRid()
					clientId := AddressToClientId(c)
					req := NewRequest(c, msg, rid, clientId)
					*(rid)++
					c.Handles.SendToTasks(req)
				} else {
					break
				}
			}
			continue
		}

		if recvBuff == "" {
			continue
		}

		msg := &Message{
			Data:   []byte(recvBuff),
			Length: uint32(len(recvBuff)),
		}
		recvBuff = ""
		recvBuffLen = 0

		//调用通知处理
		rid := c.server.GetRid()
		clientId := AddressToClientId(c)
		req := NewRequest(c, msg, rid, clientId)
		*(rid)++
		c.Handles.SendToTasks(req)
	}
}

//启动连接
func (c *Connection) Start() {
	Log.Info("connection coming in, ConnID = %d, Addr = %s", c.ConnID, c.GetConn().RemoteAddr().String())
	//启动读数据业务
	go c.StartReader()
	// 启动写数据业务
	go c.StartWriter()

	go c.server.runOnStart(c)
}

//关闭连接
func (c *Connection) Stop() {
	Log.Info("connection close, ConnID = %d, Addr = %s", c.ConnID, c.GetConn().RemoteAddr().String())

	if c.isClosed == true {
		return
	}
	c.isClosed = true
	c.server.runOnStop(c)
	c.Conn.Close()
	c.ExitChan <- true
	close(c.ExitChan)
	close(c.dataChan)

	c.connWg.Done() //连接wg -1
	c.server.GetManager().Del(c)
}

//获取当前连接绑定的conn
func (c *Connection) GetConn() *net.TCPConn {
	return c.Conn
}

//获取当前连接的id
func (c *Connection) GetID() uint32 {
	return c.ConnID
}

//获取远程客户端信息
func (c *Connection) RemoteAddr() net.Addr {
	return c.Conn.RemoteAddr()
}

//发送数据
func (c *Connection) Send(data []byte) error {
	if c.isClosed {
		return errors.New("Connection closes")
	}

	if c.packProto != nil {
		data = c.packProto.Pack(data)
	}
	data, _ = Utf8ToGb(data) //处理中文
	c.dataChan <- data
	return nil
}

//获取主sever
func (c *Connection) GetServer() IServer {
	return c.server
}

//添加当前连接额外kye-value
func (c *Connection) SetProperty(key string, val interface{}) {
	c.propertyLock.Lock()
	defer c.propertyLock.Unlock()

	c.property[key] = val
}
func (c *Connection) GetProperty(key string) (interface{}, error) {
	c.propertyLock.RLock()
	defer c.propertyLock.RUnlock()

	if val, ok := c.property[key]; ok {
		return val, nil
	}
	return nil, errors.New("No Property")
}
func (c *Connection) DelProperty(key string) {
	c.propertyLock.Lock()
	defer c.propertyLock.Unlock()

	delete(c.property, key)
}

func (c *Connection) SetProtoPack(proto IPack) {
	c.packProto = proto
}

func GbToUtf8(s []byte) ([]byte, error) {
	reader := transform.NewReader(bytes.NewReader(s), simplifiedchinese.GBK.NewDecoder())
	d, e := ioutil.ReadAll(reader)
	if e != nil {
		return nil, e
	}
	return d, nil
}

func Utf8ToGb(s []byte) ([]byte, error) {
	reader := transform.NewReader(bytes.NewReader(s), simplifiedchinese.GBK.NewEncoder())
	d, e := ioutil.ReadAll(reader)
	if e != nil {
		return nil, e
	}
	return d, nil
}
