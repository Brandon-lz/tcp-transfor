package utils

import "encoding/json"

func SerilizeData(data interface{}) []byte {
	bytes, err := json.Marshal(data)
	if err != nil {
		panic(err)
	}
	return bytes
}
