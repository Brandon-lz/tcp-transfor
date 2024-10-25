package common

import (
	"io"
	"log"
	"net"

	"github.com/Brandon-lz/tcp-transfor/utils"
)

// 1-> 4   3->2

func TransForConnDataServer(user2serverConn net.Conn, server2clientConn net.Conn) {
	defer utils.RecoverAndLog()
	defer user2serverConn.Close()
	defer server2clientConn.Close()

	// user -> server

	server2clientConnSocket := NewConnSocket(server2clientConn)

	go func() {
		defer utils.RecoverAndLog()

		defer user2serverConn.Close()
		defer server2clientConn.Close()

		// user -> server
		// user2serverConn.SetDeadline(time.Now().Add(200 * time.Second))
		readBuff := make([]byte, 1024)
		for {
			utils.PrintDataAsJson("1------------1-")
			n, err := user2serverConn.Read(readBuff) // receive from user
			if err != nil {
				if err == io.EOF {
					return
				}
				log.Println("receive data from user error, close conn", err.Error(), utils.GetCodeLine(1))
				break
			}
			utils.PrintDataAsJson(string(readBuff[:n]))
			utils.PrintDataAsJson("1------------2-")
			if n == 0 {
				if CheckConnIsClosed(user2serverConn) {
					log.Println("receive empty data from user, close conn", utils.GetCodeLine(1))
					break
				}
				continue
			}

			// log.Println("从用户接收到并传送给客户端的数据:", string(readBuff[:n]))

			// _data, err := utils.AESEncryptWithKey(readBuff)
			// if err != nil {
			// 	log.Println("failed to decrypt data from user:", err)
			// 	return
			// }
			// _, err = server2clientConn.Write([]byte(_data))

			// _, err = server2clientConn.Write(utils.AESEncrypt(readBuff[:n]))
			// aesd := utils.AESEncrypt(readBuff[:n])
			// utils.PrintDataAsJson(aesd)
			// err = SendCmd(server2clientConn, aesd)     // send to client
			_, err = server2clientConnSocket.SendLine(readBuff[:n]) // send to client
			// err = SendCmd(server2clientConn, readBuff[:n]) // send to client
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

		// readbuffer := make([]byte, 1024)
		for {
			// n, err := server2clientConn.Read(readbuffer)      // receive from client
			utils.PrintDataAsJson("2------------1-")
			// data, err := ReadCmd(server2clientConn) // bug wait here!
			data, err := server2clientConnSocket.RecvLine()
			if err != nil {
				if err == io.EOF {
					return
				}
				log.Println("receive data from client error, close conn", err, utils.GetCodeLine(1))
				return
			}
			utils.PrintDataAsJson(string(data))
			utils.PrintDataAsJson("2------------2-")

			// if n == 0 {
			// 	if CheckConnIsClosed(server2clientConn) {
			// 		log.Println("receive empty data from client, close conn")
			// 		return
			// 	}
			// 	continue
			// }

			// log.Println("从客户端接收到并传送给用户的数据:", string(readbuffer[:n]))

			// daesData, err := utils.AESDecrypt(data)
			// if err != nil {
			// 	log.Println("failed to decrypt data from client", err)
			// }
			// _, err = user2serverConn.Write(daesData) // send to user
			// _, err = user2serverConn.Write(data) // send to user
			_, err = user2serverConn.Write(data)
			if err != nil {
				log.Println("failed to write data to user:", err)
				return
			}
		}

	}
}

func TransForConnDataClient(local2clientConn net.Conn, client2serverConn net.Conn) {
	defer utils.RecoverAndLog()
	defer local2clientConn.Close()
	defer client2serverConn.Close()

	client2serverConnSocket := NewConnSocket(client2serverConn)

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
			utils.PrintDataAsJson("3------------1-")
			// n, err := local2clientConn.Read(readBuff) // read from local    // bug wait here!
			n, err := local2clientConn.Read(readBuff)
			if err != nil {
				if err == io.EOF {
					return
				}
				log.Println("receive data from local error, close conn", err, utils.GetCodeLine(1))
				return
			}
			utils.PrintDataAsJson(string(readBuff[:n]))

			utils.PrintDataAsJson("3------------2-")
			if n == 0 {
				// if CheckConnIsClosed(local2clientConn) {
				utils.PrintDataAsJson("receive empty data from local, close conn")
				// 	return
				// }
				continue
			}
			// log.Println("从本地接收到并传送给服务器的数据:", string(readBuff[:n]))
			// _data, err := utils.AESEncryptWithKey(readBuff[:n])
			// if err != nil {
			// 	log.Println("failed to encrypt data from local:", err)
			// 	return
			// }
			// _, err = client2serverConn.Write([]byte(_data))
			// aesd := utils.AESEncrypt(readBuff[:n])
			// utils.PrintDataAsJson(aesd)
			// err = SendCmd(client2serverConn,aesd) // send to server
			// err = SendCmd(client2serverConn, readBuff[:n]) // send to server
			_, err = client2serverConnSocket.SendLine(readBuff[:n]) // send to server
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

	// readbuffer := make([]byte, 1024)
	for {
		// n, err := client2serverConn.Read(readbuffer)
		utils.PrintDataAsJson("4------------1-")
		// data, err := ReadCmd(client2serverConn) // read from server
		data, err := client2serverConnSocket.RecvLine()
		if err != nil {
			if err == io.EOF {
				return
			}
			log.Println("receive data from server error, close conn", err, utils.GetCodeLine(1))
			return
		}
		utils.PrintDataAsJson(string(data))
		utils.PrintDataAsJson("4------------2-")
		// if n == 0 {
		// 	if CheckConnIsClosed(client2serverConn) {
		// 		log.Println("receive empty data from server, close conn")
		// 		return
		// 	}
		// 	continue
		// }
		// log.Println("从服务器接收到并传送给本地的数据:", string(readbuffer[:n]))
		// _data, err := utils.AESDecryptWithKey(string(readbuffer[:n]))
		// if err != nil {
		// 	log.Println("failed to decrypt data from server:", err)
		// 	return
		// }
		// _, err = local2clientConn.Write(_data)
		// _, err = local2clientConn.Write(readbuffer[:n])
		// daesData, err := utils.AESDecrypt(data)
		// if err != nil {
		// 	log.Println("failed to decrypt data from server:", err)
		// }
		// _, err = local2clientConn.Write(daesData) // send  to local
		// _, err = local2clientConn.Write(data) // send  to local
		_, err = local2clientConn.Write(data)
		if err != nil {
			log.Println("failed to write data to local:", err)
			return
		}
	}

}
