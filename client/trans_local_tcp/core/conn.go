package core

import (
	"fmt"
	"log"
	"net"
	"strconv"
	"strings"

	"github.com/Brandon-lz/tcp-transfor/client/config"
	"github.com/Brandon-lz/tcp-transfor/common"
	"github.com/Brandon-lz/tcp-transfor/utils"
)

var serverConnReadySignalWithId = make(map[int]chan bool) // connId : readySignal

func CreateNewConnToServer() (net.Conn, error) {
	defer utils.RecoverAndLog(func(err error) {
		utils.PrintDataAsJson("Error occurred in CommunicateToServer" + err.Error())
	})
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

// 创建本地端口tcp连接
func CreateNewConnToLocalPort(localPort int) (net.Conn, error) {
	return CreateNewConn("127.0.0.1", localPort)
}

// 创建与本机能访问到的  host+port的连接
func CreateNewConnToRemotePort(ip string, port int) (net.Conn, error) {
	return CreateNewConn(ip, port)
}

func CreateNewConn(host string, localPort int) (net.Conn, error) {
	// 创建本地端口tcp连接
	tcpAddr, err := net.ResolveTCPAddr("tcp", host+":"+strconv.Itoa(localPort))
	if err != nil {
		return nil, utils.WrapErrorLocation(err, fmt.Sprintf("Failed to resolve local address: %v\n", err))
	}
	// localConn, err := net.Dial("tcp", localTarget)

	newConn, err := net.DialTCP("tcp", nil, tcpAddr)
	if err != nil {
		localTarget := host + ":" + strconv.Itoa(localPort)
		return nil, utils.WrapErrorLocation(err, fmt.Sprintf("Failed to create local connection %s: %v\n", localTarget, err))
	}
	return common.NewConnLocked(newConn), nil
}
