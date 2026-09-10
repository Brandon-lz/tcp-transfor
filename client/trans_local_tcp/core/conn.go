package core

import (
	"fmt"
	"log"
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/Brandon-lz/tcp-transfor/client/config"
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

// CreateNewConnToRemotePortTimeout 建一条到 host:port 的连接，带连接超时。
// 目标无响应时（例如 IP 被静默丢包）net 的默认拨号超时可能长达两分钟，
// 而调用方必须在有限时间内得出结论，所以这里显式设上限。
func CreateNewConnToRemotePortTimeout(ip string, port int, timeout time.Duration) (net.Conn, error) {
	addr := net.JoinHostPort(ip, strconv.Itoa(port))
	conn, err := net.DialTimeout("tcp", addr, timeout)
	if err != nil {
		return nil, utils.WrapErrorLocation(err, fmt.Sprintf("Failed to connect %s within %s", addr, timeout))
	}
	return conn, nil
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
	return newConn, nil
}
