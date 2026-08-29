package common

import (
	"bufio"
	"bytes"
	"net"
)

var sep = []byte(";;;\n")

// LineReader 带缓冲的行读取器，保证同一条连接上多次读取不会丢数据
// （旧版 ReadCmd 每次都新建 bufio.Reader，上一轮读取残留在缓冲区里的字节会被直接丢弃）
type LineReader struct {
	conn net.Conn
	rd   *bufio.Reader
	buf  bytes.Buffer
}

func NewLineReader(conn net.Conn) *LineReader {
	return &LineReader{conn: conn, rd: bufio.NewReaderSize(conn, 64*1024)}
}

// ReadOne 读取一条以 ";;;\n" 结尾的消息，返回消息体（不含分隔符）
func (lr *LineReader) ReadOne() ([]byte, error) {
	lr.buf.Reset()
	for {
		d, err := lr.rd.ReadBytes('\n')
		if err != nil {
			return nil, err
		}
		if _, err := lr.buf.Write(d); err != nil {
			return nil, err
		}
		b := lr.buf.Bytes()
		l := len(b)
		if l >= 4 && bytes.HasSuffix(b, sep) {
			return append([]byte(nil), b[:l-4]...), nil
		}
	}
}

// ReadCmd 一次性读取场景用（内部新建 Reader）；同一连接要反复读取请复用 LineReader
func ReadCmd(conn net.Conn) ([]byte, error) {
	return NewLineReader(conn).ReadOne()
}
