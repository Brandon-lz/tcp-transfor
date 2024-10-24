package common

import (
	"bufio"
	"bytes"
	"net"

	"github.com/Brandon-lz/tcp-transfor/utils"
)

func ReadCmd(conn *net.TCPConn) ([]byte, error) {
	utils.PrintDataAsJson("ReadCmd strart at: "+utils.GetCodeLine(2))
	defer utils.PrintDataAsJson("ReadCmd end at: "+utils.GetCodeLine(2))
	buf := bytes.Buffer{}
	rd := bufio.NewReader(conn)
	var _data = []byte{}
	for {
		d, err := rd.ReadBytes('\n')
		if err != nil {
			utils.PrintDataAsJson("readBytes error: "+err.Error())
			utils.PrintDataAsJson(len(buf.Bytes()))
			return nil, err
		}
		_, err = buf.Write(d)
		if err != nil {
			return nil, err
		}
		_data = buf.Bytes()
		l := len(_data)
		if l > 3 {
			if _data[l-2] == '\r' && _data[l-3] == '\r' {
				return _data[:l-3], nil
			}
		}

	}
}
