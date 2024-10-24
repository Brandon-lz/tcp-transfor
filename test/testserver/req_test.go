package main

import (
	"sync"
	"testing"

	"github.com/Brandon-lz/tcp-transfor/utils"
	"github.com/stretchr/testify/assert"
)

func TestRequest(t *testing.T) {
	errTimes := 0
	wg := &sync.WaitGroup{}
	for range 10 {
		wg.Add(1)
		go func(){
			defer wg.Done()
			res, err := utils.GetRequest("http://127.0.0.1:9091")
			assert.NoError(t,err)
			if err!= nil {
				errTimes++
				// t.Log(err)
				return
			}
			t.Log(res.Status)

		}() 
	}

	wg.Wait()

}
