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
        sourceConn, err := listener.Accept()
        if err != nil {
            log.Printf("Failed to accept connection: %v", err)
            continue
        }

        // 连接到目标地址
        targetConn, err := net.Dial("tcp", targetAddr)
        if err != nil {
            log.Printf("Failed to connect to target: %v", err)
            sourceConn.Close()
            continue
        }

        // 启动goroutine来处理数据转发
        go forward(sourceConn, targetConn)
    }
}

func forward(sourceConn, targetConn net.Conn) {
    // 将源连接的数据复制到目标连接
    for {
		sourceConn.Read()
       
    }
}
