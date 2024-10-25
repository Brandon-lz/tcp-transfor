package common

import (
	"net"

	"github.com/Brandon-lz/tcp-transfor/utils"
)

func SendCmd(conn net.Conn, cmd []byte) error {
	cmd = append(utils.AESEncrypt(cmd), []byte(";;\n")...)
	_, err := conn.Write(cmd)
	return err
}
