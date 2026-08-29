package utils

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"net/url"
	"time"
)

const reqTimeOut = 30 * time.Second

// 共享的 http.Client（Timeout 覆盖整个请求生命周期，同时复用底层 TCP 连接）
var httpClient = &http.Client{Timeout: reqTimeOut}

func PostRequest(urlstr string, body string) (res *http.Response, err error) {
	var b io.Reader
	if body == "" {
		b = bytes.NewBuffer([]byte{})
	} else {
		b = bytes.NewBuffer([]byte(body))
	}
	req, err := http.NewRequestWithContext(context.Background(), "POST", urlstr, b)
	if err != nil {
		return
	}
	req.Header.Set("Content-Type", "application/json")
	res, err = httpClient.Do(req)
	return
}

type QueryParam struct {
	Name  string
	Value string
}

func GetRequest(urlstr string, query ...QueryParam) (res *http.Response, err error) {
	values := url.Values{}
	for _, q := range query {
		values.Add(q.Name, q.Value)
	}
	if len(values) > 0 {
		urlstr += "?" + values.Encode()
	}
	req, err := http.NewRequestWithContext(context.Background(), "GET", urlstr, nil)
	if err != nil {
		return
	}
	res, err = httpClient.Do(req)
	return
}

func PutRequest(urlstr string, body string) (res *http.Response, err error) {
	var b io.Reader
	if body == "" {
		b = bytes.NewBuffer([]byte{})
	} else {
		b = bytes.NewBuffer([]byte(body))
	}
	req, err := http.NewRequestWithContext(context.Background(), "PUT", urlstr, b)
	if err != nil {
		return
	}
	req.Header.Set("Content-Type", "application/json")
	res, err = httpClient.Do(req)
	return
}
