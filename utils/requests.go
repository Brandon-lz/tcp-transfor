package utils

import (
    "bytes"
    "context"
    "io"
    "net/http"
    "net/url"
    "time"
)

var reqTimeOut = 30 * time.Second

func PostRequest(urlstr string, body string) (res *http.Response, err error) {
    client := &http.Client{}
    var b io.Reader
    if body == "" {
        b = bytes.NewBuffer([]byte{})
    } else {
        b = bytes.NewBuffer([]byte(body))
    }
    req, err := http.NewRequest("POST", urlstr, b)
    if err != nil {
        return
    }
    ctx := context.Background()
    ctx, _ = context.WithTimeout(ctx, reqTimeOut)
    // defer cancel()
    req = req.WithContext(ctx)
    req.Header.Set("Content-Type", "application/json")
    // req.Header.Set("accept", "application/json")
    res, err = client.Do(req)
    return
}

type QueryParam struct {
    Name  string
    Value string
}

func GetRequest(urlstr string, query ...QueryParam) (res *http.Response, err error) {
    client := &http.Client{}

    for i, q := range query {
        if i == 0 {
            urlstr += "?" + url.QueryEscape(q.Name) + "=" + url.QueryEscape(q.Value)
        } else {
            urlstr += "&" + url.QueryEscape(q.Name) + "=" + url.QueryEscape(q.Value)
        }
    }
    req, err := http.NewRequest("GET", urlstr, nil)
    if err != nil {
        return
    }
    ctx := context.Background()
    ctx, _ = context.WithTimeout(ctx, reqTimeOut)
    req = req.WithContext(ctx)
    // req.Header.Set("Content-Type", "application/json")
    // req.Header.Set("accept", "application/json")
    res, err = client.Do(req)
    if err != nil {
        return
    }
    return
}

func PutRequest(urlstr string, body string) (res *http.Response, err error) {
    client := &http.Client{}
    var b io.Reader
    if body == "" {
        // b = nil
        b = bytes.NewBuffer([]byte{})
    } else {
        b = bytes.NewBuffer([]byte(body))
    }
    req, err := http.NewRequest("PUT", urlstr, b)
    if err != nil {
        return
    }
    ctx := context.Background()
    ctx, _ = context.WithTimeout(ctx, reqTimeOut)
    req = req.WithContext(ctx)
    req.Header.Set("Content-Type", "application/json")
    // req.Header.Set("accept", "application/json")
    res, err = client.Do(req)
    return
}
