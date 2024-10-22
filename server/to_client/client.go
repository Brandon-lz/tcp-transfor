package toclient

import (
	"fmt"
	"log"
	"net"
	"time"

	"github.com/Brandon-lz/tcp-transfor/common"
	"github.com/Brandon-lz/tcp-transfor/utils"
)

var ClientSet = make(map[string]*Client)

type Client struct {
	Name    string `json:"name"`
	Conn    *net.TCPConn
	SubConn map[int]*net.TCPConn
	Map     []struct {
		LocalPort  int `json:"local-port"`
		ServerPort int `json:"server-port"`
	} `json:"map"`
}

func AddNewClient(client *Client) error {
	if _, ok := ClientSet[client.Name]; ok {
		return fmt.Errorf("Client name already exists")
	}

	ports := []int{}

	for _, m := range client.Map {
		ports = append(ports, m.ServerPort)
	}

	for _, c := range ClientSet {
		for _, m := range c.Map {
			for _, p := range ports {
				if m.ServerPort == p {
					return fmt.Errorf("server port: %d already exists", p)
				}
			}
		}
	}

	ClientSet[client.Name] = client
	return nil
}

func RemoveClient(name string) error {
	if _, ok := ClientSet[name]; ok {
		delete(ClientSet, name)
		return nil
	}
	return fmt.Errorf("Client name not found")
}

func GetClientByName(name string) (*Client, error) {
	if _, ok := ClientSet[name]; ok {
		return ClientSet[name], nil
	}
	return nil, fmt.Errorf("Client name not found")
}

func getClientByClientPort(clientPort int) (*Client, error) {
	for _, c := range ClientSet {
		for _, m := range c.Map {
			if m.LocalPort == clientPort {
				return c, nil
			}
		}
	}
	return nil, fmt.Errorf("Client Conn not found")
}

func CheckClientAlive() {
	defer utils.RecoverAndLog()
	for {
		time.Sleep(2 * time.Second)
		for _, c := range ClientSet {
			isDisconnect := false
			// if err := c.Conn.SetWriteDeadline(time.Now().Add(2 * time.Second)); err != nil {
			// 	isDisconnect = true
			// }
			func(c *Client) {
				ccm, ok := CCMList[c.Name]
				if !ok {
					return
				}
				ccm.Cmdrwlock.Lock()
				defer ccm.Cmdrwlock.Unlock()
				// utils.PrintDataAsJson(c.Name)
				// if _, err := c.Conn.Read([]byte{}); err != nil {
				// 	isDisconnect = true
				// }
				if err := common.SendCmd(c.Conn, utils.SerilizeData(common.ServerCmd{Type: "ping"})); err != nil {
					isDisconnect = true
				}
				
			}(c)

			if isDisconnect {
				log.Println("Client ", c.Name, " disconnected")
				for _, ccm := range CCMList {
					if ccm.ClientName == c.Name {
						quitAgent.Publish(ccm.ClientName, "quit")
						delete(CCMList, ccm.ClientName)
					}
				}
				delete(ClientSet, c.Name)
				c.Conn.Close()
				delete(CCMList, c.Name)
			}
		}

		// log.Println("keep alive")
	}
}
