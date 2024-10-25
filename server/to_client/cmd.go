package toclient

import (
	"log"

	"github.com/Brandon-lz/tcp-transfor/common"
	"github.com/Brandon-lz/tcp-transfor/utils"
)

func cmdToClientGetNewConn(ccm *clientConnManager, connId, LocalPort, ServerPort int) error {
	ccm.Cmdrwlock.Lock()
	defer ccm.Cmdrwlock.Unlock()
	sercmd := common.ServerCmd{
		Type: "new-conn-request",
		Data: common.NewConnCreateRequestMessage{
			ConnId:     connId,
			LocalPort:  LocalPort,
			ServerPort: ServerPort,
		},
	}
	// _, err := clientConn.Write(utils.SerilizeData(sercmd))
	err := common.SendCmd(ccm.ClientConn, utils.SerilizeData(sercmd))
	log.Println("send new conn request to client", utils.PrintDataAsJson(sercmd))
	if err != nil {
		return utils.WrapErrorLocation(err, "cmdToClientGetNewConn")
	}

	// ack
	// _, err = clientConn.Read(make([]byte, 1024))
	_, err = common.ReadCmd(ccm.ClientConn)
	if err != nil {
		log.Println("read ack error", utils.WrapErrorLocation(err))
		return err
	}

	return nil

}
