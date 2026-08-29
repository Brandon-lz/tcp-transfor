package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	_ "net/http/pprof" // 通过 init 注册到 DefaultServeMux

	"github.com/Brandon-lz/tcp-transfor/client/config"
	translocaltcp "github.com/Brandon-lz/tcp-transfor/client/trans_local_tcp"
	"github.com/Brandon-lz/tcp-transfor/utils"
)

// startPprof 启动 pprof 调试端点（失败不影响主流程）
func startPprof(addr string) {
	go func() {
		log.Printf("pprof listening on %s", addr)
		if err := http.ListenAndServe(addr, nil); err != nil {
			log.Printf("pprof server stopped: %v", err)
		}
	}()
}

func main() {
	startPprof("localhost:6061")
	log.Println("client start")
	initLog()
	defer utils.RecoverAndLog()

	if err := config.LoadConfig(); err != nil {
		log.Fatalf("load client config failed: %v", err)
	}
	fmt.Println("config loaded")
	for {
		translocaltcp.CommunicateToServer() // block, until fail
		time.Sleep(time.Second * 2)
	}
}

func initLog() *os.File {
	fileName := "sys.log"
	var logFile *os.File
	_, err := os.Stat(fileName)
	if os.IsNotExist(err) {
		logFile, err = os.Create("sys.log")
		if err != nil {
			log.Fatal(err)
		}
	} else {
		logFile, err = os.OpenFile(fileName, os.O_APPEND|os.O_WRONLY, 0600)
		if err != nil {
			log.Fatal(err)
		}
	}
	logOutput := io.MultiWriter(os.Stdout, logFile)
	log.SetOutput(logOutput)
	return logFile
}

// GOOS=windows GOARCH=amd64 go build .
// CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build .
