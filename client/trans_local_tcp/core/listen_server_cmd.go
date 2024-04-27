package core

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"os"

	"github.com/Brandon-lz/tcp-transfor/client/config"
	"github.com/Brandon-lz/tcp-transfor/utils"
)


type HelloMessage struct {
	Type   string `json:"type"`                // main or sub
	Client struct {
		Name string `json:"name"`
	} `json:"client"`
	Map []struct {
		LocalPort  int `json:"local-port"`
		ServerPort int `json:"server-port"`
	} `json:"map"`

	ConnId int `json:"conn-id"` // 服务端-本客户端之间有多个连接，每个连接都有唯一的conn-id，拿着conn-id返回给服务端去注册新连接
}

type HelloRecv struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
}

type ServerCmd struct {
	Type string      `json:"type"`
	Data interface{} `json:"data"`
}

type ResponseToServer struct {
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data"`
}

type NewConnCreateRequestMessage struct {
	ConnId     int `json:"conn-id"` // 服务端-本客户端之间有多个连接，每个连接都有唯一的conn-id，拿着conn-id返回给服务端去注册新连接
	LocalPort  int `json:"local-port"`
	ServerPort int `json:"server-port"`
}

func ListenServerCmd(serverConn net.Conn) {
	for {
		msgData, err := io.ReadAll(serverConn)
		if err != nil {
			log.Printf("Failed to communicate with server: %v\n", utils.WrapErrorLocation(err))
			os.Exit(1)
		}

		cmd := utils.DeSerializeData(msgData, &ServerCmd{})
		switch cmd.Type {
		case "ping":
			log.Printf("Received ping message from server: %s\n", msgData)
			resData, _ := json.Marshal(ResponseToServer{Code: 200, Msg: "pong"})
			serverConn.Write(resData)
		case "new-conn-request":
			newcmd := utils.DeSerializeData(cmd.Data, &NewConnCreateRequestMessage{})
			serverSubConn, err := CreateNewConnToServer()
			if err != nil {
				continue
			}
			localConn, err := CreateNewConnToLocalPort(newcmd.LocalPort)
			if err != nil {
				serverSubConn.Write(utils.SerilizeData(ResponseToServer{Code: 500, Msg: fmt.Sprintf("Failed to create local connection:%d", newcmd.LocalPort)}))
				continue
			}
			hm := HelloMessage{Type:"sub",ConnId: newcmd.ConnId}
			hm.Client.Name = config.Config.Client.Name
			serverSubConn.Write(utils.SerilizeData(hm))
			serverConn.Write(utils.SerilizeData(ResponseToServer{Code: 200, Msg: "New connection created", Data: newcmd.ConnId}))
			serverConnSet[newcmd.ConnId] = localConn
			TransForConnData(localConn, serverSubConn)
		}
	}

}