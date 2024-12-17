package main

import (
	"io"
	"log"
	"os"

	"github.com/Brandon-lz/tcp-transfor/server/config"
	toclient "github.com/Brandon-lz/tcp-transfor/server/to_client"
	"github.com/Brandon-lz/tcp-transfor/utils"

    "net/http"
	_ "net/http/pprof"
)



func main(){
    go func() {
        log.Println(http.ListenAndServe(":6060", nil))
    }()
    defer initLog().Close()
	defer utils.RecoverAndLog()
    config.LoadConfig()
    utils.AESInit()
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
