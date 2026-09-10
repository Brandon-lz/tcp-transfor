package toclient

import (
	"fmt"
	"log"
	"time"

	"github.com/Brandon-lz/tcp-transfor/common"
	"github.com/Brandon-lz/tcp-transfor/utils"
)

// cmdToClientGetNewConn 通过主连接通知客户端为本次用户连接新建本地连接。
// 锁只覆盖发送；应答由 readLoop 按 Id 分派，等待期间不持有 ckmu，
// 因此某个客户端的慢应答不会堵住别的新建连接或心跳。
func cmdToClientGetNewConn(ccm *clientConnManager, connId int, clientLocalHost string, LocalPort, ServerPort int) error {
	// 先注册等待者再发送，避免应答先于注册到达而丢失
	waiter := ccm.addAck(connId)
	defer ccm.removeAck(connId)

	sercmd := common.ServerCmd{
		Id:   connId,
		Type: "new-conn-request",
		Data: common.NewConnCreateRequestMessage{
			ConnId:     connId,
			LocalHost:  clientLocalHost,
			LocalPort:  LocalPort,
			ServerPort: ServerPort,
		},
	}

	ccm.ckmu.Lock()
	err := common.SendCmd(ccm.ClientConn, utils.SerilizeData(sercmd))
	ccm.ckmu.Unlock()
	log.Println("send new conn request to client", utils.PrintDataAsJson(sercmd))
	if err != nil {
		return utils.WrapErrorLocation(err, "cmdToClientGetNewConn")
	}

	select {
	case resp := <-waiter:
		if resp.Code != 200 {
			return fmt.Errorf("client %s refused new conn id:%d: %s", ccm.ClientName, connId, resp.Msg)
		}
		return nil
	case <-time.After(ackTimeout):
		return fmt.Errorf("timeout waiting new-conn-request ack from client %s id:%d", ccm.ClientName, connId)
	}
}
