/*
@Time : 2019/9/24 16:03 
@Author : yanKoo
@File : consumerClient
@Software: GoLand
@Description:
*/
package kafka

import (
	"github.com/Shopify/sarama"
	"state-server/logger"
	"state-server/model"
	"time"
)

type consumerClient struct {
	client                 sarama.Client
	config                 *sarama.Config
	consumer               sarama.Consumer
	offsetManager          sarama.OffsetManager
	partitionOffsetManager sarama.PartitionOffsetManager
	partitionConsumer      sarama.PartitionConsumer

	receiverQueue chan interface{}
}

const kafkaMetadata = "modified metadata"

func newConsumerClient(receiverQueue chan interface{}, config *model.ConsumerConfig) *consumerClient {
	var (
		cc  = &consumerClient{receiverQueue: receiverQueue}
		err error
	)
	logger.Debugln("start consumer Client init")

	cc.config = sarama.NewConfig()
	cc.config.ClientID = "GRPC_consumer" + config.Group

	//提交offset的间隔时间，每秒提交一次给kafka // TODO 这个时间间隔会存在服务重启的时候重复消费没有提交的数据
	cc.config.Consumer.Offsets.CommitInterval = 1 * time.Second

	//设置使用的kafka版本,如果低于V0_10_0_0版本,消息中的timestrap没有作用.需要消费和生产同时配置
	cc.config.Version = sarama.V2_2_0_0

	//consumer新建的时候会新建一个client，这个client归属于这个consumer，并且这个client不能用作其他的consumer
	cc.consumer, err = sarama.NewConsumer([]string{config.Addr}, cc.config)
	if err != nil { // 初始化失败就直接不让服务运行起来
		logger.Fatalf("newConsumerClient sarama.NewConsumer fail : %+v", err)
		return nil
	}

	//新建一个client，为了后面offsetManager做准备
	cc.client, err = sarama.NewClient([]string{config.Addr}, cc.config)
	if err != nil {
		logger.Fatalf("newConsumerClient sarama.NewClient fail : %+v", err)
		return nil
	}

	//新建offsetManager，为了能够手动控制offset
	cc.offsetManager, err = sarama.NewOffsetManagerFromClient(config.Group, cc.client)
	if err != nil {
		logger.Fatalf("newConsumerClient sarama.NewOffsetManagerFromClient fail : %+v", err)
		return nil
	}

	//创建一个第0分区的offsetManager，每个partition都维护了自己的offset
	cc.partitionOffsetManager, err = cc.offsetManager.ManagePartition(config.Topic, config.Partition)
	if err != nil {
		logger.Fatalf("newConsumerClient cc.offsetManager.ManagePartition fail : %+v", err)
		return nil
	}
	logger.Debugln("consumer Client init success")

	//获取broker那边的情况
	topics, _ := cc.consumer.Topics()
	logger.Debugf("this kafka server has topics: %+v", topics)
	partitions, _ := cc.consumer.Partitions(config.Topic)
	logger.Debugf("this kafka client will consume partitions: %+v", partitions)

	//第一次的offset从kafka获取(发送OffsetFetchRequest)，之后从本地获取，由MarkOffset()得来
	nextOffset, _ := cc.partitionOffsetManager.NextOffset()
	logger.Debugf("this kafka client last offset partitions: %+v", nextOffset)

	//创建一个分区consumer，从上次提交的offset开始进行消费
	cc.partitionConsumer, err = cc.consumer.ConsumePartition(config.Topic, config.Partition, nextOffset+1)
	if err != nil {
		logger.Fatalf("newConsumerClient cc.offsetManager.ManagePartition fail : %+v", err)
		return nil
	}

	return cc
}

func (cc *consumerClient) Recv() {
	for {
		select {
		case msg := <-cc.partitionConsumer.Messages():
			logger.Debugf("Consumed message offset %d message:%s\n", msg.Offset, string(msg.Value))
			cc.receiverQueue <- msg.Value

			//拿到下一个offset
			nextOffset, offsetString := cc.partitionOffsetManager.NextOffset()
			logger.Debugln("consumerClient", nextOffset+1, "...", offsetString)

			//提交offset，默认提交到本地缓存，每秒钟往broker提交一次
			cc.partitionOffsetManager.MarkOffset(nextOffset+1, kafkaMetadata)
		}
	}
}
