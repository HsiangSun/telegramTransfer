package telegram

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"telgramTransfer/crypt"
	"testing"
	"time"
)

func TestApi(t *testing.T) {

	var orderId = "p71584121t1676703144401"

	auth := crypt.TokenGenerate()

	client := http.Client{Timeout: 5 * time.Second}

	//var apiUrl = config.Apic.Url

	req, _ := http.NewRequest("GET", fmt.Sprintf("%s/botapi/Xc/Order?id=%s", "https://fourpay-intest.ncjimmy.com", orderId), nil)
	req.Header = http.Header{
		"Authorization": {auth},
	}

	resp, err := client.Do(req)

	if err != nil {
		fmt.Printf("查询订单信息失败:%s", err.Error())
	}

	rspBytes, _ := io.ReadAll(resp.Body)

	//fmt.Println("RESPONSE:" + string(rspBytes))

	var rsp ApiRspCode

	jerr := json.Unmarshal(rspBytes, &rsp)
	if jerr != nil {
		fmt.Printf("查询订单响应错误:%s,api msg:%s", string(rspBytes), string(rspBytes))
	}

	if rsp.Code != 0 {
		fmt.Printf("error:%s", rsp.Msg)
	}

	var rspData ApiRsp

	jerr = json.Unmarshal(rspBytes, &rspData)
	if jerr != nil {
		fmt.Printf("理论上不会出现错误:%s,api msg:%s", string(rspBytes), string(rspBytes))
	}
}
