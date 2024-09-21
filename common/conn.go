package common

import (
	"log"
	"net"
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

func CheckConnIsClosed(c *net.TCPConn) bool {
	_, err := c.Write([]byte{})
	if err != nil {
		c.Close()
		return true
	}
	return false
}
