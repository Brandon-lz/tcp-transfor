package toclient

import (
	"fmt"
	"io"
	"log"
	"net"
	"strings"
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
			if strings.Contains(err.Error(), "use of closed network connection") {
				return
			}
			continue
		}
		log.Printf("New client conn from %s", conn.RemoteAddr())
		go dealCmdFromClient(conn)
	}
}

var CCMList = make(map[string]*clientConnManager)

type ConnCouple struct {
	UserConn      net.Conn
	ClientSubConn net.Conn
	Id            int
}

type clientConnManager struct {
	ClientName          string
	ClientConn          net.Conn
	ClientSubConnWithId map[int]net.Conn
	clientSubConnIdSet  map[int]struct{}
	clientSubConnIdLock sync.Mutex
	Quit                chan bool
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

func dealCmdFromClient(clientConn net.Conn) {
	log.Printf("New client conn from %s", clientConn.RemoteAddr())

	defer utils.RecoverAndLog()
	// defer clientConn.Close()
	// clientName := clientConn.RemoteAddr().String()
	// for {
	log.Println("wait for hello message from client")

	hellodata, err := common.ReadConn(clientConn)
	// hellodata, err := io.ReadAll(clientConn)
	if err != nil {
		log.Printf("Failed to read hello message from client: %v", err)
		return
	}

	log.Println("receive hello message from client ", string(hellodata))
	hello := common.HelloMessage{}
	func(){
		defer utils.RecoverAndLog(func(err error){})
		utils.DeSerializeData(hellodata, &hello)
	}()
	switch hello.Type {
	case "main":

		if err := AddNewClient(&Client{Name: hello.Client.Name, Conn: clientConn, Map: hello.Map}); err != nil {
			hrc := utils.SerilizeData(&common.HelloRecv{Code: 500, Msg: fmt.Sprintf("client hello faild:%v", err)})

			clientConn.Write(hrc)
			return
		}

		// create new client conn manager
		ccm := clientConnManager{
			ClientConn:          clientConn,
			ClientSubConnWithId: make(map[int]net.Conn),
			clientSubConnIdSet:  make(map[int]struct{}),
			clientSubConnIdLock: sync.Mutex{},
			Quit:                make(chan bool),
		}

		ccm.ClientName = hello.Client.Name

		CCMList[hello.Client.Name] = &ccm

		// create new listener to client map port for listen user request
		log.Println("listen on client map port")
		var listen_fail = make(chan bool)
		var wg = sync.WaitGroup{}

		for _, m := range hello.Map {
			wg.Add(1)
			go newListenerOnClientMapPort(&ccm, m.ServerPort, m.LocalPort, listen_fail, &wg)
		}

		wg.Wait()

		select {
		case <-listen_fail:
			clientConn.Write(utils.SerilizeData(common.HelloRecv{Code: 500, Msg: "listen on client map port faild"}))
			return
		default:
			// finally success
			log.Printf("success listen on client %s map port %v", hello.Client.Name, hello.Map)
			clientConn.Write(utils.SerilizeData(common.HelloRecv{Code: 200, Msg: "hello success"})) // response to client main conn result
		}
		close(listen_fail)
	case "sub":
		// new sub conn from client
		ccm := CCMList[hello.Client.Name]
		ccm.ClientSubConnWithId[hello.ConnId] = clientConn
		return
	case "ping":

	default:
		log.Println("unknown hello type", hello.Type)
	}
}

func newListenerOnClientMapPort(ccm *clientConnManager, listenPort, clientLocalPort int, failSign chan bool, wg *sync.WaitGroup) {
	defer utils.RecoverAndLog()
	defer wg.Done()

	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", listenPort))
	if err != nil {
		log.Printf("Failed to listen on %s: %d: %v", ccm.ClientName, listenPort, err)
		failSign <- true
		return
	}
	defer listener.Close()

	log.Printf("New listener on %s:%d", ccm.ClientName, listenPort)
	go func() {
		defer utils.RecoverAndLog()
		for {
			userConn, err := listener.Accept()
			if err != nil {
				log.Printf("Failed to accept connection: %v", err)
				if strings.Contains(err.Error(), "use of closed network connection") { // exit goroutine
					return
				}
				continue
			}

			go whenNewUserConnComeIn(ccm, userConn, clientLocalPort, listenPort)
		}
	}()

	// } // wait for quit
	<-ccm.Quit
	log.Printf("listener on %s:%d quit", ccm.ClientName, listenPort)

}

func whenNewUserConnComeIn(ccm *clientConnManager, userConn net.Conn, clientLocalPort, listenPort int) {
	defer utils.RecoverAndLog()
	// new conn to server
	log.Println("new user conn ")
	connId := ccm.getNewConnId()
	if err := cmdToClientGetNewConn(ccm.ClientConn, connId, clientLocalPort, listenPort); err != nil {
		log.Printf("Failed to get new conn to client: %v", utils.WrapErrorLocation(err, "cmdToClientGetNewConn"))
		return
	}

	log.Println("success get new conn to client id:", connId)

	// wait new conn from client .....
	timeoutCount := 0
	for {
		if newSubConn, ok := ccm.ClientSubConnWithId[connId]; ok {
			// go TransForConnData(userConn, newSubConn, connId, ccm)
			go common.TransForConnData(userConn, newSubConn)
			return
		} else {
			timeoutCount++
			time.Sleep(20 * time.Microsecond)
		}
		if timeoutCount > 200 {
			log.Printf("ERROR: Timeout to get new conn from client id:%d", connId)
			return
		}
	}
}

func TransForConnData(src net.Conn, dst net.Conn, connid int, ccm *clientConnManager) {
	defer utils.RecoverAndLog()
	defer ccm.delConnId(connid)
	defer src.Close()
	defer dst.Close()

	defer log.Println("tcp channel exit")

	// quit := make(chan bool)
	go func() {
		defer utils.RecoverAndLog(func(err error) {
			// quit <- true
		})
		for {
			// src.SetDeadline(time.Now().Add(8 * time.Hour))
			// if count < 9600 {
			// 	userConn.SetReadDeadline(time.Now().Add(time.Duration(count) * 3 * time.Second))
			// } else {
			// 	userConn.SetDeadline(time.Now().Add(8 * time.Hour))
			// }

			_, err := io.Copy(dst, src)

			// data, err := common.ReadConn(userConn)
			// if err != nil {
			// 	panic(fmt.Errorf("Failed to copy data from %s to %s: %v\n", userConn.RemoteAddr(), dst.RemoteAddr(), utils.WrapErrorLocation(err)))
			// }
			// if len(data) == 0 {
			// 	log.Println("receive empty data from client, close conn")
			// 	break
			// }
			// _, err = dst.Write(data)

			if err != nil {
				panic(fmt.Errorf("Failed to write data to %s: %v\n", dst.RemoteAddr(), utils.WrapErrorLocation(err)))
			}
		}
	}()
	count := 1

	// trans:
	for {
		// select {
		// case <-quit:
		// break trans
		// default:
		// src.SetDeadline(time.Now().Add(8 * time.Hour))
		// dst.SetDeadline(time.Now().Add(8 * time.Hour))
		if count < 9600 {
			src.SetReadDeadline(time.Now().Add(time.Duration(count*config.Config.Timeout) * time.Second))
		} else {
			src.SetDeadline(time.Now().Add(8 * time.Hour))
		}

		count++
		_, err := io.Copy(src, dst)
		if err != nil {
			panic(fmt.Errorf("Failed to copy data from %s to %s: %v\n", dst.RemoteAddr(), src.RemoteAddr(), utils.WrapErrorLocation(err)))
		}
		// }
	}
}
