package network

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"sync"
)

type IConnection interface {
	Start()
	Stop()

	GetTCPConnection() *net.TCPConn
	GetConnID() uint32
	RemoteAddr() net.Addr

	SendMessage(uint32, []byte) error
	SendBuffMessage(uint32, []byte) error

	SetProperty(string, interface{})
	GetProperty(string) (interface{}, error)
	RemoveProperty(string)
}

type HandFunc func(*net.TCPConn, []byte, int) error

type Connection struct {
	TCPServer IServer
	Conn      *net.TCPConn
	ConnID    uint32
	isClosed  bool

	messageHandler  IMessageHandle
	ExitBuffChan    chan bool
	messageChan     chan []byte
	messageBuffChan chan []byte

	property     map[string]interface{}
	propertyLock *sync.RWMutex
}

func NewConnection(server IServer, conn *net.TCPConn, connID uint32, h IMessageHandle) *Connection {
	c := &Connection{
		TCPServer:       server,
		Conn:            conn,
		ConnID:          connID,
		isClosed:        false,
		messageHandler:  h,
		ExitBuffChan:    make(chan bool, 1),
		messageChan:     make(chan []byte),
		messageBuffChan: make(chan []byte, Setting.MaxMessageChanLen),
		property:        make(map[string]interface{}),
		propertyLock:    &sync.RWMutex{},
	}
	c.TCPServer.GetConnManager().Add(c)
	return c
}

func (c *Connection) StartReader() {
	log.Printf("reader goroutine is running... %d", c.ConnID)
	defer fmt.Println(c.RemoteAddr().String(), " conn reader exit")
	defer c.Stop()

	for {
		pack := NewPack()

		headData := make([]byte, pack.GetHeadLen())
		if _, err := io.ReadFull(c.GetTCPConnection(), headData); err != nil {
			log.Println("read message head error ", err)
			c.ExitBuffChan <- true
			continue
		}

		msg, err := pack.Unpack(headData)
		if err != nil {
			log.Println("unpack error ", err)
			c.ExitBuffChan <- true
			continue
		}

		var data []byte
		if msg.GetDataLen() > 0 {
			data = make([]byte, msg.GetDataLen())
			if _, err := io.ReadFull(c.GetTCPConnection(), data); err != nil {
				log.Println("read message data error ", err)
				c.ExitBuffChan <- true
				continue
			}
		}
		msg.SetData(data)

		req := Request{
			conn: c,
			data: msg,
		}
		if Setting.WorkerPoolSize > 0 {
			c.messageHandler.SendMessageToTaskQueue(&req)
		} else {
			go c.messageHandler.DoMessageHandler(&req)
		}
	}
}

func (c *Connection) StartWriter() {
	log.Println("writer goroutine is running")
	defer log.Println(c.RemoteAddr(), " conn writer exit")

	for {
		select {
		case data := <-c.messageChan:
			if _, err := c.Conn.Write(data); err != nil {
				log.Println("send data error ", err, " conn writer exit")
				return
			}
		case data, ok := <-c.messageBuffChan:
			if ok {
				if _, err := c.Conn.Write(data); err != nil {
					log.Println("send buff data error ", err, " conn writer exit")
					return
				}
			} else {
				log.Println("messageBuffChan is Closed")
				break
			}
		case <-c.ExitBuffChan:
			return
		}
	}
}

func (c *Connection) Start() {
	go c.StartReader()
	go c.StartWriter()

	// hook start
	c.TCPServer.CallOnConnStart(c)

	for {
		select {
		case <-c.ExitBuffChan:
			return
		}
	}
}

func (c *Connection) Stop() {
	if c.isClosed {
		return
	}
	c.isClosed = true

	// hook stop
	c.TCPServer.CallOnConnStop(c)

	err := c.Conn.Close()
	if err != nil {
		log.Println(err)
		return
	}

	c.ExitBuffChan <- true

	c.TCPServer.GetConnManager().Remove(c)

	close(c.ExitBuffChan)
	close(c.messageBuffChan)
}

func (c *Connection) GetTCPConnection() *net.TCPConn {
	return c.Conn
}

func (c *Connection) GetConnID() uint32 {
	return c.ConnID
}

func (c *Connection) RemoteAddr() net.Addr {
	return c.Conn.RemoteAddr()
}

func (c *Connection) SendMessage(id uint32, data []byte) error {
	if c.isClosed {
		return errors.New("connection closed when send message")
	}

	pack := NewPack()
	msg, err := pack.Pack(NewMessage(id, data))
	if err != nil {
		log.Println("pack error message id = ", id)
		return errors.New("pack error message")
	}

	// write
	c.messageChan <- msg

	return nil
}

func (c *Connection) SendBuffMessage(id uint32, data []byte) error {
	if c.isClosed {
		return errors.New("connection closed when send buff message")
	}

	pack := NewPack()
	msg, err := pack.Pack(NewMessage(id, data))
	if err != nil {
		log.Println("pack error message id = ", id)
		return errors.New("pack error message")
	}

	// write
	c.messageBuffChan <- msg

	return nil
}

func (c *Connection) SetProperty(key string, value interface{}) {
	c.propertyLock.Lock()
	defer c.propertyLock.Unlock()

	c.property[key] = value
}

func (c *Connection) GetProperty(key string) (interface{}, error) {
	c.propertyLock.RLock()
	defer c.propertyLock.RUnlock()

	if value, ok := c.property[key]; ok {
		return value, nil
	} else {
		return nil, errors.New("no property found")
	}
}

func (c *Connection) RemoveProperty(key string) {
	c.propertyLock.Lock()
	defer c.propertyLock.Unlock()

	delete(c.property, key)
}
