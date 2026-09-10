package core

import (
	"fmt"
	"net"
	"testing"
	"time"

	"github.com/Brandon-lz/tcp-transfor/common"
	"github.com/Brandon-lz/tcp-transfor/utils"
)

// TestPingAnsweredWhileNewConnRequestInFlight 验证：某条新建连接还在慢慢拨号时，
// 主连接上的 ping 依然能被立刻应答。
//
// 旧实现把 new-conn-request 同步放在读命令循环里处理，慢拨号会吃掉整个心跳窗口，
// server 侧等不到 pong 就把整条隧道判死（所有映射端口一起消失，线上表现为掉线十几秒）。
// 这个测试就是那条回归线：pong 必须在 1.5s 内回来，而不是等拨号结束。
func TestPingAnsweredWhileNewConnRequestInFlight(t *testing.T) {
	origDial := dialLocalPort
	defer func() { dialLocalPort = origDial }()
	// 模拟目标不可达时客户端被卡住的拨号：3 秒后才失败
	dialLocalPort = func(ip string, port int, timeout time.Duration) (net.Conn, error) {
		time.Sleep(3 * time.Second)
		return nil, fmt.Errorf("simulated slow dial failure")
	}

	lis, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	defer lis.Close()

	go func() {
		c, err := lis.Accept()
		if err != nil {
			return
		}
		ListenServerCmd(c, common.NewLineReader(c))
	}()

	conn, err := net.Dial("tcp", lis.Addr().String())
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	defer conn.Close()
	lr := common.NewLineReader(conn)

	// 1) 下发一个 new-conn-request，客户端随即进入慢拨号
	req := common.ServerCmd{
		Id:   1,
		Type: "new-conn-request",
		Data: common.NewConnCreateRequestMessage{ConnId: 1, LocalHost: "127.0.0.1", LocalPort: 6553, ServerPort: 19999},
	}
	if err := common.SendCmd(conn, utils.SerilizeData(req)); err != nil {
		t.Fatalf("send new-conn-request: %v", err)
	}

	// 2) 200ms 后发 ping：拨号还在进行中，pong 必须马上回来
	time.Sleep(200 * time.Millisecond)
	if err := common.SendCmd(conn, utils.SerilizeData(common.ServerCmd{Type: "ping"})); err != nil {
		t.Fatalf("send ping: %v", err)
	}

	deadline := time.Now().Add(3 * time.Second)
	_ = conn.SetReadDeadline(deadline)
	pingSent := time.Now()
	for {
		line, err := lr.ReadOne()
		if err != nil {
			t.Fatalf("没等到 pong（读失败: %v）——读命令循环很可能被新建连接的处理堵住了", err)
		}
		if string(line) == "pong" {
			if d := time.Since(pingSent); d > 1500*time.Millisecond {
				t.Fatalf("pong 迟到了 %v，说明读命令循环被新建连接的处理阻塞了", d)
			}
			return
		}
		// 慢拨号最终失败会先回一条 500 ack，跳过它继续找 pong
	}
}
