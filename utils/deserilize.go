package utils

import (
	"encoding/json"
	"fmt"
	"log"
)



func DeSerializeData[T interface{}](source any, target *T) (T,error) { // target必须为指针类型
	var jsonData []byte
	var err error

	if sourceString, isString := source.(string); isString {
		jsonData = []byte(sourceString)
	} else if sourceBytes, isBytes := source.([]byte); isBytes {
		jsonData = sourceBytes
	} else {
		jsonData, err = json.Marshal(source)
		if err != nil {
			log.Printf("JSON序列化失败: %s", wrapErrorLocation(err))
			return *target,err
		}
	}
	err = json.Unmarshal(jsonData, target)
	if err != nil {
		log.Printf("JSON反序列化失败: %s\n%s", wrapErrorLocation(err), jsonData)
		return *target,err
	}
	return *target,nil
}


// WrapErrorLocation 包装错误信息，将发生错误的行数信息添加到错误信息中
func wrapErrorLocation(err error, msg ...string) error {
    if err != nil {
        addtionMsg := ""
        for _, m := range msg {
            addtionMsg += (" | " + m)
        }
        return fmt.Errorf("error occurred at %s %s\n\t%w", GetCodeLine(3), addtionMsg, err)
    }
    return nil
}
