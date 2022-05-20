package network

type IMessage interface {
	GetID() uint32
	GetData() []byte
	GetDataLen() uint32

	SetID(uint32)
	SetData([]byte)
	SetDataLen(uint32)
}

type Message struct {
	ID      uint32
	DataLen uint32
	Data    []byte
}

func NewMessage(ID uint32, data []byte) *Message {
	return &Message{ID: ID, Data: data, DataLen: uint32(len(data))}
}

func (m *Message) GetID() uint32 {
	return m.ID
}

func (m *Message) GetData() []byte {
	return m.Data
}

func (m *Message) GetDataLen() uint32 {
	return m.DataLen
}

func (m *Message) SetID(id uint32) {
	m.ID = id
}

func (m *Message) SetData(data []byte) {
	m.Data = data
}

func (m *Message) SetDataLen(l uint32) {
	m.DataLen = l
}
