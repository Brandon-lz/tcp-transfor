package core

import (
	"fmt"
	"log"
	"net"
	"sync"
	"time"

	"github.com/Brandon-lz/tcp-transfor/client/config"
	"github.com/Brandon-lz/tcp-transfor/common"
	"github.com/Brandon-lz/tcp-transfor/utils"
)

const (
	// localDialTimeout 拨本地端口的上限。它要明显小于 server 侧等 ack 的超时（10s），
	// 这样本地连不上时 server 能立刻收到 500 而不是干等；同时它也是命令循环被拨号
	// 阻塞的最长时间，所以还必须小于 server 的心跳等待窗口（6s），
	// 否则一个慢拨号会让心跳应答迟到、被误判断连。
	localDialTimeout = 4 * time.Second

	// subConnReadyTimeout 等 server 回 sub-conn-ready 的上限。server 若已放弃这条用户
	// 连接（比如它那边的等待超时了），本地连接和子连接不能一直挂着。
	subConnReadyTimeout = 10 * time.Second
)

type readyEntry struct {
	ch chan struct{} // 收到 server 的 sub-conn-ready 信号后关闭
}

// serverCmdHandler 管理"已发出 new-conn-request、等待 server 确认 ready"的连接
// entries 由自身的 mu 保护，主循环与转发 goroutine 并发安全
type serverCmdHandler struct {
	mu      sync.Mutex
	entries map[int]*readyEntry
}

func newServerCmdHandler() *serverCmdHandler {
	return &serverCmdHandler{entries: make(map[int]*readyEntry)}
}

func (h *serverCmdHandler) add(connId int) *readyEntry {
	h.mu.Lock()
	defer h.mu.Unlock()
	e := &readyEntry{ch: make(chan struct{})}
	h.entries[connId] = e
	return e
}

// markReady 关闭 ch（重复通知安全）
func (h *serverCmdHandler) markReady(connId int) {
	e := h.get(connId)
	if e == nil {
		return // 已清理，忽略
	}
	select {
	case <-e.ch:
		// 已关闭
	default:
		close(e.ch)
	}
}

func (h *serverCmdHandler) get(connId int) *readyEntry {
	h.mu.Lock()
	defer h.mu.Unlock()
	return h.entries[connId]
}

func (h *serverCmdHandler) remove(connId int) {
	h.mu.Lock()
	defer h.mu.Unlock()
	delete(h.entries, connId)
}

// ListenServerCmd 主循环：读取 server 下发的命令并处理（阻塞直到断开）
// lr 与 hello 握手复用同一个 LineReader，避免预读字节丢失
func ListenServerCmd(conn net.Conn, lr *common.LineReader) {
	defer utils.RecoverAndLog(func(err error) {
		log.Printf("Error occurred in ListenServerCmd: %v\n", utils.WrapErrorLocation(err))
	})

	serverConn := common.NewConnSocketWithReader(conn, lr)
	h := newServerCmdHandler()

	for {
		msgdata, err := serverConn.RecvLine()
		if err != nil {
			log.Printf("Failed to communicate with server: %v\n", utils.WrapErrorLocation(err))
			return
		}

		cmd, err := utils.DeSerializeData(msgdata, &common.ServerCmd{})
		if err != nil {
			log.Println("Failed to deserialize server command: ", err)
			continue
		}
		fmt.Println("new command received from server: ", utils.PrintDataAsJson(cmd))
		switch cmd.Type {
		case "ping":
			err = common.SendCmd(serverConn, []byte("pong"))
			if err != nil {
				log.Println("Failed to send pong to server: ", err)
				return
			}
		case "new-conn-request":
			log.Println("Received new connection request from server")
			newcmd, err := utils.DeSerializeData(cmd.Data, &common.NewConnCreateRequestMessage{})
			if err != nil {
				log.Println("failed to deserialize new connection request message: ", err)
				continue
			}
			ip := "127.0.0.1"
			if newcmd.LocalHost != "" {
				ip = newcmd.LocalHost
			}
			// 先拨本地端口，再把结果回给 server：本地连不上时 server 立刻收到 500，
			// 不用像旧实现那样把失败推给子连接、靠 server 侧 10s 超时兜底
			localConn, err := CreateNewConnToRemotePortTimeout(ip, newcmd.LocalPort, localDialTimeout)
			if err != nil {
				log.Println("failed to create local connection: ", utils.WrapErrorLocation(err))
				common.SendCmd(serverConn, utils.SerilizeData(common.ResponseToServer{
					Id:   cmd.Id,
					Code: 500,
					Msg:  fmt.Sprintf("Failed to create local connection:%d", newcmd.LocalPort),
				}))
				continue
			}

			// ack to server（带上 conn-id：server 侧读取已异步化，靠 id 把并发请求的应答匹配回去）
			if err := common.SendCmd(serverConn, utils.SerilizeData(common.ResponseToServer{
				Id:   cmd.Id,
				Code: 200,
				Msg:  "new-conn-request-ack",
			})); err != nil {
				log.Println("failed to send new-conn-request-ack to server", utils.WrapErrorLocation(err))
				localConn.Close()
				continue
			}

			newServerSubConn, err := CreateNewConnToServer()
			if err != nil {
				log.Println("failed to create new connection to server: ", err)
				localConn.Close()
				continue
			}

			hello := common.HelloMessage{Type: "sub", ConnId: newcmd.ConnId}
			hello.Client.Name = config.Config.Client.Name
			if err := common.SendCmd(newServerSubConn, utils.SerilizeData(hello)); err != nil { // hello to server
				log.Println("fail to hello to server")
				newServerSubConn.Close()
				localConn.Close()
				continue
			}
			ok, err := common.ReadCmd(newServerSubConn) // receive "ok"
			if err != nil {
				log.Println("fail to ack new conn from server")
				newServerSubConn.Close()
				localConn.Close()
				continue
			}
			if string(ok) != "ok" {
				utils.PrintDataAsJson("fail to ack new conn from server:" + string(ok))
				newServerSubConn.Close()
				localConn.Close()
				continue
			}
			utils.PrintDataAsJson(fmt.Sprintf("success new sub connection to server: %d", hello.ConnId))

			entry := h.add(newcmd.ConnId)
			if err := common.SendCmd(newServerSubConn, []byte("ready")); err != nil {
				log.Println("fail to send ready to server")
				newServerSubConn.Close()
				localConn.Close()
				h.remove(newcmd.ConnId)
				continue
			}
			// 等 server 确认双向 ready 后开始转发（替代原来并发不安全的 map 信号）。
			// 带超时：server 若已放弃这条用户连接就不会再发 sub-conn-ready，
			// 无限等待的话本地连接和子连接会一直挂着。
			go func(connId int, lc net.Conn, sc net.Conn, e *readyEntry) {
				defer utils.RecoverAndLog()
				select {
				case <-e.ch:
				case <-time.After(subConnReadyTimeout):
					utils.PrintDataAsJson(fmt.Sprintf("timeout waiting sub-conn-ready from server, close sub conn %d", connId))
					h.remove(connId)
					lc.Close()
					sc.Close()
					return
				}
				utils.PrintDataAsJson(fmt.Sprintf("ready to transfor data %d", connId))
				go common.TransForConnDataClient(lc, sc)
				h.remove(connId)
				log.Println("sub connection established", connId)
			}(newcmd.ConnId, localConn, newServerSubConn, entry)
		case "sub-conn-ready":
			id, isId := cmd.Data.(float64)
			if !isId {
				log.Println("invalid sub-conn-ready data:", cmd.Data)
				continue
			}
			h.markReady(int(id))
		default:
			log.Println("Unknown command received from server", cmd.Type)
		}
	}
}
