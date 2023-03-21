package util

import (
	"io"
	"net/http"
	"net/url"
	"os"
	"time"

	"lng_fetch_us/config"
)

var cfg = config.Cfg

// MMCFToWM3 MMCF -> 万方
func MMCFToWM3(value int64) int64 {
	return value * 28316850 / 100000000
}

// FetchDataByDate 获取某天的数据
func GetDataFile(key string) (file *os.File, err error) {
	var fileUrl string
	switch key {
	case "inject":
		fileUrl = cfg.InjectionUrl
	case "withdraw":
		fileUrl = cfg.WithdrawUrl
	default:
		fileUrl = cfg.InjectionUrl
	}

	client := &http.Client{
		Timeout: time.Duration(10) * time.Second,
	}

	// 生产环境 有代理
	// log.Println("current proxy: " + cfg.Proxy)
	if len(cfg.Proxy) > 0 {
		u, _ := url.Parse(cfg.Proxy)
		t := &http.Transport{
			MaxIdleConns:    10,
			MaxConnsPerHost: 10,
			IdleConnTimeout: time.Duration(time.Second) * 15,
			Proxy:           http.ProxyURL(u),
		}
		client.Transport = t
	}

	method := "GET"
	req, err := http.NewRequest(method, fileUrl, nil)
	if err != nil {
		return nil, err
		return
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
		return
	}
	defer resp.Body.Close()

	file, err = os.Create(key + ".xls")
	if err != nil {
		return nil, err
	}

	_, err = io.Copy(file, resp.Body)
	if err != nil {
		return nil, err
	}

	return file, nil
}
