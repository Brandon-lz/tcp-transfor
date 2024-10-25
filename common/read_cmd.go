package common

import (
	"bufio"
	"bytes"
	"net"
)

func ReadCmd(conn net.Conn) ([]byte, error) {
	buf := bytes.Buffer{}
	rd := bufio.NewReader(conn)
	for {
		d, err := rd.ReadBytes('\n')
		if err != nil {
		}
		buf.Write(d)
		l := len(buf.Bytes())
		if buf.Bytes()[l-2] == '\r' {
			return buf.Bytes()[:l-2],nil
		} 
	}
}
