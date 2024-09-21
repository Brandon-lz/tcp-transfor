package core

import (
	// "encoding/json"
	"fmt"
	"log"
	"net"

	"github.com/Brandon-lz/tcp-transfor/client/config"
	"github.com/Brandon-lz/tcp-transfor/common"
	"github.com/Brandon-lz/tcp-transfor/utils"
)

type ResponseToServer struct {
	Id   int         `json:"id"`
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data"`
}

func ListenServerCmd(serverConn *net.TCPConn) {
	for {

		// msgData, err := io.ReadAll(serverConn)
		msgData, err := common.ReadConn(serverConn)
		if err != nil {
			log.Printf("Failed to communicate with server: %v\n", utils.WrapErrorLocation(err))
			// os.Exit(1)
			return
		}

		cmd := utils.DeSerializeData(msgData, &common.ServerCmd{})
		switch cmd.Type {
		case "ping":
			// log.Printf("Received ping message from server: %s\n", msgData)
			// resData, _ := json.Marshal(ResponseToServer{Id: cmd.Id, Code: 200, Msg: "pong"})
			// serverConn.Write(resData)

		case "new-conn-request":
			log.Println("Received new connection request from server")
			newcmd := utils.DeSerializeData(cmd.Data, &common.NewConnCreateRequestMessage{})
			newServerSubConn, err := CreateNewConnToServer()
			if err != nil {
				continue
			}
			localConn, err := CreateNewConnToLocalPort(newcmd.LocalPort)
			if err != nil {
				newServerSubConn.Write(utils.SerilizeData(ResponseToServer{Id: cmd.Id, Code: 500, Msg: fmt.Sprintf("Failed to create local connection:%d", newcmd.LocalPort)}))
				continue
			}
			hello := common.HelloMessage{Type: "sub", ConnId: newcmd.ConnId}
			hello.Client.Name = config.Config.Client.Name
			newServerSubConn.Write(utils.SerilizeData(hello)) // hello to server
			newServerSubConn.Read(make([]byte, 1024))   // wait for hello response from server
			// serverConn.Write(utils.SerilizeData(ResponseToServer{Code: 200, Msg: "New connection created", Data: newcmd.ConnId})) // 是否还需要通知？，可能会降低性能
			serverConnSet[newcmd.ConnId] = localConn
			// go TransForConnData(localConn, newServerSubConn)
			var ready = make(chan bool, 2)
			go common.TransForConnDataClient(localConn, newServerSubConn, ready)
			<-ready
			newServerSubConn.Write([]byte("ready"))
			close(ready)
			log.Println("success new sub connection to server", hello.ConnId)
		default:
			log.Println("Unknown command received from server", cmd.Type)
		}
	}

}
