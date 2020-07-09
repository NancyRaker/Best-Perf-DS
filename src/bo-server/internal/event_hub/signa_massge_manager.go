/*
@Time : 2019/9/26 18:28 
@Author : yanKoo
@File : signa_massge_manager
@Software: GoLand
@Description: 监听consumer发过来的消息，进行分类处理
*/
package event_hub

import (
	"bo-server/internal/mq/mq_receiver"
	"bo-server/logger"
	"bo-server/model"
	"time"
)

type signalMsgManager struct {
	signalMags               []*model.SignalMsg
	createGroupEventHub      *createGroupEvent
	alterParticipantEventHub *alterParticipantEvent
}

type QueryParams struct {
	GroupId             int
	Timeout             time.Duration
	AlterParticipantReq *model.AlterParticipantReq
}

const (
	// EventType

	// response signal
	CREATE_GROUP_RESP      = 1 // 创建房间消息的回复
	DELETE_GROUP_RESP      = 2 // 删除房间消息的回复
	ALTER_PARTICIPANT_RESP = 3 // 改变房间内群成员角色的回复
)

var signalManager *signalMsgManager

func init() {
	signalManager = &signalMsgManager{
		createGroupEventHub:      NewCreateGroupEventHub(),      // 注册群组消息管理器
		alterParticipantEventHub: NewAlterParticipantEventHub(), // 注册改变房间角色消息管理器
	}
}

// 对外暴露启动信令管理器
func SignalManagerInit() *signalMsgManager {
	// 挂起信令类消息处理模块
	go signalManager.manageSignal()
	return signalManager
}

// 暴露对外查询消息回复的接口
func QueryEvent(eventType int, params QueryParams) bool {
	time.Sleep(params.Timeout)
	switch eventType {
	case CREATE_GROUP_RESP:
		return queryCreateGroupEvent(params)
	case DELETE_GROUP_RESP:
	// ... TODO
	case ALTER_PARTICIPANT_RESP:
		return queryAlterParticipantEvent(params)
	}
	return false
}

// 查询创建房间
func queryCreateGroupEvent(params QueryParams) bool {
	resp := signalManager.createGroupEventHub.Get(params.GroupId)
	logger.Debugf("queryCreateGroupEvent with params %+v, res:%+v", params, resp)
	if resp == nil || resp.ResCode != 0 { // 错误码
		return false
	}
	return true
}

// 查询改变房间角色操作结果
func queryAlterParticipantEvent(p QueryParams) bool {
	resp := signalManager.alterParticipantEventHub.Get(p.AlterParticipantReq)
	logger.Debugf("queryAlterParticipantEvent with params %+v, res:%+v", p, resp)
	if resp == nil || resp.Code != 0 { // 错误码
		return false
	}
	return true
}

// 从mq的receiver那里获取操作回复信号来统一管理
func (smm *signalMsgManager) manageSignal() {
	var signalDealerChan = smm.createSignalDealer()
	for {
		var activeSignalDealerChan chan *model.SignalMsg
		var msg *model.SignalMsg
		if len(smm.signalMags) > 0 {
			msg = smm.signalMags[0]
			if msg.CreateGroupResp != nil && msg.Res != nil {
				msg.CreateGroupResp.ResCode = msg.Res.Code
			}
			activeSignalDealerChan = signalDealerChan
		}
		select {
		case responseMsg := <-mq_receiver.SignalMessage():
			logger.Debugf("signalMsgManager responseMsg:%+v", responseMsg)
			if responseMsg == nil {
				continue
			}
			smm.signalMags = append(smm.signalMags, responseMsg)
		// 根据返回类型进行分类处理
		case activeSignalDealerChan <- msg:
			smm.signalMags = smm.signalMags[1:]
		}
	}
}

// 处理janus回复的消息
func (smm *signalMsgManager) createSignalDealer() chan *model.SignalMsg {
	var signalChan = make(chan *model.SignalMsg)
	go smm.dealSignal(signalChan)
	return signalChan
}


// 处理janus信令类的消息
func (smm *signalMsgManager) dealSignal(signalChannel chan *model.SignalMsg) {
	for {
		response := <-signalChannel
		logger.Debugf("signalMsgManager dealSignal deal %+v, res:%+v", response, response.Res)
		switch response.SignalType {
		case CREATE_GROUP_RESP:
			smm.createGroupEventHub.Insert(response.CreateGroupResp, response.Res) // 添加到信号存储
		case DELETE_GROUP_RESP:
			// TODO
		case ALTER_PARTICIPANT_RESP: // 改变房间内的角色的回复
			logger.Debugf("signalMsgManager dealSignal deal %+v, res:%+v, params:%+v", response, response.Res, response.AlterParticipantResp)
			// 在事件hub里面增加一个改变角色返回结果的管理
			smm.alterParticipantEventHub.Insert(response.AlterParticipantResp, response.Res)
		}
	}
}
