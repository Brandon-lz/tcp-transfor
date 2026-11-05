package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"text/tabwriter"
	"time"

	"github.com/Brandon-lz/tcp-transfor/server/config"
)

// runClientsCmd 是 `server clients` 子命令：向本机状态接口取一次在线客户端列表并打印。
func runClientsCmd(args []string) int {
	fs := flag.NewFlagSet("clients", flag.ContinueOnError)
	addr := fs.String("addr", "", "状态接口地址，默认 127.0.0.1:<ui-port>")
	asJSON := fs.Bool("json", false, "直接输出原始 JSON")
	if err := fs.Parse(args); err != nil {
		return 2
	}

	// LoadConfig 会往 stdout 打配置内容，先把它挪到 stderr，避免污染表格输出。
	// 只有没给 -addr 时才需要读配置；配置缺失也不该让命令直接崩掉。
	target := *addr
	if target == "" {
		target = fmt.Sprintf("127.0.0.1:%d", loadUIPort())
	}

	cli := &http.Client{Timeout: 5 * time.Second}
	resp, err := cli.Get("http://" + target + "/api/clients")
	if err != nil {
		fmt.Fprintf(os.Stderr, "连接状态接口失败: %v\n", err)
		fmt.Fprintf(os.Stderr, "确认 server 正在运行，且 config.toml 的 ui-port（当前 %d）没被改过\n", uiPort())
		return 1
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Fprintf(os.Stderr, "读取响应失败: %v\n", err)
		return 1
	}
	if resp.StatusCode != http.StatusOK {
		fmt.Fprintf(os.Stderr, "状态接口返回 %d: %s\n", resp.StatusCode, string(body))
		return 1
	}

	if *asJSON {
		fmt.Println(string(body))
		return 0
	}

	var st StatusResp
	if err := json.Unmarshal(body, &st); err != nil {
		fmt.Fprintf(os.Stderr, "解析响应失败: %v\n", err)
		return 1
	}
	printClientsTable(st)
	return 0
}

// loadUIPort 读取 config.toml 里的 ui-port；读不到就回退默认端口，不 panic。
func loadUIPort() (port int) {
	savedStdout := os.Stdout
	os.Stdout = os.Stderr
	defer func() {
		os.Stdout = savedStdout
		if r := recover(); r != nil {
			fmt.Fprintf(os.Stderr, "读取 config.toml 失败(%v)，使用默认端口 %d\n", r, defaultUIPort)
			port = defaultUIPort
		}
	}()
	config.LoadConfig()
	return uiPort()
}

func printClientsTable(st StatusResp) {
	fmt.Printf("服务端口: %d\n", st.ServerPort)
	fmt.Printf("在线客户端: %d\n\n", st.Count)
	if st.Count == 0 {
		fmt.Println("(当前没有客户端在线)")
		return
	}

	tw := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(tw, "CLIENT\tREMOTE\tLISTEN\t->\tLOCAL-TARGET")
	for _, c := range st.Clients {
		if len(c.Map) == 0 {
			fmt.Fprintf(tw, "%s\t%s\t-\t-\t-\n", c.Name, c.RemoteAddr)
			continue
		}
		for i, m := range c.Map {
			name, remote := "", ""
			if i == 0 {
				name, remote = c.Name, c.RemoteAddr
			}
			host := m.LocalHost
			if host == "" {
				host = "127.0.0.1"
			}
			fmt.Fprintf(tw, "%s\t%s\t:%d\t->\t%s:%d\n", name, remote, m.ServerPort, host, m.LocalPort)
		}
	}
	_ = tw.Flush()
}
