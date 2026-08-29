package common

import "net"

// SendCmd 以 ";;;\n" 作为消息分隔符写出一条命令
func SendCmd(conn net.Conn, cmd []byte) error {
	out := make([]byte, len(cmd)+4)
	copy(out, cmd)
	copy(out[len(cmd):], sep)
	_, err := conn.Write(out)
	return err
}
