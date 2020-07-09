/*
@Time : 2019/8/29 16:47 
@Author : yanKoo
@File : send_https_request
@Software: GoLand
@Description:
*/
package utils

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"web-api/logger"
	"web-api/model"
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

func (w WebGateWay) GetDispatcherServerAddr(params map[string]string) *model.WebGateWayResp {
	return w.get(params)
}

func (w WebGateWay) TransDevicePost(data *DeviceImportReq) bool {
	jsonStr, _ := json.Marshal(data)
	return w.post(jsonStr)
}

func (w WebGateWay) ImportDevicePost(data *DeviceImportReq) bool {
	jsonStr, _ := json.Marshal(data)
	return w.post(jsonStr)
}

func (w WebGateWay) CreateDispatcherPost(data *DispatcherAdd) bool {
	jsonStr, _ := json.Marshal(data)
	return w.post(jsonStr)
}

func (w WebGateWay) post(jsonStr []byte) bool {
	req, err := http.NewRequest(POST, w.Url, bytes.NewBuffer(jsonStr))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		logger.Debugf("WebGateWay POST err:%+v", err)
		return false
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		return true
	} else {
		return false
	}
}

func (w WebGateWay) get(params map[string]string) *model.WebGateWayResp {
	var (
		res = &model.WebGateWayResp{}
		client = &http.Client{}
	)

	resp, err := client.Get(w.Url + joinParam(params))
	if err != nil {
		return nil
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil
	}
	if err := json.Unmarshal(body, res); err != nil {
		return nil
	}
	return res
}

func joinParam(params map[string]string) string {
	var res = "?"
	for key, value := range params {
		res = res + key + "=" + value + "&"
	}
	return res[:len(res)-1]
}
