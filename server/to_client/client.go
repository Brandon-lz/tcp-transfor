package toclient

import (
	"fmt"
	"net"
	"sort"
	"sync"
)

var (
	clientMu  sync.RWMutex
	ClientSet = make(map[string]*Client)
)

type Client struct {
	Name string `json:"name"`
	Conn net.Conn
	Map  []struct {
		LocalHost string `json:"local-host"`
		LocalPort  int    `json:"local-port"`
		ServerPort int    `json:"server-port"`
	} `json:"map"`
}

func AddNewClient(client *Client) error {
	clientMu.Lock()
	defer clientMu.Unlock()

	if _, ok := ClientSet[client.Name]; ok {
		return fmt.Errorf("Client name already exists")
	}

	for _, c := range ClientSet {
		for _, m := range c.Map {
			for _, m2 := range client.Map {
				if m.ServerPort == m2.ServerPort {
					return fmt.Errorf("server port: %d already exists", m2.ServerPort)
				}
			}
		}
	}

	ClientSet[client.Name] = client
	return nil
}

// RemoveClient 删除客户端记录并关闭它的连接。
// conn 不为 nil 时，只有记录里的连接仍是这一条才删除——老会话迟到的清理
// 不能把同名新会话刚写入的连接一起关掉（那会让新客户端只看到 EOF）。
func RemoveClient(name string, conn net.Conn) int {
	clientMu.Lock()
	defer clientMu.Unlock()
	c, ok := ClientSet[name]
	if !ok {
		return 0
	}
	if conn != nil && c.Conn != conn {
		return 0
	}
	delete(ClientSet, name)
	if c.Conn != nil {
		_ = c.Conn.Close()
	}
	return 1
}

// MapView 是单条端口映射的快照
type MapView struct {
	LocalHost  string `json:"local-host"`
	LocalPort  int    `json:"local-port"`
	ServerPort int    `json:"server-port"`
}

// ClientView 是给管理界面用的客户端快照（拷贝出来的值，不持锁）
type ClientView struct {
	Name       string    `json:"name"`
	RemoteAddr string    `json:"remote-addr"`
	Map        []MapView `json:"map"`
}

// SnapshotClients 返回当前在线客户端的快照，按名字排序
func SnapshotClients() []ClientView {
	clientMu.RLock()
	defer clientMu.RUnlock()
	out := make([]ClientView, 0, len(ClientSet))
	for name, c := range ClientSet {
		v := ClientView{Name: name}
		if c.Conn != nil {
			v.RemoteAddr = c.Conn.RemoteAddr().String()
		}
		for _, m := range c.Map {
			v.Map = append(v.Map, MapView{
				LocalHost:  m.LocalHost,
				LocalPort:  m.LocalPort,
				ServerPort: m.ServerPort,
			})
		}
		out = append(out, v)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out
}

func GetClientByName(name string) (*Client, bool) {
	clientMu.RLock()
	defer clientMu.RUnlock()
	c, ok := ClientSet[name]
	return c, ok
}
