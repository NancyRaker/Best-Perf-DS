/*
@Time : 2019/8/29 16:47 
@Author : yanKoo
@File : send_https_request
@Software: GoLand
@Description:
*/
package utils

import (
	"net/http"
)

const (
	POST = "POST"
	GET  = "GET"
	Account_NAME = "account-name"
	Account_ID = "account-id"
)

type WebGateWay struct {
	Url string
}

type DeviceImportReq struct {
	Imeis     []string `json:"imeis"`
	AccountId int32    `json:"account_id"`
}

// 用来接收改变设备的调度员
type DispatcherAdd struct {
	AccountName string `json:"account_name"`
	AccountId   uint   `json:"account_id"`
	CreatorId   uint   `json:"creator_id"`
}

func (w WebGateWay) ChangeDispatcherServerAddr(params map[string]string) bool{
	return w.get(params)
}

func (w WebGateWay) get(params map[string]string) bool {
	var (
		client = &http.Client{}
	)

	resp, err := client.Get(w.Url + joinParam(params))
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	//body, err := ioutil.ReadAll(resp.Body)
	//if err != nil {
	//	return false
	//}
	//if err := json.Unmarshal(body, res); err != nil {
	//	return false
	//}
	return false
}

func joinParam(params map[string]string) string {
	var res = "?"
	for key, value := range params {
		res = res + key + "=" + value + "&"
	}
	return res[:len(res)-1]
}
