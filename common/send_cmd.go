package common

import "net"

func SendCmd(conn net.Conn, cmd []byte) error {
	cmd = append(cmd, []byte("\r\n")...)
	_, err := conn.Write(cmd)
	return err
}
