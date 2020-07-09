/*
@Time : 2019/9/24 10:53 
@Author : yanKoo
@File : transport
@Software: GoLand
@Description:
*/
package mq_sender

import (
	"state-server/mq/mq_provider/kafka"
)

type Sender interface {
	SendMsg(interface{})
}

type mqSender struct {
	transRunner Sender
	messages    chan interface{}
}

var sender mqSender

const (
	// event type
	SignalType = 2

	// request signal type
	CreateGroupReq  = 1
)


func init() {
	sender = mqSender{messages: make(chan interface{}, 100)}
	sender.transRunner = kafka.NewKafkaProducer()
}

// 对外暴露发送消息的接口
func SendMsg(msg interface{}) {
	sender.transRunner.SendMsg(msg)
}

