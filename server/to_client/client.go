package toclient

import (
	"fmt"
	"net"
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

func RemoveClient(name string) int {
	clientMu.Lock()
	defer clientMu.Unlock()
	if c, ok := ClientSet[name]; ok {
		delete(ClientSet, name)
		if c.Conn != nil {
			_ = c.Conn.Close()
		}
		return 1
	}
	return 0
}

func GetClientByName(name string) (*Client, bool) {
	clientMu.RLock()
	defer clientMu.RUnlock()
	c, ok := ClientSet[name]
	return c, ok
}
