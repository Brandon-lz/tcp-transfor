package common



type ServerCmd struct {
	Id   int         `json:"id"`
	Type string      `json:"type"`
	Data interface{} `json:"data"`
}


type NewConnCreateRequestMessage struct {
	ConnId     int `json:"conn-id"` // 服务端-本客户端之间有多个连接，每个连接都有唯一的conn-id，拿着conn-id返回给服务端去注册新连接
	LocalHost string `json:"local-host"`
	LocalPort  int `json:"local-port"`
	ServerPort int `json:"server-port"`
}

// ResponseToServer 客户端对服务端指令的应答。服务端的读取已异步化（readLoop 分派），
// 所以应答必须带上指令的 Id，服务端才能把并发请求的应答匹配回各自的等待者。
type ResponseToServer struct {
	Id   int         `json:"id"`
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data"`
}
