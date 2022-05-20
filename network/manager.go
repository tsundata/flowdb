package network

import (
	"errors"
	"log"
	"sync"
)

type IConnManager interface {
	Add(IConnection)
	Remove(IConnection)
	Get(uint32) (IConnection, error)
	Len() int
	ClearConn()
}

type ConnManager struct {
	connections map[uint32]IConnection
	connLock    *sync.RWMutex
}

func NewConnManager() *ConnManager {
	return &ConnManager{
		connLock:    &sync.RWMutex{},
		connections: make(map[uint32]IConnection),
	}
}

func (c *ConnManager) Add(conn IConnection) {
	c.connLock.Lock()
	defer c.connLock.Unlock()

	c.connections[conn.GetConnID()] = conn
	log.Println("conn add to connManager successfully: conn num = ", c.Len())
}

func (c *ConnManager) Remove(conn IConnection) {
	c.connLock.Lock()
	defer c.connLock.Unlock()

	delete(c.connections, conn.GetConnID())
	log.Println("conn remove connID = ", conn.GetConnID(), " successfully: conn num = ", c.Len())
}

func (c *ConnManager) Get(id uint32) (IConnection, error) {
	c.connLock.RLock()
	defer c.connLock.RUnlock()

	if conn, ok := c.connections[id]; ok {
		return conn, nil
	} else {
		return nil, errors.New("connection not found")
	}
}

func (c *ConnManager) Len() int {
	return len(c.connections)
}

func (c *ConnManager) ClearConn() {
	c.connLock.Lock()
	defer c.connLock.Unlock()

	for connID, conn := range c.connections {
		conn.Stop()
		delete(c.connections, connID)
	}
	log.Println("clear all connections successfully: conn num = ", c.Len())
}
