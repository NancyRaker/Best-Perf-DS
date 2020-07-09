/*
@Time : 2019/9/24 16:03 
@Author : yanKoo
@File : producerClient
@Software: GoLand
@Description:
*/
package kafka

import (
	"bo-server/conf"
	"bo-server/logger"
	"github.com/Shopify/sarama"
)

type producerClient struct {
	producer    sarama.AsyncProducer
	senderQueue chan interface{}
}

func newProducerClient(senderQueue chan interface{}) *producerClient {
	pc := &producerClient{
		senderQueue: senderQueue,
	}
	// 初始化kafka的producer
	config := sarama.NewConfig()
	//等待服务器所有副本都保存成功后的响应
	config.Producer.RequiredAcks = sarama.WaitForAll
	//随机向partition发送消息
	config.Producer.Partitioner = sarama.NewRandomPartitioner
	//是否等待成功和失败后的响应,只有上面的RequireAcks设置不是NoReponse这里才有用.
	config.Producer.Return.Successes = true
	config.Producer.Return.Errors = true
	//设置使用的kafka版本,V2_2_0_0,消息中的timestrap没有作用.需要消费和生产同时配置
	//注意，版本设置不对的话，kafka会返回很奇怪的错误，并且无法成功发送消息
	config.Version = sarama.V2_2_0_0


	logger.Debugln("newProducerClient start make producer")
	//使用配置,新建一个异步生产者
	var e error
	pc.producer, e = sarama.NewAsyncProducer(conf.ProducerAddrs, config)
	if e != nil { //创建失败就不让服务运行起来
		logger.Fatalf("newProducerClient sarama.NewAsyncProducer fail: %+v", e)
		return nil
	}
	//defer pc.producer.AsyncClose()
	return pc
}

func (pc *producerClient) Send() {
	var msgBufferQueue []interface{}
	var msgSender = pc.createMsgSender()
	//tick := time.NewTicker(time.Second * time.Duration(5))
	for {
		var activeExecu chan interface{}
		var activeTask interface{}
		if len(msgBufferQueue) > 0 {
			activeExecu = msgSender
			activeTask = msgBufferQueue[0]
		}
		select {
		case msg := <-pc.senderQueue:
			msgBufferQueue = append(msgBufferQueue, msg)
		case activeExecu <- activeTask:
			msgBufferQueue = msgBufferQueue[1:]
			// 回执
		case suc := <-pc.producer.Successes():
			logger.Debugln("offset: ", suc.Offset, "timestamp: ", suc.Timestamp.String(), "partitions: ", suc.Partition)
		case fail := <-pc.producer.Errors():
			logger.Debugln("err: ", fail)
		}
	}
}

func (pc *producerClient) createMsgSender() chan interface{} {
	msgChan := make(chan interface{})
	go pc.sendMsg(msgChan)
	return msgChan
}

func (pc *producerClient) sendMsg(msgChan chan interface{}) {
	for {
		value := <-msgChan
		// 这里的msg必须得是新构建的变量，不然你会发现发送过去的消息内容都是一样的，因为批次发送消息的关系。
		msg := &sarama.ProducerMessage{
			Topic: conf.ProducerTopic,
		}

		// 如果传递的value是一个结构体就要反序列化。

		//将字符串转化为字节数组。
		msg.Value = sarama.ByteEncoder(value.([]uint8))
		logger.Debugf("producerClient sendMsg:%+v", string(value.([]uint8)))

		pc.producer.Input() <- msg //使用通道发送
	}
}
