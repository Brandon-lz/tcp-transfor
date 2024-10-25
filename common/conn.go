package common

import (
	"net"
	"sync"
)

type ConnLocked struct {
	*net.TCPConn
	rlock *sync.Mutex
	wlock *sync.Mutex
}

func NewConnLocked(conn *net.TCPConn) *ConnLocked {
	return &ConnLocked{
		TCPConn: conn,
		rlock: new(sync.Mutex),
		wlock: new(sync.Mutex),
	}
}

func (c *ConnLocked) Read(b []byte) (n int, err error) {
	c.rlock.Lock()
	defer c.rlock.Unlock()
	return c.TCPConn.Read(b)
}

func (c *ConnLocked) Write(b []byte) (n int, err error) {
	c.wlock.Lock()
	defer c.wlock.Unlock()
	return c.TCPConn.Write(b)
}


func CheckConnIsClosed(c net.Conn) bool {
	_, err := c.Write([]byte{})
	if err != nil {
		c.Close()
		return true
	}
	return false
}
