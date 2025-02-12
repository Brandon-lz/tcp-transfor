package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"time"

	"github.com/Brandon-lz/tcp-transfor/client/config"
	translocaltcp "github.com/Brandon-lz/tcp-transfor/client/trans_local_tcp"
	"github.com/Brandon-lz/tcp-transfor/utils"

	"net/http"
	_ "net/http/pprof"
)


func main() {
	go func() {
		log.Println(http.ListenAndServe("localhost:6060", nil))
	}()
	log.Println("client start")
	initLog()
	// defer .Close()
	defer utils.RecoverAndLog()

	config.LoadConfig()
	fmt.Println("config loaded")
	utils.AESInit()
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