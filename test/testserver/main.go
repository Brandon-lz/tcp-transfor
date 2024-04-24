package main

import (
	"fmt"
	"net/http"
)

// helloHandler 处理函数，返回 "Hello, World!"
func helloHandler(w http.ResponseWriter, r *http.Request) {
    fmt.Fprintln(w, "Hello, World!")
}

func main() {
    http.HandleFunc("/", helloHandler) // 设置路由

    fmt.Println("Starting server at port 8080") // 服务器将在8080端口上监听
    if err := http.ListenAndServe(":8080", nil); err != nil { // 启动服务器
        panic(err)
    }
}
