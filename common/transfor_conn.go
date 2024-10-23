package common

import (
	"bufio"
	"bytes"
	"io"
	"log"
	"net"

	"github.com/Brandon-lz/tcp-transfor/utils"
)

type TcpSocket struct {
	*net.TCPConn
	_buf bytes.Buffer
}

func (ts *TcpSocket) SendBytes(b []byte) error {
	b = append(b, []byte("\r\n")...)
	_, err := ts.Write(b)
	return err
}

func (ts *TcpSocket) ReadLine() ([]byte, error) {
	for {
		d, err := bufio.NewReader(ts.TCPConn).ReadBytes('\n')
		if err != nil {
			return nil, err
		}
		ts._buf.Write(d)
		l := len(ts._buf.Bytes())
		if ts._buf.Bytes()[l-2] == '\r' {
			return ts._buf.Bytes()[:l-2], nil
		}
	}
}

func TransForConnDataServer(user2serverConn *net.TCPConn, server2clientConn *net.TCPConn) {
	defer utils.RecoverAndLog()
	defer user2serverConn.Close()
	defer server2clientConn.Close()

	user2serverConnSocket := TcpSocket{TCPConn: user2serverConn, _buf: bytes.Buffer{}}
	server2clientConnSocket := TcpSocket{TCPConn: server2clientConn, _buf: bytes.Buffer{}}

	go func() {
		defer utils.RecoverAndLog()

		defer user2serverConn.Close()
		defer server2clientConn.Close()

		var d = make([]byte,1024)
		for {
			n, err := user2serverConnSocket.Read(d)
			if err != nil {
				if err == io.EOF {
					return
				}
				log.Println("receive data from user error, close conn", err.Error(), utils.GetCodeLine(1))
				break
			}
			if n == 0{

			}

			log.Println("从用户接收到并传送给客户端的数据:", string(d))

			// if err != nil {
			// 	log.Println("failed to decrypt data from user:", err)
			// 	return
			// }
			// _, err = server2clientConn.Write([]byte(_data))
			err = server2clientConnSocket.SendBytes(utils.AESEncrypt(d))
			if err != nil {
				log.Println("failed to write data to server:", err)
				return
			}
		}

	}()

	// server -> user

	// err := user2serverConn.SetDeadline(time.Now().Add(200 * time.Second))
	// if err != nil {
	// 	panic(fmt.Errorf("failed to set deadline for %s: %v", user2serverConn.RemoteAddr(), utils.WrapErrorLocation(err)))
	// }
	// err = server2clientConn.SetDeadline(time.Now().Add(200 * time.Second))
	// if err != nil {
	// 	panic(fmt.Errorf("failed to set deadline for %s: %v", server2clientConn.RemoteAddr(), utils.WrapErrorLocation(err)))
	// }

	for {
		d, err := server2clientConnSocket.ReadLine()
		if err != nil {
			if err == io.EOF {
				return
			}
			log.Println("receive data from client error, close conn", err, utils.GetCodeLine(1))
			return
		}

		// _data, err := utils.AESDecryptWithKey(string(readbuffer[:n]))
		// if err != nil {
		// 	log.Println("failed to decrypt data from server:", err)
		// 	return
		// }
		d, err = utils.AESDecrypt(d)
		if err != nil {
			log.Println("failed to decrypt data from server:", err)
			continue
		}
		log.Println("从客户端接收到并传送给用户的数据:", string(d))
		err = user2serverConnSocket.SendBytes(d)
		if err != nil {
			log.Println("failed to write data to user:", err)
			return
		}
	}
}

func TransForConnDataClient(local2clientConn *net.TCPConn, client2serverConn *net.TCPConn) {
	defer utils.RecoverAndLog()
	defer local2clientConn.Close()
	defer client2serverConn.Close()

	local2clientConnSocket := TcpSocket{TCPConn: local2clientConn, _buf: bytes.Buffer{}}
	client2serverConnSocket := TcpSocket{TCPConn: client2serverConn, _buf: bytes.Buffer{}}

	go func() {
		defer local2clientConn.Close()
		defer client2serverConn.Close()

		for {
			d, err := local2clientConnSocket.ReadLine()
			if err != nil {
				if err == io.EOF {
					return
				}
				log.Println("receive data from local error, close conn", err, utils.GetCodeLine(1))
				return
			}

			// _data, err := utils.AESEncryptWithKey(readBuff[:n])
			// if err != nil {
			// 	log.Println("failed to encrypt data from local:", err)
			// 	return
			// }
			// _, err = client2serverConn.Write([]byte(_data))
			log.Println("从本地接收到并传送给服务器的数据:", string(d))

			err = client2serverConnSocket.SendBytes(utils.AESEncrypt(d))
			if err != nil {
				log.Println("failed to write data to server:", err)
				return
			}
		}

	}()

	// server -> local

	// err := local2clientConn.SetDeadline(time.Now().Add(80 * time.Second))
	// if err != nil {
	// 	panic(fmt.Errorf("failed to set deadline for %s: %v", local2clientConn.RemoteAddr(), utils.WrapErrorLocation(err)))
	// }
	// err = client2serverConn.SetDeadline(time.Now().Add(80 * time.Second))
	// if err != nil {
	// 	panic(fmt.Errorf("failed to set deadline for %s: %v", client2serverConn.RemoteAddr(), utils.WrapErrorLocation(err)))
	// }

	// if count < 9600 {
	// 	src.SetReadDeadline(time.Now().Add(time.Duration(count*60) * time.Second))
	// } else {
	// 	src.SetDeadline(time.Now().Add(8 * time.Hour))
	// }

	for {
		d, err := client2serverConnSocket.ReadLine()
		if err != nil {
			if err == io.EOF {
				return
			}
			log.Println("receive data from server error, close conn", err, utils.GetCodeLine(1))
			return
		}

		// _data, err := utils.AESDecryptWithKey(string(readbuffer[:n]))
		// if err != nil {
		// 	log.Println("failed to decrypt data from server:", err)
		// 	return
		// }
		// _, err = local2clientConn.Write(_data)
		d, err = utils.AESDecrypt(d)
		if err != nil {
			log.Println("failed to decrypt data from server:", err)
			return
		}

		log.Println("从服务器接收到并传送给本地的数据:", string(d))

		err = local2clientConnSocket.SendBytes(d)
		if err != nil {
			log.Println("failed to write data to local:", err)
			return
		}
	}
}
