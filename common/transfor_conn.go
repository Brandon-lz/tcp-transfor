package common

import (
	"fmt"
	"io"
	"log"
	"net"
	"time"

	"github.com/Brandon-lz/tcp-transfor/utils"
)

func TransForConnDataServer(user2serverConn *net.TCPConn, server2clientConn *net.TCPConn) {
	defer utils.RecoverAndLog()
	defer user2serverConn.Close()
	defer server2clientConn.Close()

	go func() {
		defer utils.RecoverAndLog()

		// user2serverConn -> server2clientConn
		for {
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
			_data, err := utils.AESEncryptWithKey(readBuff[:n])
			if err != nil {
				log.Println("failed to decrypt data from user:", err)
				return
			}
			_, err = server2clientConn.Write([]byte(_data))
			if err != nil {
				log.Println("failed to write data to server:", err)
				return
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

		count++
		for {
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
			// fmt.Println("receive data from server:", string(readbuffer[:n]))
			_data, err := utils.AESDecryptWithKey(string(readbuffer[:n]))
			if err != nil {
				log.Println("failed to decrypt data from server:", err)
				return
			}
			_, err = user2serverConn.Write(_data)
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

	// quit := make(chan bool)
	go func() {
		defer utils.RecoverAndLog(func(err error) {
			// quit <- true
		})

		// local2clientConn -> client2serverConn
		for {
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
			_data, err := utils.AESEncryptWithKey(readBuff[:n])
			if err != nil {
				log.Println("failed to encrypt data from local:", err)
				return
			}
			_, err = client2serverConn.Write([]byte(_data))
			if err != nil {
				log.Println("failed to write data to server:", err)
				return
			}
		}

	}()

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

		for {
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
			_data, err := utils.AESDecryptWithKey(string(readbuffer[:n]))
			if err != nil {
				log.Println("failed to decrypt data from server:", err)
				return
			}
			_, err = local2clientConn.Write(_data)
			if err != nil {
				log.Println("failed to write data to local:", err)
				return
			}
		}

	}
}
