package common

import (
	"io"
	"log"
	"net"
	"time"
)

func ReadConn(clientConn *net.TCPConn) ([]byte, error) {
	buffer := make([]byte, 1024)
	n, err := clientConn.Read(buffer)
	if err != nil {
		log.Println("读取响应失败:", err)
		return []byte{}, err
	}
	return buffer[:n], nil
}

func CheckConnIsClosed(c *net.TCPConn) {
	c.SetReadDeadline(time.Now())
	var one []byte
	if _, err := c.Read(one); err == io.EOF {
		log.Printf("Client disconnect: %s", c.RemoteAddr())
		c.Close()
		c = nil
	} else {
		var zero time.Time
		c.SetReadDeadline(zero)
	}
}
