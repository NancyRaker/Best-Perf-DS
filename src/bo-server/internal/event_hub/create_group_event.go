/*
@Time : 2019/10/21 16:14 
@Author : yanKoo
@File : create_group_event
@Software: GoLand
@Description:
*/
package event_hub

import (
"bo-server/logger"
"bo-server/model"
"sync"
)

// 创建房间的返回
type createGroupEvent struct {
	m       sync.Mutex
	respMap map[int]*model.CreateGroupResp
}

var createGroupEventHub *createGroupEvent

func init() {
	createGroupEventHub = &createGroupEvent{
		respMap: make(map[int]*model.CreateGroupResp),
	}
}

func NewCreateGroupEventHub() *createGroupEvent {
	return createGroupEventHub
}

func (*createGroupEvent) Insert(e *model.CreateGroupResp, res *model.Res) {
	if e == nil {
		return
	}
	createGroupEventHub.m.Lock()
	defer createGroupEventHub.m.Unlock()
	createGroupEventHub.respMap[e.GId] = e
	logger.Debugf("createGroupEventHub Insert group %d res:%+v", e.GId, e)
}

func (*createGroupEvent) Get(groupId int) *model.CreateGroupResp {
	createGroupEventHub.m.Lock()
	defer createGroupEventHub.m.Unlock()
	target, ok := createGroupEventHub.respMap[groupId]
	if !ok {
		return nil
	}
	return target
}

