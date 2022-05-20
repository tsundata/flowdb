package network

import (
	"bytes"
	"encoding/binary"
	"errors"
)

type IPack interface {
	GetHeadLen() uint32
	Pack(IMessage) ([]byte, error)
	Unpack([]byte) (IMessage, error)
}

type Pack struct {
}

func NewPack() *Pack {
	return &Pack{}
}

// ID (4 bytes) + DataLen (4 bytes)
func (p Pack) GetHeadLen() uint32 {
	return 8
}

func (p Pack) Pack(message IMessage) ([]byte, error) {
	buf := bytes.NewBuffer([]byte{})

	if err := binary.Write(buf, binary.LittleEndian, message.GetDataLen()); err != nil {
		return nil, err
	}

	if err := binary.Write(buf, binary.LittleEndian, message.GetID()); err != nil {
		return nil, err
	}

	if err := binary.Write(buf, binary.LittleEndian, message.GetData()); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func (p Pack) Unpack(data []byte) (IMessage, error) {
	buf := bytes.NewBuffer(data)

	m := &Message{}

	if err := binary.Read(buf, binary.LittleEndian, &m.DataLen); err != nil {
		return nil, err
	}

	if err := binary.Read(buf, binary.LittleEndian, &m.ID); err != nil {
		return nil, err
	}

	if Setting.MaxPacketSize > 0 && m.DataLen > Setting.MaxPacketSize {
		return nil, errors.New("too large message data receive")
	}

	return m, nil
}
