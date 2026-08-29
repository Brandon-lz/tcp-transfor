package translocaltcp

import (
	"errors"
	"log"
	"net"

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

	// 主连接整个生命周期复用同一个 LineReader，
	// 避免握手和命令循环之间因各自新建 bufio 导致预读字节丢失
	lr := common.NewLineReader(serverConn)

	err = sayMainHelloToServer(serverConn, lr) // establish connection with server
	if err != nil {
		log.Printf("Failed to say hello to server: %v\n", utils.WrapErrorLocation(err))
		return
	}
	log.Println("success establish connection with server")

	// go core.KeepAlive(serverConn)
	core.ListenServerCmd(serverConn, lr) //block
}

func sayMainHelloToServer(serverConn net.Conn, lr *common.LineReader) error {
	defer utils.RecoverAndLog(func(err error) {
		log.Printf("Error occurred in CommunicateToServer: %v\n", utils.WrapErrorLocation(err))
	})
	var hello = common.HelloMessage{Type: "main"}
	hello.Client.Name = config.Config.Client.Name
	for i := range config.Config.Map {
		hello.Map = append(hello.Map, struct {
			LocalHost string `json:"local-host"`
			LocalPort  int    `json:"local-port"`
			ServerPort int    `json:"server-port"`
		}{
			LocalHost: config.Config.Map[i].LocalHost,
			LocalPort:  config.Config.Map[i].LocalPort,
			ServerPort: config.Config.Map[i].ServerPort,
		})
	}

	// 发送数据
	if err := common.SendCmd(serverConn, utils.SerilizeData(hello)); err != nil {
		return utils.WrapErrorLocation(err)
	}

	log.Println("Sent hello message to server")
	// 接收数据
	msgdata, err := lr.ReadOne()
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
		return utils.WrapErrorLocation(errors.New(helloRecv.Msg))
	}
	return nil
}
