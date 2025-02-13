package common

import (
	"bufio"
	"bytes"
	"net"
)

func ReadCmd(conn net.Conn) ([]byte, error) { // 这种连续读取的情况下会丢数据，还是需要用面向对象编程

	buf := bytes.Buffer{}
	rd := bufio.NewReader(conn)
	// conn.SetReadDeadline(time.Now().Add(2*time.Second))
	for {
		d, err := rd.ReadBytes('\n')
		if err != nil {
			return nil, err
		}
		buf.Write(d)
		l := len(buf.Bytes())
		if l > 3 && buf.Bytes()[l-2] == ';' && buf.Bytes()[l-3] == ';' && buf.Bytes()[l-4] == ';' {
			return buf.Bytes()[:l-4],nil
		}
	}
}
