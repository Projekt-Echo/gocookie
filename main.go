//go:generate goversioninfo
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

const (
	Endpoint = "https://ncm.supaku.cn"
)

func main() {
	client := NewClient()

	key := getKeyGeneRes(client)
	getQRRes(client, key)
	checkQrScan(client, key)

}

func NewClient() *resty.Client {
	client := resty.New()
	client.SetBaseURL(Endpoint)
	client.SetHeader("Content-Type", "application/json")
	client.SetHeader("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/000000000 Safari/537.36")
	client.SetTimeout(10 * time.Second)
	return client
}

func getKeyGeneRes(client *resty.Client) string {
	kr := &KeyGeneRes{}
	client.R().SetResult(kr).SetQueryParams(map[string]string{
		"t": unixTime(),
	}).Get("/login/qr/key")
	fmt.Println(kr.Data.Unikey)
	return kr.Data.Unikey
}

func getQRRes(client *resty.Client, key string) *QRRes {
	qr := &QRRes{}
	config := qrterminal.Config{
		Level:     qrterminal.M,
		Writer:    os.Stdout,
		BlackChar: qrterminal.BLACK,
		WhiteChar: qrterminal.WHITE,
		QuietZone: 1,
	}
	client.R().SetResult(qr).SetQueryParams(map[string]string{
		"key": key,
		"t":   unixTime(),
	}).Get("/login/qr/create")
	qrterminal.GenerateWithConfig(qr.Data.Qrurl, config)
	return qr
}

func checkQrScan(client *resty.Client, key string) {
	cr := &CookieRes{}
	for range time.Tick(1 * time.Second) {
		_, err := client.R().SetResult(cr).SetQueryParams(map[string]string{
			"key": key,
			"t":   unixTime(),
		}).Get("/login/qr/check")
		if err != nil {
			fmt.Println("Error occurred while accessing /login/qr/check: ", err)
			time.Sleep(5 * time.Second)
			continue
		}
		if cr.Code == 802 {
			fmt.Println("已扫描，请确认")
			continue
		}
		if cr.Code == 803 {
			clipboard.Write(clipboard.FmtText, []byte(cr.Cookie))
			break
		}
	}
}

func unixTime() string {
	return strconv.FormatInt(time.Now().Unix(), 10)
}
