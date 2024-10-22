package core

import (
	"net"
	"time"

	"github.com/Brandon-lz/tcp-transfor/utils"
)

func KeepAlive(serverConn *net.TCPConn) {
	defer utils.RecoverAndLog()
	for {
		time.Sleep(time.Second * 5)
		// ping := common.ServerCmd{
		// 	Type: "ping",
		// }
		// // _, err := serverConn.Write(utils.SerilizeData(ping))
		// err := common.SendCmd(serverConn, utils.SerilizeData(ping))
		// if err != nil {
		// 	panic(err)
		// }
		_,err := serverConn.Write([]byte{})
		if err!=nil{
			utils.PrintDataAsJson(err)
			serverConn.Close()
		}
	}
}
