/*
@Time : 2019/9/24 10:53 
@Author : yanKoo
@File : transport
@Software: GoLand
@Description:
*/
package mq_receiver

import (
	"encoding/json"
	"file-server/logger"
	"file-server/model"
	"file-server/mq/mq_provider/kafka"
	//"fmt"
)

type Receiver interface {
	ListenMsg()
}
type mqReceiver struct {
	transRunner   Receiver
	messages      chan interface{}          // 从consumer发送过来的消息
	pttMessage    chan *model.InterphoneMsg // ptt对讲消息
	signalMessage chan *model.SignalMsg     // janus回复的消息
}

const (
	KAFKA_PTTMSG     = 1 // 接收到的kafka消息是对讲消息
	KAFKA_SIGNAL_MSG = 2 // 接收到的消息是请求处理回复类的消息

	//SIGNAL_CREATE_RESPOSE = 1 // 创建房间的返回
	//SIGNAL_DELETE_RESPOSE = 2 // 创建房间的返回
)

var receiver mqReceiver

func init() {
	receiver = mqReceiver{
		messages:      make(chan interface{}, 100),
		pttMessage:    make(chan *model.InterphoneMsg, 100),
		signalMessage: make(chan *model.SignalMsg, 100),
	}
	receiver.transRunner = kafka.NewKafkaConsumer(receiver.messages)

	// 对consumer的消息分类处理
	go receiver.dispatchMsg()

	// 挂起consumer消费消息
	go receiver.transRunner.ListenMsg()
}

func (mr *mqReceiver) dispatchMsg() {
	var bufferQueue []interface{}

	//var pttMsgChan = mr.createPttDispatcher()
	//var signalChan = mr.createSignalDispatcher()

	for {
		var activeSignalChan chan *model.SignalMsg
		var activePttMsgChan chan *model.InterphoneMsg

		var signalMsg *model.SignalMsg
		var pttMsg *model.InterphoneMsg

		if len(bufferQueue) > 0 {
			// 反序列化，根据消息类型把消息内容往不同的channel发送
			logger.Debugln("mqReceiver: ", string((bufferQueue[0]).([]byte)))
			janusNotify := &model.JanusNotifyGrpc{}
			if err := json.Unmarshal((bufferQueue[0]).([]byte), janusNotify); err != nil {
				logger.Debugf("error mq msg:%+v", err)
				continue
			}
			switch janusNotify.MsgType {
			case KAFKA_PTTMSG:
				activePttMsgChan = mr.pttMessage //pttMsgChan
				pttMsg = janusNotify.PttMsg
			case KAFKA_SIGNAL_MSG:
				activeSignalChan = mr.signalMessage //signalChan
				signalMsg = janusNotify.SignalMsg
			}
		}

		select {
		case message := <-mr.messages:
			bufferQueue = append(bufferQueue, message)
		case activeSignalChan <- signalMsg:
			bufferQueue = bufferQueue[1:]
		case activePttMsgChan <- pttMsg:
			bufferQueue = bufferQueue[1:]
		}
	}
}

// 对外暴露消息的接收
func PttMessage() <-chan *model.InterphoneMsg {
	return receiver.pttMessage
}

// 对外暴露消息的接收
func SignalMessage() <-chan *model.SignalMsg {
	return receiver.signalMessage
}
