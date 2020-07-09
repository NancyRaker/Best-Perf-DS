/*
@Time : 2019/9/2 15:23 
@Author : yanKoo
@File : account_map
@Software: GoLand
@Description:
*/
package bean

import "sync"

type AccountStru struct {
	accounts map[string]uint // 调度员的账号和id进行映射
	m        sync.Mutex
}

var AccountMap AccountStru

func init() {
	AccountMap = AccountStru{
		accounts: make(map[string]uint),
	}
}

func (a *AccountStru) Set(accountName string, id uint) {
	a.m.Lock()
	defer a.m.Unlock()
	a.accounts[accountName] = id
}

func (a *AccountStru) Get(accountName string) int {
	if id, ok := a.accounts[accountName]; ok {
		return int(id)
	}
	return -1
}
