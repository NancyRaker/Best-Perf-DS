/*
@Time : 2019/9/24 10:53 
@Author : yanKoo
@File : transport
@Software: GoLand
@Description:
*/
package mq_receiver

import (
	"bo-server/internal/mq/mq_provider/kafka"
	"bo-server/logger"
	"bo-server/model"
	"encoding/json"
	//"fmt"
)

type Receiver interface {
	ListenMsg()
}
type mqReceiver struct {
	consumerRunner     Receiver
	messages           chan interface{}           // 从consumer发送过来的消息
	talkQualityMessage chan *model.TalkQualityMsg // 对讲质量类消息
	pttMessage         chan *model.InterphoneMsg  // ptt对讲消息
	signalMessage      chan *model.SignalMsg      // janus回复的消息
}

const (
	KAFKA_PTTMSG      = 1 // 接收到的kafka消息是对讲消息
	KAFKA_SIGNAL_MSG  = 2 // 接收到的消息是请求处理回复类的消息
	KAFKA_QUALITY_MSG = 3 // 接收到的消息是通话质量类消息

	//SIGNAL_CREATE_RESPOSE = 1 // 创建房间的返回
	//SIGNAL_DELETE_RESPOSE = 2 // 创建房间的返回
)

var receiver mqReceiver

func init() {
	receiver = mqReceiver{
		messages:           make(chan interface{}, 100),
		pttMessage:         make(chan *model.InterphoneMsg, 100),
		talkQualityMessage: make(chan *model.TalkQualityMsg, 100),
		signalMessage:      make(chan *model.SignalMsg, 100),
	}
	receiver.consumerRunner = kafka.NewKafkaConsumer(receiver.messages)

	// 对consumer的消息分类处理
	go receiver.dispatchMsg()

	// 挂起consumer消费消息
	go receiver.consumerRunner.ListenMsg()
}

func (mr *mqReceiver) dispatchMsg() {
	var bufferQueue []interface{}

	for {
		var activeSignalChan chan *model.SignalMsg
		var activePttMsgChan chan *model.InterphoneMsg
		var activeQualityMsgChan chan *model.TalkQualityMsg

		var signalMsg *model.SignalMsg
		var pttMsg *model.InterphoneMsg
		var qualityMsg *model.TalkQualityMsg

		if len(bufferQueue) > 0 {
			// 反序列化，根据消息类型把消息内容往不同的channel发送
			logger.Debugln("will deal:", string((bufferQueue[0]).([]byte)))
			janusNotify := &model.JanusNotifyGrpc{}
			if err := json.Unmarshal((bufferQueue[0]).([]byte), janusNotify); err != nil {
				logger.Debugf("error mq msg:%+v", err)
				continue
			}

			logger.Debugln("will deal:", string((bufferQueue[0]).([]byte)))
			switch janusNotify.MsgType {
			case KAFKA_PTTMSG:
				activePttMsgChan = mr.pttMessage //pttMsgChan
				pttMsg = janusNotify.PttMsg
			case KAFKA_SIGNAL_MSG:
				activeSignalChan = mr.signalMessage //signalChan
				signalMsg = janusNotify.SignalMsg
			case KAFKA_QUALITY_MSG:
				activeQualityMsgChan = mr.talkQualityMessage // talk quality chan
				qualityMsg = janusNotify.QualityMsg
			}
		}

		select {
		case message := <-mr.messages:
			bufferQueue = append(bufferQueue, message)
		case activeSignalChan <- signalMsg:
			bufferQueue = bufferQueue[1:]
		case activePttMsgChan <- pttMsg:
			bufferQueue = bufferQueue[1:]
		case activeQualityMsgChan <- qualityMsg:
			bufferQueue = bufferQueue[1:]
		}
	}
}

func (mr *mqReceiver) createPttDispatcher() chan *model.InterphoneMsg {
	var temp = make(chan *model.InterphoneMsg)
	go func(t chan *model.InterphoneMsg) {
		for {
			msg := <-t
			//fmt.Println("*************///////")
			mr.pttMessage <- msg
		}
	}(temp)
	return temp
}

func (mr *mqReceiver) createSignalDispatcher() chan *model.SignalMsg {
	var temp = make(chan *model.SignalMsg)
	go func(t chan *model.SignalMsg) {
		for {
			msg := <-temp
			mr.signalMessage <- msg
		}
	}(temp)
	return temp
}

// 对外暴露消息的接收
func TalkQualityMessage() <-chan *model.TalkQualityMsg {
	return receiver.talkQualityMessage
}

// 对外暴露消息的接收
func PttMessage() <-chan *model.InterphoneMsg {
	return receiver.pttMessage
}

// 对外暴露消息的接收
func SignalMessage() <-chan *model.SignalMsg {
	return receiver.signalMessage
}
