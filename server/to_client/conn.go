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

const (
	pingInterval   = 2 * time.Second  // 心跳间隔
	pongTimeout    = 15 * time.Second // 等 pong 超时（放宽：客户端偶发卡顿不应直接判死整条隧道）
	ackTimeout     = 20 * time.Second // 等 new-conn-request 应答超时（客户端可能要先拨本地端口）
	subConnTimeout = 20 * time.Second // 等客户端子连接就绪超时
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
	ClientName string
	ClientConn net.Conn
	LR         *common.LineReader // 主连接复用的行读取器，只由 readLoop 读取

	ckmu   *sync.Mutex // 只用于串行化主连接上的写入，并保护 closed（不在持锁期间做网络等待）
	closed bool
	done   chan struct{} // 首次关闭时关闭，唤醒所有等待者

	pongWait chan struct{}                        // 心跳 pong 通知（容量 1，只有心跳 goroutine 消费）
	ackMu    sync.Mutex                           // 保护 acks
	acks     map[int]chan common.ResponseToServer // connId -> new-conn-request 应答等待者

	subMu               sync.Mutex            // 保护子连接表和就绪等待者
	ClientSubConnWithId map[int]net.Conn      // connId -> 客户端子连接
	subWaiters          map[int]chan net.Conn // connId -> 子连接就绪通知
}

func NewclientConnManager(clientConn net.Conn, clientName string, lr *common.LineReader) *clientConnManager {
	ccm := &clientConnManager{
		ClientName:          clientName,
		ClientConn:          clientConn,
		LR:                  lr,
		ClientSubConnWithId: make(map[int]net.Conn),
		subWaiters:          make(map[int]chan net.Conn),
		acks:                make(map[int]chan common.ResponseToServer),
		pongWait:            make(chan struct{}, 1),
		done:                make(chan struct{}),
	}
	ccm.ckmu = &sync.Mutex{}
	return ccm
}

// markClosed 标记关闭并唤醒所有等待者（幂等，返回是否首次关闭）
func (cm *clientConnManager) markClosed() bool {
	cm.ckmu.Lock()
	first := !cm.closed
	cm.closed = true
	cm.ckmu.Unlock()

	if !first {
		return false
	}
	close(cm.done) // 唤醒等子连接就绪、等 pong 的 goroutine
	cm.ClientConn.Close()

	// 唤醒正在等 new-conn-request 应答的请求者，避免它们空等一整个超时
	cm.ackMu.Lock()
	for id, ch := range cm.acks {
		select {
		case ch <- common.ResponseToServer{Id: id, Code: 500, Msg: "client connection closed"}:
		default:
		}
	}
	cm.ackMu.Unlock()
	return true
}

// cleanup 客户端确认失联后的统一清理：通知 listener 退出、摘掉 client 与 ccm（只执行一次）
func (cm *clientConnManager) cleanup(reason string) {
	if !cm.markClosed() {
		return
	}
	log.Printf("client %s disconnected: %s", cm.ClientName, reason)
	quitAgent.Publish(cm.ClientName, "quit")
	RemoveClient(cm.ClientName)
	removeCCMByName(cm.ClientName)
}

// ---- 主连接读取分派 ----

// readLoop 是主连接唯一的读取者：把 pong 和 new-conn-request 应答分派给各自的等待者。
// 读取过程不持有 ckmu，所以心跳等待、新建连接的等待、子连接注册互不阻塞。
// 也不使用 SetReadDeadline：读取超时全部由等待者自己的 select 超时表达，避免残留的
// 读截止时间影响这条连接上后续的读取。
func (cm *clientConnManager) readLoop() {
	defer utils.RecoverAndLog()
	for {
		line, err := cm.LR.ReadOne()
		if err != nil {
			cm.cleanup(fmt.Sprintf("main conn read error: %v", err))
			return
		}
		if string(line) == "pong" {
			cm.notifyPong()
			continue
		}
		resp, derr := utils.DeSerializeData(line, &common.ResponseToServer{})
		if derr != nil {
			utils.PrintDataAsJson(fmt.Sprintf("client %s: unexpected main conn message: %s", cm.ClientName, line))
			continue
		}
		cm.dispatchAck(resp)
	}
}

// drainPong 发新 ping 前清掉可能残留的旧 pong，避免把上一轮的应答当成这一轮的
func (cm *clientConnManager) drainPong() {
	select {
	case <-cm.pongWait:
	default:
	}
}

func (cm *clientConnManager) notifyPong() {
	select {
	case cm.pongWait <- struct{}{}:
	default: // 已有未消费的 pong，丢弃
	}
}

// addAck 注册应答等待者（必须在发送请求之前调用，避免应答先于注册到达）
func (cm *clientConnManager) addAck(connId int) chan common.ResponseToServer {
	ch := make(chan common.ResponseToServer, 1)
	cm.ackMu.Lock()
	cm.acks[connId] = ch
	cm.ackMu.Unlock()
	return ch
}

func (cm *clientConnManager) removeAck(connId int) {
	cm.ackMu.Lock()
	delete(cm.acks, connId)
	cm.ackMu.Unlock()
}

func (cm *clientConnManager) dispatchAck(resp common.ResponseToServer) {
	cm.ackMu.Lock()
	ch, ok := cm.acks[resp.Id]
	cm.ackMu.Unlock()
	if !ok {
		utils.PrintDataAsJson(fmt.Sprintf("client %s: unexpected ack id:%d", cm.ClientName, resp.Id))
		return
	}
	select {
	case ch <- resp:
	default:
	}
}

// ---- 子连接注册与等待 ----

// waitSubConnReady 阻塞等待指定 id 的子连接注册，等待期间不持有任何互斥锁
func (cm *clientConnManager) waitSubConnReady(connId int, timeout time.Duration) (net.Conn, bool) {
	cm.subMu.Lock()
	if sub, ok := cm.ClientSubConnWithId[connId]; ok { // 等待前就已经注册好了
		cm.subMu.Unlock()
		return sub, true
	}
	ch := make(chan net.Conn, 1)
	cm.subWaiters[connId] = ch
	cm.subMu.Unlock()
	defer func() {
		cm.subMu.Lock()
		delete(cm.subWaiters, connId)
		cm.subMu.Unlock()
	}()

	select {
	case sub := <-ch:
		return sub, true
	case <-cm.done: // 客户端已断开
		return nil, false
	case <-time.After(timeout):
		return nil, false
	}
}

// registerSub 注册子连接并唤醒等待者
func (cm *clientConnManager) registerSub(connId int, sub net.Conn) bool {
	cm.subMu.Lock()
	defer cm.subMu.Unlock()
	if _, dup := cm.ClientSubConnWithId[connId]; dup {
		return false
	}
	cm.ClientSubConnWithId[connId] = sub
	if ch, waiting := cm.subWaiters[connId]; waiting {
		select {
		case ch <- sub:
		default:
		}
	}
	return true
}

// delSubConn 转发结束后释放子连接
func (cm *clientConnManager) delSubConn(id int) {
	cm.subMu.Lock()
	sub, ok := cm.ClientSubConnWithId[id]
	if ok {
		delete(cm.ClientSubConnWithId, id)
	}
	cm.subMu.Unlock()
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
		go dealCmdFromClient(conn)
	}
}

// CheckClientAlive 每 2 秒给所有客户端发一次 ping。
// 分两阶段：先在持锁期间只做发送（微秒级），再在不持锁的情况下统一等 pong。
// 旧实现持着 ckmu 发 ping、读 pong（最长 6s），同一时间段内的 new-conn-request
// 和子连接注册全被堵住；读取超时也不再依赖 SetReadDeadline（那种截止时间会残留
// 在这条连接上，影响后续读取）。
func CheckClientAlive() {
	defer utils.RecoverAndLog()
	ping := utils.SerilizeData(common.ServerCmd{Type: "ping"})
	for {
		time.Sleep(pingInterval)

		ccmMu.RLock()
		ccms := make([]*clientConnManager, 0, len(CCMList))
		for _, ccm := range CCMList {
			ccms = append(ccms, ccm)
		}
		ccmMu.RUnlock()

		// 阶段一：只发 ping，锁只覆盖写入
		pinged := make([]*clientConnManager, 0, len(ccms))
		for _, ccm := range ccms {
			ccm.ckmu.Lock()
			if ccm.closed {
				ccm.ckmu.Unlock()
				continue
			}
			ccm.drainPong()
			err := common.SendCmd(ccm.ClientConn, ping)
			ccm.ckmu.Unlock()
			if err != nil {
				ccm.cleanup("ping send failed: " + err.Error())
				continue
			}
			pinged = append(pinged, ccm)
		}

		// 阶段二：等 pong，不持任何锁；pong 由 readLoop 分派
		for _, ccm := range pinged {
			select {
			case <-ccm.pongWait:
			case <-ccm.done: // readLoop 已判定断开并清理
			case <-time.After(pongTimeout):
				ccm.cleanup("ping timeout")
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

		// 主连接的读取统一交给 readLoop：pong 与 new-conn-request 应答都由它分派
		go ccm.readLoop()
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

			// 直接用 *net.TCPConn：转发路径上每条连接只有一个读 goroutine 和一个写 goroutine，
			// 主连接上的命令读写由 ccm.ckmu 串行化，包装加锁只会平白多两次 mutex 开销
			go whenNewUserConnComeIn(ccm, userConn, clientLocalHost, clientLocalPort, listenPort)
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

	// 等待子连接 ready（超时由 subConnTimeout 控制），等待期间不持任何锁
	subConn, ok := ccm.waitSubConnReady(connId, subConnTimeout)
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
