package znets

type IRequest interface {
	GetConnection() IConnection
	GetData() []byte

	GetID() uint32
	GetLen() uint32
	getRid() *uint32
	SetWorkId(uint32)

	GetWorkId() uint32

	//获取客户端连接id,封装的地址及连接信息字符串
	GetClientId() string
}
