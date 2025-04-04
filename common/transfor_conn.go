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

		
			_, err = server2clientConnSocket.SendLine(readBuff[:n]) // send to client
			if err != nil {
				log.Println("failed to write data to server:", err)
				return
			}
		}

	}()

	// server -> user
	for {

		
		for {
			utils.PrintDataAsJson("2------------1-")
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


	go func() {
	
		defer local2clientConn.Close()
		defer client2serverConn.Close()

		// local -> server
		readBuff := make([]byte, 1024)
		for {
			utils.PrintDataAsJson("3------------1-")
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
				utils.PrintDataAsJson("receive empty data from local, close conn")
				continue
			}
			
			_, err = client2serverConnSocket.SendLine(readBuff[:n]) // send to server
			if err != nil {
				log.Println("failed to write data to server:", err)
				return
			}
		}

	}()

	// server -> local


	for {
		utils.PrintDataAsJson("4------------1-")
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
		
		_, err = local2clientConn.Write(data)
		if err != nil {
			log.Println("failed to write data to local:", err)
			return
		}
	}

}
