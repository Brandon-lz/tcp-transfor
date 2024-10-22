package utils

import (
	"encoding/json"
	"fmt"
	"runtime"
	"time"
)

// 定义一些颜色的 ANSI 转义码
const (
	Reset  = "\033[0m"  // 重置颜色
	Red    = "\033[31m" // 红色
	Green  = "\033[32m" // 绿色
	Yellow = "\033[33m" // 黄色
	Blue   = "\033[34m" // 蓝色
	Purple = "\033[35m" // 紫色
	Cyan   = "\033[36m" // 青色
	White  = "\033[37m" // 白色
)

var loc, _ = time.LoadLocation("Asia/Shanghai")

func PrintDataAsJson(m interface{}, doPrint ...bool) (res string) { // 默认打印
	d, err := json.Marshal(m)
	if err != nil {
		panic(fmt.Sprintf("PrintDataAsJson Error:%+v", err))
	}
	res = string(d)
	if len(doPrint) == 0 || doPrint[0] {
		fmt.Println(Green, "-------PrintDataAsJson at:", time.Now().In(loc), getCodeLine(2), "\n", Reset, res)
	}
	return
}

// 调用 GetCodeLine 获取当前代码所在的行数
func getCodeLine(skip int) string {
	_, file, line, _ := runtime.Caller(skip)
	return fmt.Sprintf("%s:%d", file, line)
}
