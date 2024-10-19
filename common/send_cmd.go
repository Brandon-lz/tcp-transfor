package common

import "net"

func SendCmd(conn *net.TCPConn, cmd []byte) error {
	_, err := conn.Write(cmd)
	if err != nil {
		return err
	}
	_, err = conn.Write([]byte("\r\n"))
	if err != nil {
		return err
	}
	return nil
}

