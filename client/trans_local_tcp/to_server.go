package translocaltcp

import (
	"io"
	"log"
	"net"
	"os"

	"github.com/Brandon-lz/tcp-transfor/client/config"
	"github.com/Brandon-lz/tcp-transfor/client/trans_local_tcp/core"
	"github.com/Brandon-lz/tcp-transfor/utils"
)

func CommunicateToServer() {
	serverConn, err := core.CreateNewConnToServer()
	if err != nil {
		log.Printf("Failed to create main connection to server: %v\n", utils.WrapErrorLocation(err))
		os.Exit(1)
	}
	sayHelloToServer(serverConn)
	core.ListenServerCmd(serverConn)

}

// type HelloMessage struct {
// 	Type   string `json:"type"`                // main or sub
// 	Client struct {
// 		Name string `json:"name"`
// 	} `json:"client"`
// 	Map []struct {
// 		LocalPort  int `json:"local-port"`
// 		ServerPort int `json:"server-port"`
// 	} `json:"map"`
// }

// type HelloRecv struct {
// 	Code int    `json:"code"`
// 	Msg  string `json:"msg"`
// }

func sayHelloToServer(serverConn net.Conn) {
	var hello core.HelloMessage

	utils.DeSerializeData(config.Config, &hello)
	hello.Type = "main"
	// 发送数据
	_, err := serverConn.Write(utils.SerilizeData(hello))
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

	var helloRecv core.HelloRecv
	utils.DeSerializeData(msgdata, &helloRecv)

	log.Printf("Received message from server: %s\n", msgdata)
	if helloRecv.Code != 200 {
		log.Printf("Failed to init with server\n")
		os.Exit(1)
	}
}
