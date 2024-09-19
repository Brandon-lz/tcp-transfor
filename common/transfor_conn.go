package common

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net"
	"time"

	"github.com/Brandon-lz/tcp-transfor/utils"
)

// func TransForConnData(src *net.TCPConn, dst *net.TCPConn) {
// 	defer utils.RecoverAndLog()
// 	defer src.Close()
// 	defer dst.Close()

// 	quit := make(chan bool)
// 	go func() {
// 		defer utils.RecoverAndLog(func(err error) { quit <- true })
// 		for {
// 			_, err := io.Copy(dst, src)
// 			if err != nil {
// 				panic(fmt.Errorf("Failed to copy data from %s to %s: %v\n", src.RemoteAddr(), dst.RemoteAddr(), utils.WrapErrorLocation(err)))
// 			}
// 		}
// 	}()

// trans:
// 	for {
// 		select {
// 		case <-quit:
// 			break trans
// 		default:
// 			_, err := io.Copy(src, dst)
// 			if err != nil {
// 				panic(fmt.Errorf("Failed to copy data from %s to %s: %v\n", dst.RemoteAddr(), src.RemoteAddr(), utils.WrapErrorLocation(err)))
// 			}
// 		}
// 	}

// }

func TransForConnDataServer(user2serverConn *net.TCPConn, server2clientConn *net.TCPConn) {
	defer utils.RecoverAndLog()
	defer user2serverConn.Close()
	defer server2clientConn.Close()

	// quit := make(chan bool)
	go func() {
		defer utils.RecoverAndLog(func(err error) {
			// quit <- true
		})

		// user2serverConn -> server2clientConn
		for {
			_buf := bytes.Buffer{}
			for{
				readBuff := make([]byte, 1024)
				n, err := user2serverConn.Read(readBuff)
				if err != nil {
					if err == io.EOF {
						fmt.Println("receive empty data from user, close conn")
						return
					}
					// panic(fmt.Errorf("failed to read data from %s: %v", user2serverConn.RemoteAddr(), utils.WrapErrorLocation(err)))
					log.Println("failed to read data from user:", err)
					return
				}
				if n == 0 {
					fmt.Println("receive empty data from user, close conn")
					return
				}
				fmt.Println("receive data from user:", string(readBuff[:n]))
				server2clientConn.Write(readBuff[:n])
			}
			
			_, err := io.Copy(&_buf, user2serverConn) // when dst.close, it will panic
			if err != nil {
				panic(fmt.Errorf("failed to read data from %s: %v", user2serverConn.RemoteAddr(), utils.WrapErrorLocation(err)))
			}
			
			time.Sleep(1 * time.Second)
			_data := _buf.Bytes()
			aesdata, err := utils.AESEncryptWithKey(_data)
			if err != nil {
				panic(fmt.Errorf("failed to encrypt data: %v", utils.WrapErrorLocation(err)))
			}

			_, err = io.Copy(server2clientConn, bytes.NewReader([]byte(aesdata))) // when dst.close, it will panic

			if err != nil {
				panic(fmt.Errorf("failed to write data to %s: %v", server2clientConn.RemoteAddr(), utils.WrapErrorLocation(err)))
			}
		}
	}()


	// server2clientConn -> user2serverConn
	count := 1
	for {

		err := user2serverConn.SetDeadline(time.Now().Add(200 * time.Second))
		if err != nil {
			panic(fmt.Errorf("failed to set deadline for %s: %v", user2serverConn.RemoteAddr(), utils.WrapErrorLocation(err)))
		}
		err = server2clientConn.SetDeadline(time.Now().Add(200 * time.Second))
		if err != nil {
			panic(fmt.Errorf("failed to set deadline for %s: %v", server2clientConn.RemoteAddr(), utils.WrapErrorLocation(err)))
		}

		// if count < 9600 {
		// 	src.SetReadDeadline(time.Now().Add(time.Duration(count*60) * time.Second))
		// } else {
		// 	src.SetDeadline(time.Now().Add(8 * time.Hour))
		// }

		count++
		for{
			readbuffer := make([]byte, 1024)
			n, err := server2clientConn.Read(readbuffer)
			if err != nil {
				if err == io.EOF {
					fmt.Println("receive empty data from server, close conn")
					break
				}
				// panic(fmt.Errorf("failed to read data from %s: %v", server2clientConn.RemoteAddr(), utils.WrapErrorLocation(err)))
				log.Println("failed to read data from server:", err)
				return
			}
			if n == 0 {
				fmt.Println("receive empty data from server, close conn")
				return
			}
			fmt.Println("receive data from server:", string(readbuffer[:n]))
			user2serverConn.Write(readbuffer[:n])
		}
		_buf := bytes.Buffer{}
		_, err = io.Copy(&_buf, server2clientConn) // when src.close, it will panic
		if err != nil {
			panic(fmt.Errorf("failed to read data from %s: %v", server2clientConn.RemoteAddr(), utils.WrapErrorLocation(err)))
		}
		_data := _buf.Bytes()
		data, err := utils.AESDecryptWithKey(string(_data))
		if err != nil {
			panic(fmt.Errorf("failed to decrypt data: %v", utils.WrapErrorLocation(err)))
		}
		_, err = io.Copy(user2serverConn, bytes.NewReader(data)) // when src.close, it will panic
		if err != nil {
			panic(fmt.Errorf("failed to copy data from %s to %s: %v", server2clientConn.RemoteAddr(), user2serverConn.RemoteAddr(), utils.WrapErrorLocation(err)))
		}
	}
}

func TransForConnDataClient(local2clientConn *net.TCPConn, client2serverConn *net.TCPConn) {
	defer utils.RecoverAndLog()
	defer local2clientConn.Close()
	defer client2serverConn.Close()

	// quit := make(chan bool)
	go func() {
		defer utils.RecoverAndLog(func(err error) {
			// quit <- true
		})

		// local2clientConn -> client2serverConn
		for {
			for{
				readBuff := make([]byte, 1024)
				n, err := local2clientConn.Read(readBuff)
				if err != nil {
					if err == io.EOF {
						fmt.Println("receive empty data from local, close conn")
						break
					}
					// panic(fmt.Errorf("failed to read data from %s: %v", local2clientConn.RemoteAddr(), utils.WrapErrorLocation(err)))
					log.Println("failed to read data from local:", err)
					return
				}
				if n == 0 {
					fmt.Println("receive empty data from local, close conn")
					return
				}
				fmt.Println("receive data from local:", string(readBuff[:n]))
				client2serverConn.Write(readBuff[:n])
			}
			_buf := bytes.Buffer{}
			_, err := io.Copy(&_buf, local2clientConn) // when dst.close, it will panic
			if err != nil {
				panic(fmt.Errorf("failed to read data from %s: %v", local2clientConn.RemoteAddr(), utils.WrapErrorLocation(err)))
			}
			_data := _buf.Bytes()
			aesdata, err := utils.AESEncryptWithKey(_data)
			if err != nil {
				panic(fmt.Errorf("failed to encrypt data: %v", utils.WrapErrorLocation(err)))
			}

			_, err = io.Copy(client2serverConn, bytes.NewReader([]byte(aesdata))) // when dst.close, it will panic

			if err != nil {
				panic(fmt.Errorf("failed to write data to %s: %v", client2serverConn.RemoteAddr(), utils.WrapErrorLocation(err)))
			}
		}
	}()
	count := 1

	// client2serverConn -> local2clientConn
	for {

		err := local2clientConn.SetDeadline(time.Now().Add(200 * time.Second))
		if err != nil {
			panic(fmt.Errorf("failed to set deadline for %s: %v", local2clientConn.RemoteAddr(), utils.WrapErrorLocation(err)))
		}
		err = client2serverConn.SetDeadline(time.Now().Add(200 * time.Second))
		if err != nil {
			panic(fmt.Errorf("failed to set deadline for %s: %v", client2serverConn.RemoteAddr(), utils.WrapErrorLocation(err)))
		}

		// if count < 9600 {
		// 	src.SetReadDeadline(time.Now().Add(time.Duration(count*60) * time.Second))
		// } else {
		// 	src.SetDeadline(time.Now().Add(8 * time.Hour))
		// }

		for{
			readbuffer := make([]byte, 1024)
			n, err := client2serverConn.Read(readbuffer)
			if err != nil {
				if err == io.EOF {
					fmt.Println("receive empty data from server, close conn")
					break
				}
				// panic(fmt.Errorf("failed to read data from %s: %v", client2serverConn.RemoteAddr(), utils.WrapErrorLocation(err)))
				log.Println("failed to read data from server:", err)
				return
			}
			if n == 0 {
				fmt.Println("receive empty data from server, close conn")
				return
			}
			fmt.Println("receive data from server:", string(readbuffer[:n]))
			local2clientConn.Write(readbuffer[:n])
		}

		count++
		_buf := bytes.Buffer{}
		_, err = io.Copy(&_buf, client2serverConn) // when src.close, it will panic
		if err != nil {
			panic(fmt.Errorf("failed to read data from %s: %v", client2serverConn.RemoteAddr(), utils.WrapErrorLocation(err)))
		}
		_data := _buf.Bytes()
		aesdata, err := utils.AESDecryptWithKey(string(_data))
		if err != nil {
			panic(fmt.Errorf("failed to decrypt data: %v", utils.WrapErrorLocation(err)))
		}
		_, err = io.Copy(local2clientConn, bytes.NewReader(aesdata)) // when src.close, it will panic
		if err != nil {
			panic(fmt.Errorf("failed to copy data from %s to %s: %v", client2serverConn.RemoteAddr(), local2clientConn.RemoteAddr(), utils.WrapErrorLocation(err)))
		}
	}
}
