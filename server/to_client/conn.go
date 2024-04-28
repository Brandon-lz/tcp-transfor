package toclient

import (
	"fmt"
	"io"
	"log"
	"net"
	"sync"
	"time"

	"github.com/Brandon-lz/tcp-transfor/common"
	"github.com/Brandon-lz/tcp-transfor/server/config"
	"github.com/Brandon-lz/tcp-transfor/utils"
)

// listen to accecpt client main conn and sub conn
func ListenClientConn() {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", config.Config.Port))
	if err != nil {
		log.Fatalf("Failed to listen on %d: %v", config.Config.Port, err)
	}
	defer listener.Close()
	log.Println("Server is listening on port: ", config.Config.Port)

	for {
		conn, err := listener.Accept() // new client main conn or sub conn
		if err != nil {
			log.Printf("Failed to accept connection: %v", err)
			continue
		}

		go startCmdToClient(conn)
	}
}

var CCMList map[string]*clientConnManager

type ConnCouple struct {
	UserConn      *net.Conn
	ClientSubConn *net.Conn
	Id            int
}

type clientConnManager struct {
	ClientName          string
	ClientConn          net.Conn
	ClientSubConnWithId map[int]net.Conn
	clientSubConnIdSet  map[int]struct{}
	clientSubConnIdLock sync.Mutex
}

func (cm *clientConnManager) getNewConnId() int {
	cm.clientSubConnIdLock.Lock()
	defer cm.clientSubConnIdLock.Unlock()
	for i := range 1000000 {
		if _, ok := cm.clientSubConnIdSet[i]; !ok {
			cm.clientSubConnIdSet[i] = struct{}{}
			return i
		}
	}
	panic("too many conn id")
}

func (cm *clientConnManager) delConnId(id int) {
	cm.clientSubConnIdLock.Lock()
	defer cm.clientSubConnIdLock.Unlock()
	delete(cm.clientSubConnIdSet, id)
	delete(cm.ClientSubConnWithId, id)
}

func startCmdToClient(clientConn net.Conn) {
	defer utils.RecoverAndLog()
	defer clientConn.Close()
	// clientName := clientConn.RemoteAddr().String()

	for {
		hellodata, err := io.ReadAll(clientConn)
		if err != nil {
			log.Printf("Failed to read hello message from client: %v", err)
			return
		}

		hello := utils.DeSerializeData(hellodata, &common.HelloMessage{})
		switch hello.Type {
		case "main":
			if err := AddNewClient(&Client{Name: hello.Client.Name, Conn: clientConn, Map: hello.Map}); err != nil {
				clientConn.Write(utils.SerilizeData(common.HelloRecv{Code: 500, Msg: fmt.Sprintf("client hello faild:%v", err)}))
				return
			}
			// create new client conn manager
			ccm := clientConnManager{
				ClientConn:          clientConn,
				ClientSubConnWithId: make(map[int]net.Conn),
				clientSubConnIdSet:  make(map[int]struct{}),
				clientSubConnIdLock: sync.Mutex{},
			}
			ccm.ClientName = hello.Client.Name

			CCMList[hello.Client.Name] = &ccm

			// create new listener to client map port for listen user request
			var listen_fail = make(chan bool)
			var wg = sync.WaitGroup{}

			for _, m := range hello.Map {
				wg.Add(1)
				go newListenerOnClientMapPort(&ccm, m.ServerPort, m.LocalPort, listen_fail)
			}

			wg.Wait()

			select {
			case <-listen_fail:
				clientConn.Write(utils.SerilizeData(common.HelloRecv{Code: 500, Msg: "listen on client map port faild"}))
				return
			default:
				// finally success
				clientConn.Write(utils.SerilizeData(common.HelloRecv{Code: 200, Msg: "hello success"})) // response to client main conn result
			}
			close(listen_fail)
		case "sub":
			// new sub conn from client

			ccm := CCMList[hello.Client.Name]
			ccm.ClientSubConnWithId[hello.ConnId] = clientConn
			
		}
	}
}

func newListenerOnClientMapPort(ccm *clientConnManager, listenPort, clientLocalPort int, failSign chan bool) {
	defer utils.RecoverAndLog()
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", listenPort))
	if err != nil {
		log.Printf("Failed to listen on%s: %d: %v", ccm.ClientName, listenPort, err)
		failSign <- true
		return
	}
	defer listener.Close()

	for {
		userConn, err := listener.Accept()
		if err != nil {
			log.Printf("Failed to accept connection: %v", err)
			continue
		}
		// new conn to server
		connId := ccm.getNewConnId()
		if err := cmdToClientGetNewConn(ccm.ClientConn, connId, clientLocalPort, listenPort); err != nil {
			log.Printf("Failed to get new conn to client: %v", utils.WrapErrorLocation(err, "cmdToClientGetNewConn"))
			return
		}

		// wait new conn from client .....
		for {
			if newSubConn,ok:= ccm.ClientSubConnWithId[connId];ok{
				go TransForConnData(userConn, newSubConn,connId,ccm)
				break
			}else{
				time.Sleep(20*time.Microsecond)
			}
		}
		

		// cmd to client to get a new conn with client
		// client, err := getClientByName(hello.Client.Name)
		// if err != nil {
		// 	log.Printf("Failed to get client conn: %v", utils.WrapErrorLocation(err, "getClientByName"))
		// 	return err
		// }

		// wait new conn from client

	}
}





func TransForConnData(src net.Conn, dst net.Conn, connid int,ccm *clientConnManager) {
	defer utils.RecoverAndLog()
	defer ccm.delConnId(connid)
	defer src.Close()
	defer dst.Close()


	quit := make(chan bool)
	go func() {
		defer utils.RecoverAndLog(func(err error) { quit <- true })
		for {
			_, err := io.Copy(dst, src)
			if err != nil {
				panic(fmt.Errorf("Failed to copy data from %s to %s: %v\n", src.RemoteAddr(), dst.RemoteAddr(), utils.WrapErrorLocation(err)))
			}

		}
	}()

trans:
	for {
		select {
		case <-quit:
			break trans
		default:
			_, err := io.Copy(src, dst)
			if err != nil {
				panic(fmt.Errorf("Failed to copy data from %s to %s: %v\n", dst.RemoteAddr(), src.RemoteAddr(), utils.WrapErrorLocation(err)))
			}
		}
	}

}
