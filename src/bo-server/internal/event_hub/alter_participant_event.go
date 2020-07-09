/*
@Time : 2019/9/27 9:10
@Author : yanKoo
@File : eventhub
@Software: GoLand
@Description:
*/
package event_hub

import (
	"bo-server/logger"
	"bo-server/model"
	"strconv"
	"sync"
)

// 改变房间角色的返回
type alterParticipantEvent struct {
	m       sync.Mutex
	respMap map[string]*model.Res
}

var ape *alterParticipantEvent

func init() {
	ape = &alterParticipantEvent{
		respMap: make(map[string]*model.Res),
	}
}

func NewAlterParticipantEventHub() *alterParticipantEvent {
	return ape
}

func (ape *alterParticipantEvent) Insert(e *model.AlterParticipantResp, res *model.Res) {
	if e == nil {
		return
	}
	ape.m.Lock()
	defer ape.m.Unlock()
	ape.respMap[ape.makeKey(e.DispatcherId, e.Gid, e.Uid, e.Role)] = res
	logger.Debugf("alterParticipantEvent Insert e %+v res:%+v", e, res)
}

func (ape *alterParticipantEvent) makeKey(aid, gid, uid, role int) string {
	return strconv.Itoa(aid) + "_" +
		strconv.Itoa(gid) + "_" +
		strconv.Itoa(uid) + "_" +
		strconv.Itoa(role)
}

func (ape *alterParticipantEvent) Get(e *model.AlterParticipantReq) *model.Res {
	if e == nil {
		return nil
	}
	ape.m.Lock()
	defer ape.m.Unlock()
	target, ok := ape.respMap[ape.makeKey(e.DispatcherId, e.Gid, e.Uid, e.Role)]
	if !ok {
		return nil
	}
	return target
}
