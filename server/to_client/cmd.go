package toclient

import (
	"log"

	"github.com/Brandon-lz/tcp-transfor/common"
	"github.com/Brandon-lz/tcp-transfor/utils"
)

// cmdToClientGetNewConn 通过主连接通知客户端为本次用户连接新建本地连接
// 与心跳（CheckClientAlive）共享 ckmu 锁，保证主连接上的命令收发互斥
func cmdToClientGetNewConn(ccm *clientConnManager, connId int, clientLocalHost string, LocalPort, ServerPort int) error {
	ccm.ckmu.Lock()
	defer ccm.ckmu.Unlock()
	sercmd := common.ServerCmd{
		Type: "new-conn-request",
		Data: common.NewConnCreateRequestMessage{
			ConnId:     connId,
			LocalHost:  clientLocalHost,
			LocalPort:  LocalPort,
			ServerPort: ServerPort,
		},
	}
	err := common.SendCmd(ccm.ClientConn, utils.SerilizeData(sercmd))
	log.Println("send new conn request to client", utils.PrintDataAsJson(sercmd))
	if err != nil {
		return utils.WrapErrorLocation(err, "cmdToClientGetNewConn")
	}

	// 等待客户端的 ack（复用带缓冲的行读取器，避免破坏主连接上的读缓冲）
	if _, err = ccm.LR.ReadOne(); err != nil {
		log.Println("read ack error", utils.WrapErrorLocation(err))
		return err
	}

	return nil
}
