package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
)

func main() {
    data := []byte("Hello, world! This is a test.\r\n")
    reader := bufio.NewReader(bytes.NewReader(data))

	buf := bytes.Buffer{}
	for {
		d, err := bufio.NewReader(bytes.NewReader(data)).ReadBytes('\n')
		if err != nil {
			fmt.Println(err)
		}
		buf.Write(d)
		l := len(buf.Bytes())
		if buf.Bytes()[l-2] == '\r' {
			fmt.Println("2222222222",string(buf.Bytes()[:l-2]))
			break
		} 
	}

    // 使用 ReadBytes 直到遇到 '!'
    result, err := reader.ReadBytes('!')
    if err != nil && err != io.EOF {
        fmt.Println("Error:", err)
    }
    fmt.Println("ReadBytes result:", string(result))

    // 使用 ReadString 直到遇到 '.'
    resultString, err := reader.ReadString('.')
    if err != nil && err != io.EOF {
        fmt.Println("Error:", err)
    }
    fmt.Println("ReadString result:", resultString)
}