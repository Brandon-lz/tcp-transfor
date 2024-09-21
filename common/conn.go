package common

import (
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

func CheckConnIsClosed(c *net.TCPConn, zero time.Time) bool {
	// c.SetReadDeadline(time.Now())
	// var one = make([]byte, 1)
	// if _, err := c.Read(one); err == io.EOF {
	// 	log.Printf("Client disconnect: %s", c.RemoteAddr())
	// 	c.Close()
	// 	c = nil
	// 	return true
	// } else {
	// 	c.SetReadDeadline(zero)
	// 	return false
	// }
	return false
}
