package utils

import (
	"encoding/json"
	"log"
)

func DeSerializeData[T interface{}](source any, target *T) T { // target必须为指针类型
    var jsonData []byte
    var err error

    if sourceString, isString := source.(string); isString {
        jsonData = []byte(sourceString)
    } else if sourceBytes, isBytes := source.([]byte); isBytes {
        jsonData = sourceBytes
    } else {
        jsonData, err = json.Marshal(source)
        if err != nil {
            log.Printf("JSON序列化失败: %s", WrapErrorLocation(err))
            panic(err)
        }
    }

    err = json.Unmarshal(jsonData, target)
    if err != nil {
        log.Printf("JSON反序列化失败: %s", WrapErrorLocation(err))
        panic(err)
    }
    return *target
}



