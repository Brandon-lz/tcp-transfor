package translocaltcp

import (
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"strconv"

	"github.com/Brandon-lz/tcp-transfor/client/config"
	"github.com/Brandon-lz/tcp-transfor/utils"
)

func CommunicateToServer() {
	var err error
	// 连接到目标地址
	var serverConn net.Conn
	serverConn, err = net.Dial("tcp", config.Config.Server.Host)
	if err != nil {
		log.Printf("Failed to connect to server: %v\n", utils.WrapErrorLocation(err))
		os.Exit(1)
	}

	sayHelloToServer(serverConn)
	ListenServerCmd(serverConn)
}

type ServerCmd struct {
	Type string    `json:"type"`
	Data interface{} `json:"data"`
}



type HelloMessage struct {
	Client struct {
		Name string `json:"name"`
	} `json:"client"`
	Map []struct {
		LocalPort  int `json:"local-port"`
		ServerPort int `json:"server-port"`
	} `json:"map"`
}

type HelloRecv struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
}

func sayHelloToServer(serverConn net.Conn) {
	var hello HelloMessage

	utils.SerializeData(config.Config, &hello)

	// 发送数据
	_, err := serverConn.Write([]byte(utils.PrintDataAsJson(hello)))
	if err != nil {
		log.Printf("Failed to communicate with server: %v\n", err)
		os.Exit(1)
	}

	// 接收数据
	msgdata, err := io.ReadAll(serverConn)
	if err != nil {
		log.Printf("Failed to communicate with server: %v\n", err)
		os.Exit(1)
	}

	var helloRecv HelloRecv
	utils.SerializeData(msgdata, &helloRecv)

	log.Printf("Received message from server: %s\n", msgdata)
	if helloRecv.Code != 200 {
		log.Printf("Failed to init with server\n")
		os.Exit(1)
	}
}


func ListenServerCmd(serverConn net.Conn){
	for {
		msgData,err := io.ReadAll(serverConn)
		if err != nil {
			log.Printf("Failed to communicate with server: %v\n", utils.WrapErrorLocation(err))
			os.Exit(1)
		}
		
	
		cmd :=utils.SerializeData(msgData, &ServerCmd{})
		switch cmd.Type{
			case "new-conn-request":
				newmsgconfig := utils.SerializeData(cmd.Data, &NewConnCreateRequestMessage{})
				localConn, err := serverSignToCreateNewConnection(serverConn, newmsgconfig)
				if err!=nil{
					continue
				}
				LocalConnSet[newmsgconfig.LocalPort] = localConn
	}
	}
	
}


type NewConnCreateRequestMessage struct {
	LocalPort  int `json:"local-port"`
	ServerPort int `json:"server-port"`
}


type ResponseToServer struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
}


func serverSignToCreateNewConnection(serverConn net.Conn,newConnCreateRequestMessage NewConnCreateRequestMessage) (net.Conn, error) {
	localConn, err := createNewConnToLocalPort(newConnCreateRequestMessage.LocalPort)
	if err != nil {
		resData := utils.PrintDataAsJson(ResponseToServer{Code: 500, Msg: fmt.Sprintf("Failed to create local connection:%s",newConnCreateRequestMessage.LocalPort)})
		serverConn.Write([]byte(resData))
		return nil, utils.WrapErrorLocation(err, fmt.Sprintf("Failed to create local connection %s: %v\n", newConnCreateRequestMessage.LocalPort, err))
	}

	return localConn,nil
}

func createNewConnToLocalPort(localPort int) (net.Conn, error) {
	// 创建本地端口tcp连接
	localTarget := "127.0.0.1:" + strconv.Itoa(localPort)
	localConn, err := net.Dial("tcp", localTarget)
	if err != nil {
		return nil, utils.WrapErrorLocation(err, fmt.Sprintf("Failed to create local connection %s: %v\n", localTarget, err))
	}
	return localConn, nil
}
