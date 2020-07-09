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
	"state-server/conf"
	"state-server/logger"
	"state-server/model"
	"state-server/mq/mq_provider/kafka"
	//"fmt"
)

type Receiver interface {
	ListenMsg()
}
type mqReceiver struct {
	transRunners []Receiver
	messages     chan interface{}          // 从consumer发送过来的消息
	stateMessage chan *model.StateMessage // State对讲消息
}

var receiver mqReceiver

func init() {
	receiver = mqReceiver{
		messages:     make(chan interface{}, 100),
		stateMessage: make(chan *model.StateMessage, 100),
	}
	for _, config := range conf.Config.KafkaConfig.Consumers{
		receiver.transRunners = append(receiver.transRunners, kafka.NewKafkaConsumer(receiver.messages, config))
	}

	// 对consumer的消息分类处理
	go receiver.dispatchMsg()

	// 挂起consumer消费消息
	for _, transRunner := range receiver.transRunners{
		go transRunner.ListenMsg()
	}
}

func (mr *mqReceiver) dispatchMsg() {
	var bufferQueue []interface{}

	for {
		var activeStateMsgChan chan *model.StateMessage
		var StateMsg *model.StateMessage

		if len(bufferQueue) > 0 {
			// 反序列化
			logger.Debugln("mqReceiver: ", string((bufferQueue[0]).([]byte)))
			stateMessage := &model.StateMessage{}
			if err := json.Unmarshal((bufferQueue[0]).([]byte), stateMessage); err != nil {
				logger.Debugf("error mq msg:%+v", err)
				continue
			}

			activeStateMsgChan = mr.stateMessage //stateMsgChan
			StateMsg = stateMessage
		}

		select {
		case message := <-mr.messages:
			bufferQueue = append(bufferQueue, message)
		case activeStateMsgChan <- StateMsg:
			bufferQueue = bufferQueue[1:]
		}
	}
}

// 对外暴露消息的接收
func StateMessage() <-chan *model.StateMessage {
	return receiver.stateMessage
}

