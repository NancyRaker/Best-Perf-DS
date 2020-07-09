/*
@Time : 2019/9/20 16:50 
@Author : yanKoo
@File : consumer
@Software: GoLand
@Description:
*/
package main

import (
	"fmt"
	"github.com/Shopify/sarama"
	"log"
	"sync"
	"time"
)

var (
	wg sync.WaitGroup
)

func main()  {
	consumeroffset()

}

func consumeroffset() {
	fmt.Println("start consume")
	config := sarama.NewConfig()

	//提交offset的间隔时间，每秒提交一次给kafka
	config.Consumer.Offsets.CommitInterval = 1 * time.Second

	//设置使用的kafka版本,如果低于V0_10_0_0版本,消息中的timestrap没有作用.需要消费和生产同时配置
	config.Version = sarama.V0_10_0_1

	//consumer新建的时候会新建一个client，这个client归属于这个consumer，并且这个client不能用作其他的consumer
	consumer, err := sarama.NewConsumer([]string{"23.101.8.213:9092"}, config)
	if err != nil {
		panic(err)
	}

	//新建一个client，为了后面offsetManager做准备
	client, err := sarama.NewClient([]string{"23.101.8.213:9092"}, config)
	if err != nil {
		panic("client create error")
	}

	//新建offsetManager，为了能够手动控制offset
	offsetManager, err := sarama.NewOffsetManagerFromClient("group_janus_local6", client)
	if err != nil {
		panic("offsetManager create error")
	}

	//创建一个第0分区的offsetManager，每个partition都维护了自己的offset
	partitionOffsetManager, err := offsetManager.ManagePartition("janus_response", 0)
	if err != nil {
		panic("partitionOffsetManager create error")
	}

	fmt.Println("consumer init success")

	//sarama提供了一些额外的方法，以便我们获取broker那边的情况
	topics, _ := consumer.Topics()
	fmt.Println(topics)
	partitions, _ := consumer.Partitions("janus_response")
	fmt.Println(partitions)

	//第一次的offset从kafka获取(发送OffsetFetchRequest)，之后从本地获取，由MarkOffset()得来
	nextOffset, _ := partitionOffsetManager.NextOffset()
	fmt.Println(nextOffset)

	//创建一个分区consumer，从上次提交的offset开始进行消费
	partitionConsumer, err := consumer.ConsumePartition("janus_response", 0, 1000)
	if err != nil {
		panic(err)
	}

	for {
		select {
		case msg := <-partitionConsumer.Messages():
			if msg.Offset < 1593 {
				log.Printf("Consumed message offset %d\n message:%, msg BlockTimestamp:%+v, timestamp: %+v", msg.Offset, string(msg.Value), msg.BlockTimestamp, msg.Timestamp)
			}
			//拿到下一个offset
			nextOffset, _ := partitionOffsetManager.NextOffset()
			//if msg.Offset > 1593 {
				//fmt.Println(nextOffset+1, "...", offsetString)
				//break
			//}
			//提交offset，默认提交到本地缓存，每秒钟往broker提交一次（可以设置）
			partitionOffsetManager.MarkOffset(nextOffset+1, "modified metadata")
		}
	}

}
