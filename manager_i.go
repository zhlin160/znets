package znets

type IManager interface {
	Add(con IConnection)
	Del(con IConnection)
	Get(id uint32) (IConnection, error)
	Num() int
	Clear()
}
