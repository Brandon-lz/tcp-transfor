package main

import (
	"io"
	"log"
	"net/http"
	"os"

	_ "net/http/pprof" // 通过 init 注册到 DefaultServeMux

	"github.com/Brandon-lz/tcp-transfor/server/config"
	toclient "github.com/Brandon-lz/tcp-transfor/server/to_client"
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
	startPprof(":6060")
	defer initLog().Close()
	defer utils.RecoverAndLog()
	config.LoadConfig()
	go toclient.CheckClientAlive()
	toclient.ListenClientConn()
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

// CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build .
