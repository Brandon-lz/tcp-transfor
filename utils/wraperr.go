package utils

import (
    "fmt"
    "runtime"
)


// WrapErrorLocation 包装错误信息，将发生错误的行数信息添加到错误信息中
func WrapErrorLocation(err error, msg ...string) error {
    if err != nil {
        addtionMsg := ""
        for _, m := range msg {
            addtionMsg += (" | " + m)
        }
        return fmt.Errorf("error occurred at %s %s\n\t%w", GetCodeLine(2), addtionMsg, err)
    }
    return nil
}

// 调用 GetCodeLine 获取当前代码所在的行数
func GetCodeLine(skip int) string {
    _, file, line, _ := runtime.Caller(skip)
    return fmt.Sprintf("%s:%d", file, line)
}
