package toclient

import (
	"fmt"
	"log"
	"net"
	"sync"
	"sync/atomic"
	"time"

	"github.com/Brandon-lz/tcp-transfor/common"
	"github.com/Brandon-lz/tcp-transfor/server/config"
	"github.com/Brandon-lz/tcp-transfor/utils"

	gopubsub "github.com/Brandon-lz/go-pubsub"
)

var (
	ccmMu          sync.RWMutex
	CCMList        = make(map[string]*clientConnManager) // map[clientName]
	quitAgent      = gopubsub.NewAgent()
	serverListener *net.TCPListener
	connIdCounter  atomic.Uint32
)

func getCCMByName(name string) (*clientConnManager, bool) {
	ccmMu.RLock()
	defer ccmMu.RUnlock()
	ccm, ok := CCMList[name]
	return ccm, ok
}

func removeCCMByName(name string) {
	ccmMu.Lock()
	defer ccmMu.Unlock()
	delete(CCMList, name)
}

// clientConnManager 管理一个客户端的主连接和它的子连接
type clientConnManager struct {
	ClientName   string
	ClientConn   net.Conn
	LR           *common.LineReader // 主连接复用的行读取器
	ckmu         *sync.Mutex        // 保护主连接命令收发 + closed + 子连接表（与心跳互斥）
	ccmCond      *sync.Cond         // 子连接就绪信号（绑定 ckmu）
	closed       bool
	ClientSubConnWithId map[int]net.Conn
}

func NewclientConnManager(clientConn net.Conn, clientName string, lr *common.LineReader) *clientConnManager {
	ccm := &clientConnManager{
		ClientName:          clientName,
		ClientConn:          clientConn,
		LR:                  lr,
		ClientSubConnWithId: make(map[int]net.Conn),
	}
	ccm.ckmu = &sync.Mutex{}
	ccm.ccmCond = sync.NewCond(ccm.ckmu)
	return ccm
}

// markClosed 标记关闭并唤醒所有等待者
func (cm *clientConnManager) markClosed() {
	cm.ckmu.Lock()
	cm.closed = true
	cm.ccmCond.Broadcast()
	cm.ckmu.Unlock()
	cm.ClientConn.Close()
}

// waitSubConnReady 阻塞等待指定 id 的子连接注册，带超时；无忙轮询。
// 全程持有 ckmu（与心跳、命令收发串行），靠 registerSub 的 Broadcast 唤醒。
func (cm *clientConnManager) waitSubConnReady(connId int, timeout time.Duration) (net.Conn, bool) {
	cm.ckmu.Lock()
	defer cm.ckmu.Unlock()
	deadline := time.Now().Add(timeout)
	for {
		if sub, ok := cm.ClientSubConnWithId[connId]; ok {
			return sub, true
		}
		if cm.closed || time.Now().After(deadline) {
			return nil, false
		}
		cm.ccmCond.Wait()
	}
}

// registerSub 注册子连接并唤醒等待者（必须持有者内部加锁 ckmu）
func (cm *clientConnManager) registerSub(connId int, sub net.Conn) bool {
	cm.ckmu.Lock()
	defer cm.ckmu.Unlock()
	if _, dup := cm.ClientSubConnWithId[connId]; dup {
		return false
	}
	cm.ClientSubConnWithId[connId] = sub
	cm.ccmCond.Broadcast()
	return true
}

// delSubConn 转发结束后释放子连接
func (cm *clientConnManager) delSubConn(id int) {
	cm.ckmu.Lock()
	sub, ok := cm.ClientSubConnWithId[id]
	if ok {
		delete(cm.ClientSubConnWithId, id)
	}
	cm.ckmu.Unlock()
	if ok {
		sub.Close()
	}
}

func getNewConnId() int {
	return int(connIdCounter.Add(1) % 1000000)
}

// portMapping 对应 client hello 消息里 map 条目（字段与 common.HelloMessage.Map 对齐）
type portMapping struct {
	LocalHost string `json:"local-host"`
	LocalPort  int    `json:"local-port"`
	ServerPort int    `json:"server-port"`
}

func CreateListener(host string, port int) (*net.TCPListener, error) {
	addr, err := net.ResolveTCPAddr("tcp", fmt.Sprintf("%s:%d", host, port))
	if err != nil {
		return nil, err
	}
	return net.ListenTCP("tcp", addr)
}

// ListenClientConn 监听客户端的主连接和子连接
func ListenClientConn() {
	var err error
	serverListener, err = CreateListener("0.0.0.0", config.Config.Port)
	if err != nil {
		log.Printf("Failed to listen on port %d: %v", config.Config.Port, err)
		return
	}
	defer serverListener.Close()
	log.Println("Server is listening on port: ", config.Config.Port)

	for {
		conn, err := serverListener.AcceptTCP() // 新的客户端主连接或子连接
		if err != nil {
			if isClosedErr(err) {
				return
			}
			continue
		}
		log.Printf("New client conn from %s", conn.RemoteAddr())
		go dealCmdFromClient(common.NewConnLocked(conn))
	}
}

// CheckClientAlive 每 2 秒 ping 一次所有客户端，断开则清理
func CheckClientAlive() {
	defer utils.RecoverAndLog()
	for {
		time.Sleep(2 * time.Second)

		ccmMu.RLock()
		ccms := make([]*clientConnManager, 0, len(CCMList))
		for _, ccm := range CCMList {
			ccms = append(ccms, ccm)
		}
		ccmMu.RUnlock()

		for _, ccm := range ccms {
			dead := false
			ccm.ckmu.Lock()
			if err := common.SendCmd(ccm.ClientConn, utils.SerilizeData(common.ServerCmd{Type: "ping"})); err != nil {
				log.Println("与", ccm.ClientName, "的连接断开")
				dead = true
			}
			if !dead {
				ccm.ClientConn.SetReadDeadline(time.Now().Add(6 * time.Second))
				pong, err := ccm.LR.ReadOne()
				if err != nil {
					log.Println("与", ccm.ClientName, "的连接断开")
					dead = true
				} else if string(pong) != "pong" {
					log.Println("与", ccm.ClientName, "的连接断开: bad pong", string(pong))
					dead = true
				}
			}
			ccm.ckmu.Unlock()

			if dead {
				log.Println("Client ", ccm.ClientName, " disconnected")
				quitAgent.Publish(ccm.ClientName, "quit")
				ccm.markClosed()
				RemoveClient(ccm.ClientName)
				removeCCMByName(ccm.ClientName)
			}
		}
	}
}

// isClosedErr 判断是否是 listener 已关闭
func isClosedErr(err error) bool {
	return err != nil && err.Error() == "use of closed network connection"
}

func dealCmdFromClient(clientConn net.Conn) {
	defer utils.RecoverAndLog()
	log.Println("wait for hello message from client")

	lr := common.NewLineReader(clientConn)

	hellodata, err := lr.ReadOne()
	if err != nil {
		log.Printf("Failed to read hello message from client: %v", err)
		return
	}

	log.Println("receive hello message from client", string(hellodata))
	hello := common.HelloMessage{}
	_, err = utils.DeSerializeData(hellodata, &hello)
	if err != nil {
		log.Printf("Failed to deserialize hello message from client: %v", err)
		return
	}
	switch hello.Type {
	case "main":
		if err := AddNewClient(&Client{Name: hello.Client.Name, Conn: clientConn, Map: hello.Map}); err != nil {
			hrc := utils.SerilizeData(&common.HelloRecv{Code: 500, Msg: fmt.Sprintf("client hello failed:%v", err)})
			_ = common.SendCmd(clientConn, hrc)
			return
		}

		ccm := NewclientConnManager(clientConn, hello.Client.Name, lr)

		// 为 client map 的每个 server-port 建监听。
		// 注意：此时 ccm 尚未加入 CCMList，心跳不会触碰这条主连接，
		// 因此 hello 握手期间 ping 不会插队（修复客户端收到 ping 而非 hello-success 的竞态）。
		log.Println("listen on client map port")
		failSign := make(chan error, len(hello.Map))
		var wg sync.WaitGroup
		for i := range hello.Map {
			m := &hello.Map[i]
			wg.Add(1)
			go func(lh string, lp, sp int) {
				defer wg.Done()
				newListenerOnClientMapPort(ccm, sp, lh, lp, failSign)
			}(m.LocalHost, m.LocalPort, m.ServerPort)
		}
		wg.Wait()

		select {
		case lerr := <-failSign:
			log.Printf("listen on client map port failed: %v", lerr)
			RemoveClient(hello.Client.Name) // 清理，保证同名客户端可以重试
			_ = common.SendCmd(clientConn, utils.SerilizeData(common.HelloRecv{Code: 500, Msg: "listen on client map port failed"}))
			ccm.markClosed()
			return
		default:
			// 全部监听成功：先回 200，再把 ccm 暴露给心跳
			log.Printf("success listen on client %s map port %v", hello.Client.Name, hello.Map)
			_ = common.SendCmd(clientConn, utils.SerilizeData(common.HelloRecv{Code: 200, Msg: "hello success"}))
		}
		close(failSign)

		ccmMu.Lock()
		CCMList[hello.Client.Name] = ccm
		ccmMu.Unlock()
	case "sub":
		_ = common.SendCmd(clientConn, []byte("ok")) // sub 连接是新连接，无并发竞争
		d, err := lr.ReadOne()
		if err != nil {
			log.Printf("Failed to read ready from client sub conn: %v", err)
			return
		}
		if string(d) != "ready" {
			utils.PrintDataAsJson(fmt.Sprintf("client %s sub conn not ready: %s", hello.Client.Name, d))
			return
		}

		if ccm, ok := getCCMByName(hello.Client.Name); ok {
			if ccm.registerSub(hello.ConnId, clientConn) {
				utils.PrintDataAsJson(fmt.Sprintf("client %s sub conn ready id:%d", hello.Client.Name, hello.ConnId))
				// 通过主连接通知客户端：子连接已就绪，可以开始转发
				// （主连接与心跳共享 ckmu 锁，保证命令收发互斥）
				ccm.ckmu.Lock()
				_ = common.SendCmd(ccm.ClientConn, utils.SerilizeData(common.ServerCmd{Type: "sub-conn-ready", Data: hello.ConnId}))
				ccm.ckmu.Unlock()
			} else {
				utils.PrintDataAsJson(fmt.Sprintf("client %s sub conn id:%d duplicate", hello.Client.Name, hello.ConnId))
			}
		} else {
			utils.PrintDataAsJson(fmt.Sprintf("client %s sub conn: ccm not found", hello.Client.Name))
		}
		// sub 连接保持打开，转发由 whenNewUserConnComeIn 启动，结束后 delSubConn 关闭
		return
	default:
		log.Println("unknown hello type", hello.Type)
	}
}

// 在 server-port 上监听用户连接。函数返回时 listener 已建立（accept 循环在独立 goroutine）。
func newListenerOnClientMapPort(ccm *clientConnManager, listenPort int, clientLocalHost string, clientLocalPort int, failSign chan error) {
	defer utils.RecoverAndLog()

	listener, err := CreateListener("0.0.0.0", listenPort)
	if err != nil {
		log.Printf("[listener] Failed to listen on %s: %d: %v", ccm.ClientName, listenPort, err)
		failSign <- err
		return
	}

	log.Printf("New listener on %s:%d", ccm.ClientName, listenPort)

	// 订阅客户端退出信号，退出时取消订阅并关闭 listener
	suber, cancel := quitAgent.Subscribe(ccm.ClientName)
	go func() {
		<-suber.Msg
		cancel(quitAgent, suber)
		log.Printf("listener on %s:%d quit", ccm.ClientName, listenPort)
		listener.Close()
	}()

	// accept 循环放独立 goroutine：本函数要立刻返回，让调用方的 wg.Wait 得以完成
	go func() {
		defer utils.RecoverAndLog()
		defer listener.Close()
		for {
			userConn, err := listener.AcceptTCP() // 新的用户连接
			if err != nil {
				if isClosedErr(err) { // listener 已关闭，退出
					return
				}
				continue
			}

			userConnl := common.NewConnLocked(userConn)
			go whenNewUserConnComeIn(ccm, userConnl, clientLocalHost, clientLocalPort, listenPort)
		}
	}()
}

// whenNewUserConnComeIn 用户连接进来：通知客户端新建本地连接 -> 等待子连接 ready -> 双向转发
func whenNewUserConnComeIn(ccm *clientConnManager, userConn net.Conn, clientLocalHost string, clientLocalPort, listenPort int) {
	defer utils.RecoverAndLog()

	connId := getNewConnId()
	if err := cmdToClientGetNewConn(ccm, connId, clientLocalHost, clientLocalPort, listenPort); err != nil {
		log.Printf("Failed to get new conn to client: %v", utils.WrapErrorLocation(err, "cmdToClientGetNewConn"))
		userConn.Close()
		return
	}

	// 条件变量等待子连接 ready（最长 10s），替代原有的 2ms 忙轮询
	subConn, ok := ccm.waitSubConnReady(connId, 10*time.Second)
	if !ok {
		log.Printf("ERROR: Timeout to get new conn from client id:%d", connId)
		userConn.Close()
		return
	}

	utils.PrintDataAsJson(fmt.Sprintf("get sub conn success:%d", connId))
	go func() {
		defer utils.RecoverAndLog()
		common.TransForConnDataServer(userConn, subConn)
		ccm.delSubConn(connId) // 转发结束后释放子连接
	}()
}
