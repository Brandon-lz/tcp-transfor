package toclient

import (
	"fmt"
	"io"
	"log"
	"net"

	"github.com/Brandon-lz/tcp-transfor/server/config"
	"github.com/Brandon-lz/tcp-transfor/utils"
)



// listen to accecpt client main conn and sub conn
func ListenClientConn() {
	listener,err := net.Listen("tcp", fmt.Sprintf(":%d",config.Config.Port))
	if err != nil {
		log.Fatalf("Failed to listen on %d: %v", config.Config.Port, err)
	}
	defer listener.Close()

	for {
		conn, err := listener.Accept()
		if err!= nil {
			log.Printf("Failed to accept connection: %v", err)
			continue
		}
		go startCmdToClient(conn)
	}
}



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

func startCmdToClient(clientConn net.Conn){
	defer utils.RecoverAndLog()
	defer clientConn.Close()
	// clientName := clientConn.RemoteAddr().String()
	hellodata,err := io.ReadAll(clientConn)
	if err!= nil {
		log.Printf("Failed to read hello message from client: %v", err)
		return
	}

	hello := utils.DeSerializeData(hellodata, &HelloMessage{})

	if err := AddNewClient(&Client{Name: hello.Client.Name, Conn: clientConn, Map: hello.Map});err!=nil{
		clientConn.Write(utils.SerilizeData(HelloRecv{Code:500,Msg:fmt.Sprintf("client hello faild:%v",err)}))
		return
	}else{
		clientConn.Write(utils.SerilizeData(HelloRecv{Code:200,Msg:"hello success"}))
	}
	

	for _,m := range hello.Map {
		go newListenerClientMapPort(hello,m.ServerPort)
	}

}


func newListenerClientMapPort(hello HelloMessage,listenPort int) error {
	defer utils.RecoverAndLog()
	listener,err := net.Listen("tcp", fmt.Sprintf(":%d",listenPort))
	if err != nil {
		return fmt.Errorf("Failed to listen on %d: %v", listenPort, err)
	}
	defer listener.Close()

	for {
		userConn, err := listener.Accept()
		if err!= nil {
			log.Printf("Failed to accept connection: %v", err)
			continue
		}
		// new conn to server

		// cmd to client to get a new conn with client
		client,err := getClientByName(hello.Client.Name)
		if err!= nil {
			log.Printf("Failed to get client conn: %v", utils.WrapErrorLocation(err, "getClientByName"))
			return err
		}

		if err:=cmdToClientGetNewConn(client.Conn);err!=nil{
			log.Printf("Failed to get new conn to client: %v", utils.WrapErrorLocation(err, "cmdToClientGetNewConn"))
			return err
		}



	}
}


