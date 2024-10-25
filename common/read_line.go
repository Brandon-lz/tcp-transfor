package common

import (
	"bufio"
	"bytes"
	"net"

	"github.com/Brandon-lz/tcp-transfor/utils"
)

type ConnSocket struct {
	net.Conn
	buf  bytes.Buffer
	rd   *bufio.Reader
}

func NewConnSocket(conn net.Conn) *ConnSocket {
	return &ConnSocket{
		Conn: conn,
		buf:  bytes.Buffer{},
		rd:   bufio.NewReader(conn),
	}
}

func (s *ConnSocket) SendLine(b []byte) (n int, err error) {
	b = append(utils.AESEncrypt(b), []byte(";;\n")...)
	return s.Conn.Write(b)
}

func (s *ConnSocket) RecvLine() (line []byte, err error) {
	var _data = []byte{}
	for {
		d, err := s.rd.ReadBytes('\n')
		if err != nil {
			utils.PrintDataAsJson("readBytes error: " + err.Error() + " " + utils.GetCodeLine(2))
			return nil, err
		}
		_, err = s.buf.Write(d)
		if err != nil {
			utils.PrintDataAsJson("ReadCmd error at: " + utils.GetCodeLine(2))
			return nil, err
		}
		_data = s.buf.Bytes()
		l := len(_data)
		if l > 2 {
			if _data[l-2] == ';' && _data[l-3] == ';' { // 对于结束符是\r\n的情况
				s.buf.Reset()
				return utils.AESDecrypt(_data[:l-3])
			} else {
				// s := string(_data)
				// utils.PrintDataAsJson("1111111111: " + s + " " + utils.GetCodeLine(2))
			}
		}

	}
}

func (s *ConnSocket) Close() error {
	return s.Conn.Close()
}
