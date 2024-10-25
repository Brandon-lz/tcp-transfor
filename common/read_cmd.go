package common

import (
	"bufio"
	"bytes"
	"net"

	"github.com/Brandon-lz/tcp-transfor/utils"
)

func ReadCmd(conn net.Conn) ([]byte, error) {          // 这种连续读取的情况下会丢数据，还是需要用面向对象编程
	buf := bytes.Buffer{}
	rd := bufio.NewReader(conn)
	for {
		d, err := rd.ReadBytes('\n')
		if err != nil {
		}
		buf.Write(d)
		l := len(buf.Bytes())
		if l > 2 && buf.Bytes()[l-2] == ';' && buf.Bytes()[l-3] == ';' {
			return utils.AESDecrypt(buf.Bytes()[:l-3])
		} 
	}
}
