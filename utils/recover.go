package utils

import (
	"fmt"
	"log/slog"
	"os"
	"runtime/debug"
)

// RecoverAndLog recovers a panic: logs it to slog and panic.log, then calls the callbacks.
// 推荐在windows下使用，linux可以直接查看控制台输出查找问题
func RecoverAndLog(callbacks ...func(err error)) {
	r := recover()
	if r == nil {
		return
	}
	var err error
	var msg string
	if e, ok := r.(error); ok {
		err = e
		msg = "未知异常"
	} else {
		err = fmt.Errorf("%v", r)
		msg = "未知panic"
	}
	stack := fmt.Sprintf("%s：%v\n%s\n", msg, err, string(debug.Stack()))

	slog.Error("发生panic", "err", err)

	if f, ferr := os.OpenFile("panic.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644); ferr == nil {
		if _, werr := f.WriteString(stack); werr != nil {
			slog.Error("写入panic日志文件失败", "err", werr)
		}
		f.Close()
	}

	for _, cb := range callbacks {
		cb(err)
	}
}

// usage
// defer RecoverAndLog()
// defer RecoverAndLog(func(err error){quit <- true})
