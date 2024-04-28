package common



type ServerCmd struct {
	Id   int         `json:"id"`
	Type string      `json:"type"`
	Data interface{} `json:"data"`
}


type NewConnCreateRequestMessage struct {
	ConnId     int `json:"conn-id"` // 服务端-本客户端之间有多个连接，每个连接都有唯一的conn-id，拿着conn-id返回给服务端去注册新连接
	LocalPort  int `json:"local-port"`
	ServerPort int `json:"server-port"`
}
