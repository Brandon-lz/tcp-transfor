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

	user2serverConnSocket := TcpSocket{TCPConn: user2serverConn,_buf:bytes.Buffer{}}
	server2clientConnSocket := TcpSocket{TCPConn: server2clientConn,_buf:bytes.Buffer{}}

	go func() {
		defer utils.RecoverAndLog()

		defer user2serverConn.Close()
		defer server2clientConn.Close()

		// user -> server
		// user2serverConn.SetDeadline(time.Now().Add(200 * time.Second))
		for {
			// n, err := user2serverConn.Read(readBuff)
			d,err := user2serverConnSocket.ReadLine()
			if err != nil {
				if err == io.EOF {
					return
				}
				log.Println("receive data from user error, close conn", err.Error(), utils.GetCodeLine(1))
				break
			}
		

			// log.Println("从用户接收到并传送给客户端的数据:", string(readBuff[:n]))

			

			// if err != nil {
			// 	log.Println("failed to decrypt data from user:", err)
			// 	return
			// }
			// _, err = server2clientConn.Write([]byte(_data))
			_, err = server2clientConn.Write(utils.AESEncrypt(d))
			if err != nil {
				log.Println("failed to write data to server:", err)
				return
			}
		}

	}()

	// server -> user
	for {

		// err := user2serverConn.SetDeadline(time.Now().Add(200 * time.Second))
		// if err != nil {
		// 	panic(fmt.Errorf("failed to set deadline for %s: %v", user2serverConn.RemoteAddr(), utils.WrapErrorLocation(err)))
		// }
		// err = server2clientConn.SetDeadline(time.Now().Add(200 * time.Second))
		// if err != nil {
		// 	panic(fmt.Errorf("failed to set deadline for %s: %v", server2clientConn.RemoteAddr(), utils.WrapErrorLocation(err)))
		// }

		readbuffer := make([]byte, 1024)
		for {
			// n, err := server2clientConn.Read(readbuffer)
			server2clientConnSocket
			if err != nil {
				if err == io.EOF {
					return
				}
				log.Println("receive data from client error, close conn", err, utils.GetCodeLine(1))
				return
			}
			if n == 0 {
				if CheckConnIsClosed(server2clientConn) {
					log.Println("receive empty data from client, close conn")
					return
				}
				continue
			}

			// log.Println("从客户端接收到并传送给用户的数据:", string(readbuffer[:n]))

			// _data, err := utils.AESDecryptWithKey(string(readbuffer[:n]))
			// if err != nil {
			// 	log.Println("failed to decrypt data from server:", err)
			// 	return
			// }
			// _, err = user2serverConn.Write(_data)
			_, err = user2serverConn.Write(readbuffer[:n])
			if err != nil {
				log.Println("failed to write data to user:", err)
				return
			}
		}

	}
}

func TransForConnDataClient(local2clientConn *net.TCPConn, client2serverConn *net.TCPConn) {
	defer utils.RecoverAndLog()
	defer local2clientConn.Close()
	defer client2serverConn.Close()

	// fmt.Println(<-ready)

	// quit := make(chan bool)
	go func() {
		// defer utils.RecoverAndLog(func(err error) {
		// 	// quit <- true
		// })
		defer local2clientConn.Close()
		defer client2serverConn.Close()

		// local -> server
		readBuff := make([]byte, 1024)
		for {
			n, err := local2clientConn.Read(readBuff)
			if err != nil {
				if err == io.EOF {
					return
				}
				log.Println("receive data from local error, close conn", err, utils.GetCodeLine(1))
				return
			}
			if n == 0 {
				if CheckConnIsClosed(local2clientConn) {
					log.Println("receive empty data from local, close conn")
					return
				}
				continue
			}
			// log.Println("从本地接收到并传送给服务器的数据:", string(readBuff[:n]))
			// _data, err := utils.AESEncryptWithKey(readBuff[:n])
			// if err != nil {
			// 	log.Println("failed to encrypt data from local:", err)
			// 	return
			// }
			// _, err = client2serverConn.Write([]byte(_data))
			_, err = client2serverConn.Write(readBuff[:n])
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

	readbuffer := make([]byte, 1024)
	for {
		n, err := client2serverConn.Read(readbuffer)
		if err != nil {
			if err == io.EOF {
				return
			}
			log.Println("receive data from server error, close conn", err, utils.GetCodeLine(1))
			return
		}
		if n == 0 {
			if CheckConnIsClosed(client2serverConn) {
				log.Println("receive empty data from server, close conn")
				return
			}
			continue
		}
		// log.Println("从服务器接收到并传送给本地的数据:", string(readbuffer[:n]))
		// _data, err := utils.AESDecryptWithKey(string(readbuffer[:n]))
		// if err != nil {
		// 	log.Println("failed to decrypt data from server:", err)
		// 	return
		// }
		// _, err = local2clientConn.Write(_data)
		_, err = local2clientConn.Write(readbuffer[:n])
		if err != nil {
			log.Println("failed to write data to local:", err)
			return
		}
	}

}
