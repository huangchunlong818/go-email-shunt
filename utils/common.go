package utils

import (
	"bytes"
	"crypto/tls"
	"errors"
	"go.uber.org/zap"
	"io"
	"net/http"
	"strings"
	"time"
)

type ApiCallResult struct {
	Data      []byte
	HttpCode  int
	RspHeader http.Header
}

// 发起http 请求
func ApiCall(method string, url string, body io.Reader, headerOpt map[string]string, timeOut time.Duration) (*ApiCallResult, error) {
	defer func() {
		if err := recover(); err != nil {
			zap.S().Errorf("发起HTTP请求失败 url:%+v, err:%+v", url, err)
		}
	}()
	result := &ApiCallResult{}
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	httpClient := http.Client{
		Transport: tr,
		//设置超时时间
		Timeout: timeOut,
	}
	method = strings.ToUpper(method)
	//如果不是这四种请求方式则不可以
	if method != "POST" && method != "GET" && method != "PUT" && method != "DELETE" {
		return nil, errors.New("request method is invalid ")
	}
	if url == "" {
		return nil, errors.New("url can not be empty string")
	}
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}
	//设置头信息
	for k, v := range headerOpt {
		req.Header.Set(k, v)
	}
	rsp, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer rsp.Body.Close()
	result.HttpCode = rsp.StatusCode
	result.RspHeader = rsp.Header
	buff := bytes.Buffer{}
	n, err := buff.ReadFrom(rsp.Body)
	if err != nil {
		return nil, err
	}
	result.Data = buff.Bytes()[:n]
	return result, nil
}

// 使用泛型的InSlice函数，适用于多种类型
func InSlice[T comparable](needle T, haystack []T) bool {
	for _, v := range haystack {
		if v == needle {
			return true
		}
	}
	return false
}
