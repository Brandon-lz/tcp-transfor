package toclient

import (
	"log"
	"net"

	"github.com/Brandon-lz/tcp-transfor/common"
	"github.com/Brandon-lz/tcp-transfor/utils"
)

func cmdToClientGetNewConn(clientConn net.Conn, connId, LocalPort, ServerPort int) error {
	_, err := clientConn.Write(utils.SerilizeData(common.ServerCmd{
		Type: "new-conn-request",
		Data: common.NewConnCreateRequestMessage{
			ConnId:     connId,
			LocalPort:  LocalPort,
			ServerPort: ServerPort,
		},
	}))
	log.Println("send new conn request to client")
	if err != nil {
		return err
	}
	return nil

	// resdata, err := io.ReadAll(clientConn)
	// if err != nil {
	// 	return err
	// }

	// res := utils.DeSerializeData(resdata, &common.HelloRecv{})
	// if res.Code != 200 {

	// 	return err
	// }
}
