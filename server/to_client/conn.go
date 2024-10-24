package toclient

import (
	"fmt"
	"log"
	"net"
	"strings"
	"sync"
	"time"

	"github.com/Brandon-lz/tcp-transfor/common"
	"github.com/Brandon-lz/tcp-transfor/server/config"
	"github.com/Brandon-lz/tcp-transfor/utils"

	gopubsub "github.com/Brandon-lz/go-pubsub"
)

func createListener(host string, port int) (*net.TCPListener, error) {
	addr, err := net.ResolveTCPAddr("tcp", fmt.Sprintf("%s:%d", host, port))
	if err != nil {
		return nil, err
	}
	listener, err := net.ListenTCP("tcp", addr)
	if err != nil {
		return nil, err
	}
	return listener, nil
}

// listen to accecpt client main conn and sub conn
func ListenClientConn() {
	listener, err := createListener("0.0.0.0", config.Config.Port)
	if err != nil {
		log.Printf("Failed to listen on port %d: %v", config.Config.Port, err)
		return
	}
	defer listener.Close()
	log.Println("Server is listening on port: ", config.Config.Port)

	for {
		conn, err := listener.AcceptTCP() // new client main conn or sub conn
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

var quitAgent = gopubsub.NewAgent()

var CCMList = make(map[string]*clientConnManager) // map[clientName]

type ConnCouple struct {
	UserConn      *net.TCPConn
	ClientSubConn *net.TCPConn
	Id            int
}

type clientConnManager struct {
	ClientName          string
	ClientConn          *net.TCPConn
	Cmdrwlock           *sync.Mutex
	ClientSubConnWithId map[int]*net.TCPConn
	ClientSubConnReadySignalWithId map[int]chan bool          // 当前子连接准好的信号
	clientSubConnIdSet  map[int]struct{}
	clientSubConnIdLock sync.Mutex
	// Quit              string
}

func NewclientConnManager(clientConn *net.TCPConn, clientName string) clientConnManager {
	return clientConnManager{
		ClientName:          clientName,
		ClientConn:          clientConn,
		Cmdrwlock:           &sync.Mutex{},
		ClientSubConnWithId: make(map[int]*net.TCPConn),
		ClientSubConnReadySignalWithId: make(map[int]chan bool,2),
		clientSubConnIdSet:  make(map[int]struct{}),
		clientSubConnIdLock: sync.Mutex{},
		// Quit:     string,
	}
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

func dealCmdFromClient(clientConn *net.TCPConn) {
	defer utils.RecoverAndLog()
	log.Println("wait for hello message from client")

	// hellodata, err := common.ReadConn(clientConn)
	hellodata, err := common.ReadCmd(clientConn)
	// hellodata, err := io.ReadAll(clientConn)
	if err != nil {
		log.Printf("Failed to read hello message from client: %v", err)
		return
	}

	log.Println("receive hello message from client ", string(hellodata))
	hello := common.HelloMessage{}
	_, err = utils.DeSerializeData(hellodata, &hello)
	if err != nil {
		log.Printf("Failed to deserialize hello message from client: %v", err)
		return
	}
	switch hello.Type {
	case "main":

		if err := AddNewClient(&Client{Name: hello.Client.Name, Conn: clientConn, Map: hello.Map}); err != nil {
			hrc := utils.SerilizeData(&common.HelloRecv{Code: 500, Msg: fmt.Sprintf("client hello faild:%v", err)})

			// clientConn.Write(hrc)
			common.SendCmd(clientConn, hrc)
			return
		}

		// create new client conn manager
		// ccm := clientConnManager{
		// 	ClientConn:          clientConn,
		// 	ClientSubConnWithId: make(map[int]*net.TCPConn),
		// 	clientSubConnIdSet:  make(map[int]struct{}),
		// 	clientSubConnIdLock: sync.Mutex{},
		// 	Quit:                make(chan bool),
		// }
		ccm := NewclientConnManager(clientConn, hello.Client.Name)          // 从这里开始有并发竞争了

		ccm.ClientName = hello.Client.Name

		CCMList[hello.Client.Name] = &ccm

		// create new listener to client map port for listen user request
		log.Println("listen on client map port")
		var listen_fail = make(chan bool, 2)
		var wg = sync.WaitGroup{}

		for _, m := range hello.Map {
			wg.Add(1)
			go newListenerOnClientMapPort(&ccm, m.ServerPort, m.LocalPort, listen_fail, &wg)       // listen new user conn
		}

		wg.Wait()

		select {
		case <-listen_fail:
			// clientConn.Write(utils.SerilizeData(common.HelloRecv{Code: 500, Msg: "listen on client map port faild"}))
			log.Println("listen on client map port faild")
			ccm.Cmdrwlock.Lock()
			common.SendCmd(clientConn, utils.SerilizeData(common.HelloRecv{Code: 500, Msg: "listen on client map port faild"}))
			ccm.Cmdrwlock.Unlock()
			return
		default:
			// finally success
			log.Printf("success listen on client %s map port %v", hello.Client.Name, hello.Map)
			// clientConn.Write(utils.SerilizeData(common.HelloRecv{Code: 200, Msg: "hello success"})) // response to client main conn result
			ccm.Cmdrwlock.Lock()
			common.SendCmd(clientConn, utils.SerilizeData(common.HelloRecv{Code: 200, Msg: "hello success"})) // response to client main conn result
			ccm.Cmdrwlock.Unlock()
		}
		close(listen_fail)
	case "sub":
		// clientConn.Write([]byte("ok"))
		common.SendCmd(clientConn, []byte("ok"))          // 这里由于是新的连接，并不在ccm里，所以没有并发竞争
		// new sub conn from client
		// buf := make([]byte, 1024)
		// n, err := clientConn.Read(buf)
		d, err := common.ReadCmd(clientConn)
		if err != nil {
			log.Printf("Failed to read hello message from client: %v", err)
			return
		}
		if string(d) != "ready" {
			utils.PrintDataAsJson(fmt.Sprintf("client %s sub conn not ready", hello.Client.Name))
			return
		}
		ccm := CCMList[hello.Client.Name]
		ccm.ClientSubConnWithId[hello.ConnId] = clientConn
		ccm.ClientSubConnReadySignalWithId[hello.ConnId] = make(chan bool, 2)
		
		<-ccm.ClientSubConnReadySignalWithId[hello.ConnId]
		close(ccm.ClientSubConnReadySignalWithId[hello.ConnId])
		delete(ccm.ClientSubConnReadySignalWithId, hello.ConnId)
		common.SendCmd(ccm.ClientConn,utils.SerilizeData(common.ServerCmd{Type:"sub-conn-ready",Data: hello.ConnId}))
		utils.PrintDataAsJson(fmt.Sprintf("client %s sub conn ready", hello.Client.Name))
		return
	case "ping":
		common.SendCmd(clientConn, utils.SerilizeData(common.ServerCmd{Type: "Pone"}))
	default:
		log.Println("unknown hello type", hello.Type)
	}
}

func newListenerOnClientMapPort(ccm *clientConnManager, listenPort, clientLocalPort int, failSign chan bool, wg *sync.WaitGroup) {
	defer utils.RecoverAndLog()
	defer wg.Done()

	listener, err := createListener("0.0.0.0", listenPort)
	if err != nil {
		log.Printf("Failed to listen on %s: %d: %v", ccm.ClientName, listenPort, err)
		failSign <- true
		return
	}

	log.Printf("New listener on %s:%d", ccm.ClientName, listenPort)
	go func() {
		defer utils.RecoverAndLog()
		// } // wait for quit
		go func() {
			defer utils.RecoverAndLog()
			defer listener.Close()
			suber, cancel := quitAgent.Subscribe(ccm.ClientName)
			defer cancel(quitAgent, suber)
			<-suber.Msg // wait for quit, fixit
			log.Printf("listener on %s:%d quit", ccm.ClientName, listenPort)
		}()

		for {
			userConn, err := listener.AcceptTCP() // new user conn
			if err != nil {
				log.Printf("failed to accept connection: %v\n", err)
				if strings.Contains(err.Error(), "use of closed network connection") { // exit goroutine
					return
				}
				continue
			}

			go whenNewUserConnComeIn(ccm, userConn, clientLocalPort, listenPort)
		}
	}()

}

func whenNewUserConnComeIn(ccm *clientConnManager, userConn *net.TCPConn, clientLocalPort, listenPort int) {
	defer utils.RecoverAndLog()
	// new conn to server
	log.Println("new user conn ")
	connId := ccm.getNewConnId()
	if err := cmdToClientGetNewConn(ccm, connId, clientLocalPort, listenPort); err != nil {
		log.Printf("Failed to get new conn to client: %v", utils.WrapErrorLocation(err, "cmdToClientGetNewConn"))
		return
	}

	log.Println("success get new conn to client id:", connId)

	// wait new conn from client .....
	timeoutCount := 0
	for {
		if newSubConn, ok := ccm.ClientSubConnWithId[connId]; ok {
			ccm.ClientSubConnReadySignalWithId[connId] <- true
			go common.TransForConnDataServer(userConn, newSubConn) 
			return
		} else {
			timeoutCount++
			time.Sleep(2 * time.Millisecond)
		}
		if timeoutCount > 5000 {
			log.Printf("ERROR: Timeout to get new conn from client id:%d", connId)
			return
		}
	}
}
