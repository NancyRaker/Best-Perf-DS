/*
@Time : 2019/9/24 10:53 
@Author : yanKoo
@File : transport
@Software: GoLand
@Description:
*/
package mq_sender

import (
	"bo-server/internal/mq/mq_provider/kafka"
	"bo-server/model"
)

type Sender interface {
	SendMsg(*model.MsgObject)
}

type mqSender struct {
	transRunner Sender
	messages    chan interface{}
}

var sender mqSender

const (
	// event type  // 事件类型
	SignalType = 2

	// request signal type 通知janus的信号类消息的消息号定义
	CreateGroupReq      = 1 // 创建janus房间
	DestroyGroupReq     = 2 // 删除房间
	AlterParticipantReq = 3 // 修改设备在janus房间中的角色

)

func init() {
	sender = mqSender{messages: make(chan interface{}, 100)}
	sender.transRunner = kafka.NewKafkaProducer()
}

// 对外暴露发送消息的接口
func SendMsg(msgObj *model.MsgObject) {
	sender.transRunner.SendMsg(msgObj)
}
