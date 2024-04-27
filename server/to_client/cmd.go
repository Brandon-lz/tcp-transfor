package toclient

import (
	"io"
	"net"

	"github.com/Brandon-lz/tcp-transfor/utils"
)

type ServerCmd struct {
	Type string      `json:"type"`
	Data interface{} `json:"data"`
}

func cmdToClientGetNewConn(clientConn net.Conn) error {
	 _,err := clientConn.Write(utils.SerilizeData(ServerCmd{
		Type: "new-conn-request",
		Data: nil,
	}))

	resdata,err := io.ReadAll(clientConn)
	if err!= nil {
		return err
	}

	res := utils.DeSerializeData(resdata,&HelloRecv{})
	if res.Code!= 200 {
		

	return err
}
