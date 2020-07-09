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
	"os"
	"os/signal"
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
	consumer, err := sarama.NewConsumer([]string{"10.0.10.28:9092"}, config)
	if err != nil {
		panic(err)
	}

	//新建一个client，为了后面offsetManager做准备
	client, err := sarama.NewClient([]string{"10.0.10.28:9092"}, config)
	if err != nil {
		panic("client create error")
	}
	//defer client.Close()

	//新建offsetManager，为了能够手动控制offset
	offsetManager, err := sarama.NewOffsetManagerFromClient("group_Bitt", client)
	if err != nil {
		panic("offsetManager create error")
	}
	//defer offsetManager.Close()

	//创建一个第0分区的offsetManager，每个partition都维护了自己的offset
	partitionOffsetManager, err := offsetManager.ManagePartition("09241752_test", 0)
	if err != nil {
		panic("partitionOffsetManager create error")
	}
	//defer partitionOffsetManager.Close()

	fmt.Println("consumer init success")

	//defer func() {
	//	if err := consumer.Close(); err != nil {
	//		log.Fatalln(err)
	//	}
	//}()

	//sarama提供了一些额外的方法，以便我们获取broker那边的情况
	topics, _ := consumer.Topics()
	fmt.Println(topics)
	partitions, _ := consumer.Partitions("09241752_test")
	fmt.Println(partitions)

	//第一次的offset从kafka获取(发送OffsetFetchRequest)，之后从本地获取，由MarkOffset()得来
	nextOffset, _ := partitionOffsetManager.NextOffset()
	fmt.Println(nextOffset)

	//创建一个分区consumer，从上次提交的offset开始进行消费
	partitionConsumer, err := consumer.ConsumePartition("09241752_test", 0, nextOffset+1)
	if err != nil {
		panic(err)
	}

	//defer func() {
	//	if err := partitionConsumer.Close(); err != nil {
	//		log.Fatalln(err)
	//	}
	//}()

	// Trap SIGINT to trigger a shutdown.
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt)

	fmt.Println("start consume really")

ConsumerLoop:
	for {
		select {
		case msg := <-partitionConsumer.Messages():
			log.Printf("Consumed message offset %d\n message:%s", msg.Offset, string(msg.Value))
			//拿到下一个offset
			nextOffset, offsetString := partitionOffsetManager.NextOffset()
			fmt.Println(nextOffset+1, "...", offsetString)
			//提交offset，默认提交到本地缓存，每秒钟往broker提交一次（可以设置）
			partitionOffsetManager.MarkOffset(nextOffset+1, "modified metadata")

		case <-signals:
			break ConsumerLoop
		}
	}

}
