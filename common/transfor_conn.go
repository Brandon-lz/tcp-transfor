package common

import (
	"fmt"
	"io"
	"net"

	"github.com/Brandon-lz/tcp-transfor/utils"
)


func TransForConnData(src net.Conn, dst net.Conn) {
	defer utils.RecoverAndLog()
	defer src.Close()
	defer dst.Close()

	quit := make(chan bool)
	go func() {
		defer utils.RecoverAndLog(func(err error) { quit <- true })
		for {
			_, err := io.Copy(dst, src)
			if err != nil {
				panic(fmt.Errorf("Failed to copy data from %s to %s: %v\n", src.RemoteAddr(), dst.RemoteAddr(), utils.WrapErrorLocation(err)))
			}

		}
	}()

trans:
	for {
		select {
		case <-quit:
			break trans
		default:
			_, err := io.Copy(src, dst)
			if err != nil {
				panic(fmt.Errorf("Failed to copy data from %s to %s: %v\n", dst.RemoteAddr(), src.RemoteAddr(), utils.WrapErrorLocation(err)))
			}
		}
	}

}
