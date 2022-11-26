package znets

type Message struct {
	Id     uint32
	Length uint32
	Data   []byte
}

func (msg *Message) SetData(data []byte) {
	msg.Data = data
}

func (msg *Message) GetData() []byte {
	return msg.Data
}

func (msg *Message) SetId(id uint32) {
	msg.Id = id
}

func (msg *Message) GetId() uint32 {
	return msg.Id
}

func (msg *Message) SetLen(len uint32) {
	msg.Length = len
}

func (msg *Message) GetLen() uint32 {
	return msg.Length
}
