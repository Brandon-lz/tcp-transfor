package core

import (
	"fmt"
	"log"
	"net"
	"strconv"
	"strings"

	"github.com/Brandon-lz/tcp-transfor/client/config"
	"github.com/Brandon-lz/tcp-transfor/utils"
)

var serverConnSet = make(map[int]*net.TCPConn)

func CreateNewConnToServer() (*net.TCPConn, error) {
	// var err error
	// // 连接到服务器
	// var serverConn *net.TCPConn
	// serverConn, err = net.Dial("tcp", config.Config.Server.Host)
	// if err != nil {
	// 	log.Printf("Failed to connect to server: %v\n", utils.WrapErrorLocation(err))
	// 	return nil, utils.WrapErrorLocation(err)
	// }
	// log.Println("Connected to server")
	// return serverConn, nil
	_host := strings.Split(config.Config.Server.Host, ":")
	ip := _host[0]
	port, err := strconv.Atoi(_host[1])
	if err != nil {
		log.Println("Failed to parse server port: ", err)
		return nil, err
	}
	return CreateNewConn(ip, port)
}

func CreateNewConnToLocalPort(localPort int) (*net.TCPConn, error) {
	// 创建本地端口tcp连接
	return CreateNewConn("127.0.0.1", localPort)
}

func CreateNewConn(host string, localPort int) (*net.TCPConn, error) {
	// 创建本地端口tcp连接
	tcpAddr, err := net.ResolveTCPAddr("tcp", host+":"+strconv.Itoa(localPort))
	if err != nil {
		return nil, utils.WrapErrorLocation(err, fmt.Sprintf("Failed to resolve local address: %v\n", err))
	}
	// localConn, err := net.Dial("tcp", localTarget)

	localConn, err := net.DialTCP("tcp", nil, tcpAddr)
	if err != nil {
		localTarget := host + ":" + strconv.Itoa(localPort)
		return nil, utils.WrapErrorLocation(err, fmt.Sprintf("Failed to create local connection %s: %v\n", localTarget, err))
	}
	return localConn, nil
}
