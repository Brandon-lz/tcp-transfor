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
		if common.CheckConnIsClosed(serverConn) {
			log.Println("with server conn is closed")
			return
		}
		// msgData, err := common.ReadConn(serverConn)        // 这里有bug
		msgData, err := common.ReadCmd(serverConn)
		if err != nil {
			log.Printf("Failed to communicate with server: %v\n", utils.WrapErrorLocation(err))
			// os.Exit(1)
			return
		}

		cmd, err := utils.DeSerializeData(msgData, &common.ServerCmd{})
		if err != nil {
			log.Println("Failed to deserialize server command: ", err)
			continue
		}
		switch cmd.Type {
		case "ping":
			// log.Printf("Received ping message from server: %s\n", msgData)
			// resData, _ := json.Marshal(ResponseToServer{Id: cmd.Id, Code: 200, Msg: "pong"})
			// serverConn.Write(resData)
			// _, err = serverConn.Write([]byte("pong"))
			err = common.SendCmd(serverConn, []byte("pong"))
			if err != nil {
				log.Println("Failed to send pong to server: ", err)
				return
			}

		case "new-conn-request":
			log.Println("Received new connection request from server")
			newcmd, err := utils.DeSerializeData(cmd.Data, &common.NewConnCreateRequestMessage{})
			if err != nil {
				log.Println("failed to deserialize new connection request message: ", err)
				continue
			}
			// _, err = serverConn.Write([]byte("new-conn-request-ack"))         // ack to server
			err = common.SendCmd(serverConn, []byte("new-conn-request-ack")) // ack to server
			if err != nil {
				log.Println("failed to send new-conn-request-ack to server", utils.WrapErrorLocation(err))
				continue
			}
			newServerSubConn, err := CreateNewConnToServer()
			if err != nil {
				log.Println("failed to create new connection to server: ", err)
				continue
			}
			localConn, err := CreateNewConnToLocalPort(newcmd.LocalPort)
			if err != nil {
				// newServerSubConn.Write(utils.SerilizeData(ResponseToServer{Id: cmd.Id, Code: 500, Msg: fmt.Sprintf("Failed to create local connection:%d", newcmd.LocalPort)}))
				common.SendCmd(newServerSubConn, utils.SerilizeData(ResponseToServer{Id: cmd.Id, Code: 500, Msg: fmt.Sprintf("Failed to create local connection:%d", newcmd.LocalPort)}))
				continue
			}
			hello := common.HelloMessage{Type: "sub", ConnId: newcmd.ConnId}
			hello.Client.Name = config.Config.Client.Name
			// _, err = newServerSubConn.Write(utils.SerilizeData(hello)) // hello to server
			err = common.SendCmd(newServerSubConn, utils.SerilizeData(hello)) // hello to server
			if err != nil {
				log.Println("fail to hello to server")
				continue
			}
			// _, err = newServerSubConn.Read(make([]byte, 1024)) // ack
			_, err = common.ReadCmd(newServerSubConn)
			if err != nil {
				log.Println("fail to ack new conn from server")
				continue
			}
			// serverConn.Write(utils.SerilizeData(ResponseToServer{Code: 200, Msg: "New connection created", Data: newcmd.ConnId})) // 是否还需要通知？，可能会降低性能
			serverConnSet[newcmd.ConnId] = localConn
			// go TransForConnData(localConn, newServerSubConn)
			var ready = make(chan bool, 2)
			go common.TransForConnDataClient(localConn, newServerSubConn, &ready)
			<-ready
			// newServerSubConn.Write([]byte("ready"))
			common.SendCmd(newServerSubConn,[]byte("ready"))
			close(ready)
			log.Println("success new sub connection to server", hello.ConnId)
		default:
			log.Println("Unknown command received from server", cmd.Type)
		}
	}

}
