/*
@Time : 2019/9/24 16:02 
@Author : yanKoo
@File : worker
@Software: GoLand
@Description:
*/
package kafka

import (
	"state-server/model"
)

type kafkaWorker struct {
	c             *consumerClient
	p             *producerClient
	receiverQueue chan interface{}
	senderQueue   chan interface{}
}

func NewKafkaProducer() *kafkaWorker {
	worker := &kafkaWorker{
		senderQueue:   make(chan interface{}, 100),
	}
	worker.p = newProducerClient(worker.senderQueue)

	go worker.p.Send()
	return worker
}

//NewKafkaConsumer
func NewKafkaConsumer(receiverQueue chan interface{}, config *model.ConsumerConfig) *kafkaWorker {
	worker := &kafkaWorker{
		receiverQueue: receiverQueue,
	}
	worker.c = newConsumerClient(receiverQueue, config)
	return worker
}

func (w *kafkaWorker) SendMsg(msg interface{}) {
	// 把要发送的消息添加到send队列 防止出现阻塞
	go func() { w.senderQueue <- msg }()
}

func (w *kafkaWorker) ListenMsg() {
	w.c.Recv()
}
