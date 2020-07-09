/*
@Time : 2019/8/27 15:16 
@Author : yanKoo
@File : TriemA
@Software: GoLand
@Description: 用一个前缀树来保存设备应该获取的ip
*/
package bean

import "sync"

const R = 10 // 因为设备的imei登录账号只有15位都是数字

type TrieNode struct {
	Children  [R]*TrieNode
	AccountId int // 该设备属于的调度员
	m sync.Mutex
}

func NewTrieMap() *TrieNode {
	return &TrieNode{Children: [R]*TrieNode{}, AccountId: 0} // root节点不存地址
}

// 插入节点
func (root *TrieNode) Insert(iMei string, accountId int) {
	root.m.Lock()
	defer root.m.Unlock()
	cur := root
	for i := range iMei {
		c := rune(iMei[i])
		idx := c - '0'
		if idx < 0 || idx > 9 {
			continue
		}
		if cur.Children[idx] == nil {
			cur.Children[idx] = &TrieNode{Children: [R]*TrieNode{}, AccountId: 0}
		}
		cur = cur.Children[idx]
	}
	cur.AccountId = accountId
}

// 搜索节点
func (root *TrieNode) Search(iMei string) int {
	cur := root
	for i := range iMei {
		c := rune(iMei[i])
		idx := c - '0'
		if idx < 0 || idx > 9 {
			return -1
		}
		if cur.Children[idx] == nil {
			return -1
		}
		cur = cur.Children[idx]
	}
	return cur.AccountId
}

//  更新前缀树节点值 @unThreadSafe
func (root *TrieNode) ChangeAccount(iMei string, accountId int) {
	if iMei == "" || len(iMei) == 0 {
		return
	}
	root.m.Lock()
	defer root.m.Unlock()
	cur := root
	for i := range iMei {
		c := rune(iMei[i])
		idx := c - '0'
		if idx < 0 || idx > 9 {
			return
		}
		if cur.Children[idx] == nil {
			return
		}
		cur = cur.Children[idx]
	}
	cur.AccountId = accountId
}
