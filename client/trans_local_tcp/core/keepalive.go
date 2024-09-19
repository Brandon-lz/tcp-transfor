package core

import (
	"net"
	"time"

	"github.com/Brandon-lz/tcp-transfor/common"
	"github.com/Brandon-lz/tcp-transfor/utils"
)

func KeepAlive(serverConn *net.TCPConn) {
	defer utils.RecoverAndLog()
	for {
		time.Sleep(time.Second * 5)
		ping := common.ServerCmd{
			Type: "ping",
		}
		_, err := serverConn.Write(utils.SerilizeData(ping))
		if err != nil {
			panic(err)
		}
	}
}
