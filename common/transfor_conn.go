package common

import (
	"fmt"
	"net"
	"time"

	"github.com/Brandon-lz/tcp-transfor/utils"
)

// 1-> 4   3->2

func TransForConnDataServer(user2serverConn net.Conn, server2clientConn net.Conn) {
	defer utils.RecoverAndLog()
	defer user2serverConn.Close()
	defer server2clientConn.Close()

	// user -> server
	go func() {
		defer utils.RecoverAndLog()

		defer user2serverConn.Close()
		defer server2clientConn.Close()

		utils.PrintDataAsJson("1------------1-")
		err := copyWithTimeout(server2clientConn, user2serverConn, 8*time.Hour)
		if err != nil {
			utils.PrintDataAsJson(fmt.Sprintf("1Copy error: %v, bytes copied: %d", err))
		}
	}()

	// server -> user
	utils.PrintDataAsJson("2------------1-")
	err := copyWithTimeout(user2serverConn, server2clientConn, 8*time.Hour)
	if err != nil {
		utils.PrintDataAsJson(fmt.Sprintf("2Copy error: %v, bytes copied: %d", err))
	}
}

func TransForConnDataClient(local2clientConn net.Conn, client2serverConn net.Conn) {
	defer utils.RecoverAndLog()
	defer local2clientConn.Close()
	defer client2serverConn.Close()

	// fmt.Println(<-ready)

	// quit := make(chan bool)
	go func() {
		// defer utils.RecoverAndLog(func(err error) {
		// 	// quit <- true
		// })
		defer local2clientConn.Close()
		defer client2serverConn.Close()

		// local -> server
		utils.PrintDataAsJson("3------------1-")
		err := copyWithTimeout(client2serverConn, local2clientConn, 8*time.Hour)
		if err != nil {
			utils.PrintDataAsJson(fmt.Sprintf("3Copy error: %v, bytes copied: %d", err))
		}

	}()

	// server -> local

	utils.PrintDataAsJson("4------------1-")
	err := copyWithTimeout(local2clientConn, client2serverConn,8*time.Hour)
	if err != nil {
		utils.PrintDataAsJson(fmt.Sprintf("4Copy error: %v, bytes copied: %d", err))
	}

}

func copyWithTimeout(dst, src net.Conn, timeout time.Duration) error {
	buf := make([]byte, 4096) // 4KB缓冲区
	for {
		src.SetReadDeadline(time.Now().Add(timeout)) // 每次读操作前设置超时
		n, err := src.Read(buf)
		if err != nil {
			return err // 读取超时或连接关闭
		}

		dst.SetWriteDeadline(time.Now().Add(timeout)) // 每次写操作前设置超时
		_, err = dst.Write(buf[:n])
		if err != nil {
			return err // 写入超时或连接关闭
		}
	}
}
