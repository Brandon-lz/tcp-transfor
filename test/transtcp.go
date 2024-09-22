package main

import (
	"fmt"
	"io"
	"log"
	"net"
)

func main() {
	// 源地址和目标地址
	sourceAddr := "localhost:8080"
	targetAddr := "localhost:9090"

	// 监听源地址
	listener, err := net.Listen("tcp", sourceAddr)
	if err != nil {
		log.Fatalf("Failed to listen on %s: %v", sourceAddr, err)
	}
	defer listener.Close()

	fmt.Printf("Forwarding from %s to %s\n", sourceAddr, targetAddr)

	for {
		// 接受来自源地址的连接
		var sourceConn net.Conn
		sourceConn, err = listener.Accept()
		if err != nil {
			log.Printf("Failed to accept connection: %v", err)
			continue
		}

		// 连接到目标地址
		var targetConn net.Conn
		targetConn, err = net.Dial("tcp", targetAddr)
		if err != nil {
			log.Printf("Failed to connect to target: %v", err)
			sourceConn.Close()
			continue
		}

		// 启动goroutine来处理数据转发
		go transData(sourceConn, targetConn)
	}
}

func transData(sourceConn, targetConn net.Conn) {
	defer sourceConn.Close()
	defer targetConn.Close()

	quit := make(chan bool, 2)

	go func() {
		// 将源连接的数据复制到目标连接
		for {
			if _, err := io.Copy(sourceConn, targetConn); err != nil {
				log.Printf("Failed to copy data from source to target: %v", err)
				break
			}

		}
		quit <- true
	}()

	for {
		select {
		case <-quit:
			return
		default:
			// 将目标连接的数据复制到源连接
			if _, err := io.Copy(targetConn, sourceConn); err != nil {
				log.Printf("Failed to copy data from target to source: %v", err)
				break
			}
		}
	}
}
