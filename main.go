package main

import (
	"fmt"
	"github.com/go-resty/resty/v2"
	"github.com/mdp/qrterminal/v3"
	"golang.design/x/clipboard"
	"os"
	"strconv"
	"time"
)

type KeyGeneRes struct {
	Data struct {
		Code   int    `json:"code"`
		Unikey string `json:"unikey"`
	} `json:"data"`
	Code int `json:"code"`
}
type QRRes struct {
	Code int `json:"code"`
	Data struct {
		Qrurl string `json:"qrurl"`
		Qrimg string `json:"qrimg"`
	} `json:"data"`
}
type CookieRes struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Cookie  string `json:"cookie"`
}
type NeteaseClient struct {
	client *resty.Client
}

const (
	Endpoint = "https://ncm.supaku.cn"
)

func main() {
	client := NewClient()
	nc := &NeteaseClient{client: client}

	key := nc.getKeyGeneRes()
	nc.getQRRes(key)
	nc.checkQrScan(key)

	os.Exit(0)

}

func NewClient() *resty.Client {
	client := resty.New()
	client.SetBaseURL(Endpoint)
	client.SetHeader("Content-Type", "application/json")
	client.SetHeader("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/000000000 Safari/537.36")
	client.SetTimeout(10 * time.Second)
	return client
}

func (c *NeteaseClient) getKeyGeneRes() string {
	kr := &KeyGeneRes{}
	_, err := c.client.R().SetResult(kr).SetQueryParams(map[string]string{
		"t": unixTime(),
	}).Get("/login/qr/key")
	if err != nil {
		return ""
	}
	fmt.Println(kr.Data.Unikey)
	return kr.Data.Unikey
}

func (c *NeteaseClient) getQRRes(key string) *QRRes {
	qr := &QRRes{}
	config := qrterminal.Config{
		Level:     qrterminal.M,
		Writer:    os.Stdout,
		BlackChar: qrterminal.BLACK,
		WhiteChar: qrterminal.WHITE,
		QuietZone: 1,
	}
	_, err := c.client.R().SetResult(qr).SetQueryParams(map[string]string{
		"key": key,
		"t":   unixTime(),
	}).Get("/login/qr/create")
	if err != nil {
		return nil
	}
	qrterminal.GenerateWithConfig(qr.Data.Qrurl, config)
	return qr
}

func (c *NeteaseClient) checkQrScan(key string) {
	cr := &CookieRes{}
	for {
		_, err := c.client.R().SetResult(cr).SetQueryParams(map[string]string{
			"key": key,
			"t":   unixTime(),
		}).Get("/login/qr/check")
		if err != nil {
			fmt.Println("Error occurred while accessing /login/qr/check: ", err)
			time.Sleep(5 * time.Second)
			break
		}

		switch cr.Code {
		case 802:
			fmt.Println("已扫描，请确认")
		case 803:
			WriteToClipboard([]byte(cr.Cookie))
			fmt.Println("已确认，Cookie已复制到剪贴板")
			return
		}

		time.Sleep(1 * time.Second)

	}
}

func unixTime() string {
	return strconv.FormatInt(time.Now().Unix(), 10)
}

func WriteToClipboard(cookie []byte) {
	clipboard.Write(clipboard.FmtText, cookie)
}
