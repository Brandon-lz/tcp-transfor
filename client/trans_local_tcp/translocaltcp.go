package translocaltcp

import (
	"io"
	"log"
	"net"

	"github.com/Brandon-lz/tcp-transfor/utils"
)

func Start() {

	// 源地址和目标地址
	targetAddr := "localhost:8080"
	sourceAddr := "0.0.0.0:9090"

	// 监听源地址
	listener, err := net.Listen("tcp", sourceAddr)
	if err != nil {
		log.Fatalf("Failed to listen on %s: %v", sourceAddr, err)
	}
	defer listener.Close()

	log.Printf("Forwarding from %s to %s\n", sourceAddr, targetAddr)

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
		log.Println("Starting data forwarding")

		// 如何检测连接断开？？
		go transData(sourceConn, targetConn)
	}
}

func transData(sourceConn, targetConn net.Conn) {
	defer utils.RecoverAndLog()
	defer sourceConn.Close()
	defer targetConn.Close()

	quit := make(chan bool)
	quit2 := make(chan bool)

	// check conn disconnect  不行
	// go func(){
	// 	for {
	// 		if _, err := sourceConn.Write([]byte{}) ; err != nil {
	// 			log.Printf("Failed to write to source: %v", err)
	// 			quit <- true
	// 		}
	// 		time.Sleep(time.Second * 2)
	// 	}
	// }()

	go func() {
		defer utils.RecoverAndLog(func(err error){quit2 <- true})
		// 将源连接的数据复制到目标连接
		for {
			if _, err := io.Copy(sourceConn, targetConn); err != nil {
				log.Printf("Failed to copy data from source to target: %v", err)
				break
			}
		}
		
	}()

BackWard:
	for {
		select {
		case <-quit:
			return
		case <-quit2:
			return
		default:
			// 将目标连接的数据复制到源连接
			if _, err := io.Copy(targetConn, sourceConn); err != nil {
				log.Printf("Failed to copy data from target to source: %v", err)
				break BackWard
			}
		}
	}
	log.Println("Data forwarding stopped")
}
