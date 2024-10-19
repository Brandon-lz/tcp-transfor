package common

import (
	"bufio"
	"bytes"
	"net"
)

func ReadCmd(conn *net.TCPConn) ([]byte, error) {
	buf := bytes.Buffer{}
	for {
		d, err := bufio.NewReader(conn).ReadBytes('\n')
		if err != nil {
		}
		buf.Write(d)
		l := len(buf.Bytes())
		if buf.Bytes()[l-2] == '\r' {
			return buf.Bytes()[:l-2],nil
		} 
	}
}
