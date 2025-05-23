package translocaltcp

import (
	// "fmt"
	"errors"
	"log"
	"net"

	// "os"

	"github.com/Brandon-lz/tcp-transfor/client/config"
	"github.com/Brandon-lz/tcp-transfor/client/trans_local_tcp/core"
	"github.com/Brandon-lz/tcp-transfor/common"
	"github.com/Brandon-lz/tcp-transfor/utils"
)

func CommunicateToServer() {
	defer utils.RecoverAndLog(func(err error) {
		log.Printf("Error occurred in CommunicateToServer: %v\n", utils.WrapErrorLocation(err))
	})
	serverConn, err := core.CreateNewConnToServer()
	if err != nil {
		log.Printf("Failed to create main connection to server: %v\n", utils.WrapErrorLocation(err))
		return
	}
	defer serverConn.Close()

	err = sayMainHelloToServer(serverConn) // establish connection with server
	if err != nil {
		log.Printf("Failed to say hello to server: %v\n", utils.WrapErrorLocation(err))
		return
	}
	log.Println("success establish connection with server")

	// go core.KeepAlive(serverConn)
	core.ListenServerCmd(serverConn) //block

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

func sayMainHelloToServer(serverConn net.Conn) error {
	defer utils.RecoverAndLog(func(err error) {
		log.Printf("Error occurred in CommunicateToServer: %v\n", utils.WrapErrorLocation(err))
	})
	var hello = common.HelloMessage{Type: "main"}

	_, err := utils.DeSerializeData(config.Config, &hello)
	if err != nil {
		return utils.WrapErrorLocation(err)
	}
	// 发送数据
	// _, err = serverConn.Write(utils.SerilizeData(hello))
	err = common.SendCmd(serverConn, utils.SerilizeData(hello))
	if err != nil {
		return utils.WrapErrorLocation(err)
	}

	log.Println("Sent hello message to server")
	// 接收数据

	// msgdata, err := common.ReadConn(serverConn)
	msgdata,err := common.ReadCmd(serverConn)
	if err != nil {
		return utils.WrapErrorLocation(err)
	}

	var helloRecv common.HelloRecv
	_, err = utils.DeSerializeData(msgdata, &helloRecv)
	if err != nil {
		return utils.WrapErrorLocation(err)
	}

	log.Printf("Received message from server: %s\n", msgdata)
	if helloRecv.Code != 200 {
		// log.Printf("Failed to init with server\n")
		// os.Exit(1)
		return utils.WrapErrorLocation(errors.New(helloRecv.Msg))
	}
	return nil
}
