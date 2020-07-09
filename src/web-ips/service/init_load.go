/*
@Time : 2019/8/27 14:40 
@Author : yanKoo
@File : engine
@Software: GoLand
@Description: 加载用户
*/
package service

import (
	"encoding/json"
	"io/ioutil"
	"web-ips/bean"
	"web-ips/dao/customer"
	"web-ips/dao/device"
	"web-ips/logger"
	"web-ips/model"
)

var DeviceTrie = bean.NewTrieMap() // 设备树

//var ServerList []*model.ServerResp   // 目前第一个是空，第一个是微软云，第二个是东莞，方便和服务器在数据库中的编号匹配，
//var ServerBitMapList []*utils.BitMap // 目前第一个是空，第一个是微软云，第二个是东莞，方便和服务器在数据库中的编号匹配，
//								// 之所以不从0才是，是因为数据传递过程可能会和零值冲突

func InitTrie() {
	var customers []*model.Customer

	initAllCustomers(&customers)
	for _, c := range customers {
		// 3   3
		if c.AddrCode > len(ServerNodeList) {
			logger.Errorf("db server code is err")
			continue
		}
		logger.Debugf("will init customer # %+v", c)
		ServerNodeList[c.AddrCode-1].BitMapNode.Add(uint(c.Id))
		// 加载账户名称和id的映射
		bean.AccountMap.Set(c.AccountName, uint(c.Id))
		go initAllDevices(c.Id)
	}

	ServerNodeListJSON, _ := json.Marshal(ServerNodeList)
	_ = ioutil.WriteFile("server-info.json", ServerNodeListJSON, 0644)
}

// 1. 首先根据数据库里读取所有的调度员,获取调度员应该连接哪个服务
func initAllCustomers(customers *[]*model.Customer) {
	var err error
	*customers, err = customer.GetAllCustomer()
	if err != nil {
		logger.Debugf("initAllCustomers GetAllCustomer error: %+v", err)
	}
}

// 2. 调度员名下的设备也得更新前缀树
func initAllDevices(accountId int) {
	iMeis, err := device.GetAllDevice(accountId)
	if err != nil {
		return
	}
	for _, iMei := range iMeis {
		DeviceTrie.Insert(iMei, accountId)
	}
}
