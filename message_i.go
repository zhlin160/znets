package znets

type IMessage interface {
	GetData() []byte
	GetId() uint32

	SetData([]byte)
	SetId(uint32)

	SetLen(uint32)
	GetLen() uint32
}
