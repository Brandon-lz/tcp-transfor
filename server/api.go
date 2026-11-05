package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/Brandon-lz/tcp-transfor/server/config"
	toclient "github.com/Brandon-lz/tcp-transfor/server/to_client"
)

// defaultUIPort 是状态接口的默认端口
const defaultUIPort = 38081

func uiPort() int {
	if config.Config.UIPort > 0 {
		return config.Config.UIPort
	}
	return defaultUIPort
}

// StatusResp 是 /api/clients 的响应体
type StatusResp struct {
	ServerPort int                    `json:"server-port"`
	Count      int                    `json:"count"`
	Clients    []toclient.ClientView  `json:"clients"`
}

// startStatusAPI 启动只读状态接口，只监听回环地址，
// 供 `server clients` 子命令（或本机 curl）查看在线客户端与端口映射。
func startStatusAPI() {
	addr := fmt.Sprintf("127.0.0.1:%d", uiPort())

	mux := http.NewServeMux()
	mux.HandleFunc("/api/clients", func(w http.ResponseWriter, r *http.Request) {
		clients := toclient.SnapshotClients()
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.Header().Set("Cache-Control", "no-store")
		_ = json.NewEncoder(w).Encode(StatusResp{
			ServerPort: config.Config.Port,
			Count:      len(clients),
			Clients:    clients,
		})
	})

	srv := &http.Server{
		Addr:              addr,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}
	go func() {
		log.Printf("status api listening on %s", addr)
		if err := srv.ListenAndServe(); err != nil {
			log.Printf("status api stopped: %v", err)
		}
	}()
}
