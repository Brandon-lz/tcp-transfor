package core

import (
	"fmt"
	"log"
	"net"
	"strconv"

	"github.com/Brandon-lz/tcp-transfor/client/config"
	"github.com/Brandon-lz/tcp-transfor/utils"
)

var serverConnSet = make(map[int]net.Conn)

func CreateNewConnToServer() (net.Conn, error) {
	var err error
	// 连接到服务器
	var serverConn net.Conn
	serverConn, err = net.Dial("tcp", config.Config.Server.Host)
	if err != nil {
		log.Printf("Failed to connect to server: %v\n", utils.WrapErrorLocation(err))
		return nil, utils.WrapErrorLocation(err)
	}
	log.Println("Connected to server")
	return serverConn, nil
}

func CreateNewConnToLocalPort(localPort int) (net.Conn, error) {
	// 创建本地端口tcp连接
	localTarget := "127.0.0.1:" + strconv.Itoa(localPort)
	localConn, err := net.Dial("tcp", localTarget)
	if err != nil {
		return nil, utils.WrapErrorLocation(err, fmt.Sprintf("Failed to create local connection %s: %v\n", localTarget, err))
	}
	return localConn, nil
}
