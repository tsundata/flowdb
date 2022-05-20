package network

type IRequest interface {
	GetConnection() IConnection
	GetData() []byte
	GetMessageID() uint32
}

type Request struct {
	conn IConnection
	data IMessage
}

func (r Request) GetConnection() IConnection {
	return r.conn
}

func (r Request) GetData() []byte {
	return r.data.GetData()
}

func (r *Request) GetMessageID() uint32 {
	return r.data.GetID()
}
