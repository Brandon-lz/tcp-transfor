package common

import (
	"fmt"
	"io"
	"net"
	"time"

	"github.com/Brandon-lz/tcp-transfor/utils"
)

// func TransForConnData(src *net.TCPConn, dst *net.TCPConn) {
// 	defer utils.RecoverAndLog()
// 	defer src.Close()
// 	defer dst.Close()

// 	quit := make(chan bool)
// 	go func() {
// 		defer utils.RecoverAndLog(func(err error) { quit <- true })
// 		for {
// 			_, err := io.Copy(dst, src)
// 			if err != nil {
// 				panic(fmt.Errorf("Failed to copy data from %s to %s: %v\n", src.RemoteAddr(), dst.RemoteAddr(), utils.WrapErrorLocation(err)))
// 			}
// 		}
// 	}()

// trans:
// 	for {
// 		select {
// 		case <-quit:
// 			break trans
// 		default:
// 			_, err := io.Copy(src, dst)
// 			if err != nil {
// 				panic(fmt.Errorf("Failed to copy data from %s to %s: %v\n", dst.RemoteAddr(), src.RemoteAddr(), utils.WrapErrorLocation(err)))
// 			}
// 		}
// 	}

// }

func TransForConnData(src *net.TCPConn, dst *net.TCPConn) {
	defer utils.RecoverAndLog()
	defer src.Close()
	defer dst.Close()

	// quit := make(chan bool)
	go func() {
		defer utils.RecoverAndLog(func(err error) {
			// quit <- true
		})
		for {
			// src.SetDeadline(time.Now().Add(8 * time.Hour))
			// if count < 9600 {
			// 	userConn.SetReadDeadline(time.Now().Add(time.Duration(count) * 3 * time.Second))
			// } else {
			// 	userConn.SetDeadline(time.Now().Add(8 * time.Hour))
			// }

			_, err := io.Copy(dst, src) // when dst.close, it will panic

			// data, err := common.ReadConn(userConn)
			// if err != nil {
			// 	panic(fmt.Errorf("Failed to copy data from %s to %s: %v\n", userConn.RemoteAddr(), dst.RemoteAddr(), utils.WrapErrorLocation(err)))
			// }
			// if len(data) == 0 {
			// 	log.Println("receive empty data from client, close conn")
			// 	break
			// }
			// _, err = dst.Write(data)

			if err != nil {
				panic(fmt.Errorf("failed to write data to %s: %v", dst.RemoteAddr(), utils.WrapErrorLocation(err)))
			}
		}
	}()
	count := 1

	// trans:
	for {
		// select {
		// case <-quit:
		// break trans
		// default:
		err := src.SetDeadline(time.Now().Add(8 * time.Second))
		if err != nil {
			panic(fmt.Errorf("failed to set deadline for %s: %v", src.RemoteAddr(), utils.WrapErrorLocation(err)))
		}
		err = dst.SetDeadline(time.Now().Add(8 * time.Second))
		if err != nil {
			panic(fmt.Errorf("failed to set deadline for %s: %v", dst.RemoteAddr(), utils.WrapErrorLocation(err)))
		}

		// if count < 9600 {
		// 	src.SetReadDeadline(time.Now().Add(time.Duration(count*60) * time.Second))
		// } else {
		// 	src.SetDeadline(time.Now().Add(8 * time.Hour))
		// }

		count++
		_, err = io.Copy(src, dst)
		if err != nil {
			panic(fmt.Errorf("failed to copy data from %s to %s: %v", dst.RemoteAddr(), src.RemoteAddr(), utils.WrapErrorLocation(err)))
		}
		// }
	}
}
