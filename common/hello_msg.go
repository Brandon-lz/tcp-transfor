package common


type HelloMessage struct {
	Type string `json:"type"`          // main or sub
	Client struct {
		Name string `json:"name"`
	} `json:"client"`
	Map []struct {
		LocalHost string `json:"local-host"`
		LocalPort  int `json:"local-port"`
		ServerPort int `json:"server-port"`
	} `json:"map"`
	ConnId int `json:"conn-id"`
}

type HelloRecv struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
}

