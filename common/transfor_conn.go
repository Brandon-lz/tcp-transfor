package common

import (
	"fmt"
	"io"
	"net"
	"time"

	"github.com/Brandon-lz/tcp-transfor/utils"
)

const forwardTimeout = 8 * time.Hour

type connOption func(net.Conn)

// nodelayer 由 *net.TCPConn 实现。这里用接口断言而不是 *net.TCPConn 断言：
// 一旦某个连接被包装成自定义 net.Conn 类型，*net.TCPConn 断言会静默失败，
// Nagle 设置就被无声跳过（旧版 ConnLocked 就踩过这个坑）。
// 注：Go 的 net 包对 TCP 连接默认就设了 TCP_NODELAY=1，这里只是把意图写明。
type nodelayer interface{ SetNoDelay(bool) error }

func noDelay(c net.Conn) {
	if nd, ok := c.(nodelayer); ok {
		// 转发小数据包时关闭 Nagle，避免小包被攒批延迟
		_ = nd.SetNoDelay(true)
	}
}

// TransForConnDataServer 服务端：user conn <-> client sub conn 双向转发
func TransForConnDataServer(user2serverConn net.Conn, server2clientConn net.Conn) {
	defer utils.RecoverAndLog()
	defer user2serverConn.Close()
	defer server2clientConn.Close()
	noDelay(user2serverConn)
	noDelay(server2clientConn)

	// user -> server
	go func() {
		defer utils.RecoverAndLog()
		defer user2serverConn.Close()
		defer server2clientConn.Close()

		utils.PrintDataAsJson("1------------1-")
		n, err := timedCopy(server2clientConn, user2serverConn, forwardTimeout)
		if err != nil {
			utils.PrintDataAsJson(fmt.Sprintf("1Copy error: %v, bytes copied: %d", err, n))
		}
	}()

	// server -> user
	utils.PrintDataAsJson("2------------1-")
	n, err := timedCopy(user2serverConn, server2clientConn, forwardTimeout)
	if err != nil {
		utils.PrintDataAsJson(fmt.Sprintf("2Copy error: %v, bytes copied: %d", err, n))
	}
}

// TransForConnDataClient 客户端：local conn <-> server sub conn 双向转发
func TransForConnDataClient(local2clientConn net.Conn, client2serverConn net.Conn) {
	defer utils.RecoverAndLog()
	defer local2clientConn.Close()
	defer client2serverConn.Close()
	noDelay(local2clientConn)
	noDelay(client2serverConn)

	// local -> server
	go func() {
		defer utils.RecoverAndLog()
		defer local2clientConn.Close()
		defer client2serverConn.Close()

		utils.PrintDataAsJson("3------------1-")
		n, err := timedCopy(client2serverConn, local2clientConn, forwardTimeout)
		if err != nil {
			utils.PrintDataAsJson(fmt.Sprintf("3Copy error: %v, bytes copied: %d", err, n))
		}
	}()

	// server -> local
	utils.PrintDataAsJson("4------------1-")
	n, err := timedCopy(local2clientConn, client2serverConn, forwardTimeout)
	if err != nil {
		utils.PrintDataAsJson(fmt.Sprintf("4Copy error: %v, bytes copied: %d", err, n))
	}
}

// timedCopy 带闲置超时的双向数据搬运，超时的同时关闭另一方向（任一方向闲置超时视为整条通道结束）
func timedCopy(dst, src net.Conn, idleTimeout time.Duration) (int64, error) {
	src.SetReadDeadline(time.Now().Add(idleTimeout))
	dst.SetWriteDeadline(time.Now().Add(idleTimeout))
	n, err := io.Copy(dst, src)
	// 一端结束后关闭双端，避免另一端 goroutine 卡住
	dst.Close()
	src.Close()
	return n, err
}
