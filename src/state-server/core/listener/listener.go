/*
@Time : 2019/9/18 14:26 
@Author : yanKoo
@File : listener
@Software: GoLand
@Description:从redis/mq获取State消息转发
*/
package listener

import (
	"state-server/logger"
	"state-server/model"
	"state-server/mq/mq_receiver"
)

/**
 * 从消息队列中获取状态转换消息
 */
type stateMsgListener struct {
	messages chan *model.StateMessage
}

func NewStateMsgListener(fileChan chan *model.StateMessage) *stateMsgListener {
	return &stateMsgListener{messages: fileChan}
}

// 从kafka获取State消息
func (stateMsgListener) receiverStateMsgFromMQ(StateD chan *model.StateMessage) {
	for {
		select {
		case msg := <-mq_receiver.StateMessage():
			StateD <- msg
		}
	}
}

// 实现推送的client中的dispatcher方法
func (pfl stateMsgListener) Listen() {
	// 消息队列获取对讲音频信息传递给task_runner
	StateC := make(chan *model.StateMessage, 100)
	go pfl.receiverStateMsgFromMQ(StateC)

	var msgQueue []*model.StateMessage
	var dispatcherChan = pfl.createStateDispatcher()
	for {
		var activeChan chan *model.StateMessage
		var activeMsg *model.StateMessage
		if len(msgQueue) > 0 {
			activeChan = dispatcherChan
			activeMsg = msgQueue[0]

		}
		select {
		case t := <-StateC:
			logger.Debugf("receive State msg: %+v", t)
			msgQueue = append(msgQueue, t)
		case activeChan <- activeMsg:
			msgQueue = msgQueue[1:]
		}
	}
}

// 创建一个State消息的分发器
func (pfl stateMsgListener) createStateDispatcher() chan *model.StateMessage {
	messageC := make(chan *model.StateMessage)
	go pfl.StateMidHandler(messageC)
	return messageC
}

// json反序列化 现在不需要了
func (pfl stateMsgListener) StateMidHandler(c chan *model.StateMessage) {
	for {
		StateMsg := <-c
		logger.Debugf("Will send State msg: %s", StateMsg)
		if StateMsg.MsgType > 0 {
			// 往chan里面插入State消息
			pfl.messages <- StateMsg
		}
	}
}
