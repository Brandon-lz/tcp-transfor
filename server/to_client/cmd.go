package toclient

import (
	"log"
	"net"
	"sync"

	"github.com/Brandon-lz/tcp-transfor/common"
	"github.com/Brandon-lz/tcp-transfor/utils"
)

var lock = &sync.Mutex{}

func cmdToClientGetNewConn(clientConn *net.TCPConn, connId, LocalPort, ServerPort int) error {
	lock.Lock()
	defer lock.Unlock()
	sercmd := common.ServerCmd{
		Type: "new-conn-request",
		Data: common.NewConnCreateRequestMessage{
			ConnId:     connId,
			LocalPort:  LocalPort,
			ServerPort: ServerPort,
		},
	}
	_, err := clientConn.Write(utils.SerilizeData(sercmd))
	log.Println("send new conn request to client", utils.PrintDataAsJson(sercmd))
	if err != nil {
		return utils.WrapErrorLocation(err, "cmdToClientGetNewConn")
	}

	// ack
	_, err = clientConn.Read(make([]byte, 1024))
	if err != nil {
		log.Println("read ack error", utils.WrapErrorLocation(err))
		return err
	}

	return nil

}
