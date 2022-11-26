package znets

type Request struct {
	conn     IConnection //已建立的连接
	msg      IMessage    //客户端请求的数据
	rid      *uint32     //当前Request的ID
	workId   uint32      //工作池内标识id
	clientId string      //客户端连接记录标识id
}

//实例化,rid:全局请求id
func NewRequest(con IConnection, msg IMessage, rid *uint32, clientId string) IRequest {
	return &Request{
		conn:     con,
		msg:      msg,
		rid:      rid,
		clientId: clientId,
	}
}

//获取当前连接
func (r *Request) GetConnection() IConnection {
	return r.conn
}

//获取消息数据
func (r *Request) GetData() []byte {
	return r.msg.GetData()
}

//获取ID
func (r *Request) GetID() uint32 {
	return r.msg.GetId()
}

//获取数据长度
func (r *Request) GetLen() uint32 {
	return r.msg.GetLen()
}

//获取全局请求序号id
func (r *Request) getRid() *uint32 {
	return r.rid
}

//设置工作池workId
func (r *Request) SetWorkId(workId uint32) {
	r.workId = workId
}

//获取工作池workId
func (r *Request) GetWorkId() uint32 {
	return r.workId
}

//获取客户端连接封装id
func (r *Request) GetClientId() string {
	return r.clientId
}
