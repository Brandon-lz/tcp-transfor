package common

import "net"

// ConnSocket 兼容旧接口的行读取/写入封装，内部复用同一个带缓冲 LineReader
type ConnSocket struct {
	net.Conn
	lr *LineReader
}

func NewConnSocket(conn net.Conn) *ConnSocket {
	return &ConnSocket{Conn: conn, lr: NewLineReader(conn)}
}

// NewConnSocketWithReader 复用外部 LineReader（同一连接上多阶段读时必用）
func NewConnSocketWithReader(conn net.Conn, lr *LineReader) *ConnSocket {
	return &ConnSocket{Conn: conn, lr: lr}
}

func (s *ConnSocket) SendLine(b []byte) (n int, err error) {
	err = SendCmd(s.Conn, b)
	n = len(b) + 4
	return
}

// RecvLine 读取一条以 ";;;\n" 结尾的消息
func (s *ConnSocket) RecvLine() (line []byte, err error) {
	return s.lr.ReadOne()
}

func (s *ConnSocket) Close() error {
	return s.Conn.Close()
}
