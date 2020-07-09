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
	"bo-server/logger"
	"net/http"
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

const POST = "POST"

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

func (w WebGateWay) post (jsonStr []byte) bool {
	logger.Debugf("wg will send url: %s", w.Url)
	req, err := http.NewRequest(POST, w.Url, bytes.NewBuffer(jsonStr))
	logger.Debugln(err)
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
