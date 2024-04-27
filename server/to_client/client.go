package toclient

import (
	"fmt"
	"net"
)

var ClientSet = make(map[string]*Client)

type Client struct {
	Name string `json:"name"`
	Conn net.Conn
	SubConn map[int]net.Conn
	Map  []struct {
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
					return fmt.Errorf("Server port: %d already exists", p)
				}
			}
		}
	}

	ClientSet[client.Name] = client
	return nil
}


func getClientByName(name string) (*Client, error) {
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
