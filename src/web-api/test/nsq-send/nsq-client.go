/*
@Time : 2019/7/26 18:28 
@Author : yanKoo
@File : nsq-client
@Software: GoLand
@Description:
*/

//Nsq发送测试
package main

import (
	"bufio"
	"fmt"
	"github.com/nsqio/go-nsq"
	"os"
)

var producer *nsq.Producer

// 主函数
func main() {
	strIP1 := "47.100.116.26:4150"
	strIP2 := "47.100.116.26:4152"
	InitProducer(strIP1)

	running := true

	//读取控制台输入
	reader := bufio.NewReader(os.Stdin)
	for running {
		data, _, _ := reader.ReadLine()
		command := string(data)
		if command == "stop" {
			running = false
		}

		for err := Publish("test", command); err != nil; err = Publish("test", command) {
			//切换IP重连
			strIP1, strIP2 = strIP2, strIP1
			InitProducer(strIP1)
		}
	}
}

// 初始化生产者
func InitProducer(str string) {
	var err error
	fmt.Println("address: ", str)
	producer, err = nsq.NewProducer(str, nsq.NewConfig())
	if err != nil {
		panic(err)
	}
}

//发布消息
func Publish(topic string, message string) error {
	var err error
	if producer != nil {
		if message == "" { //不能发布空串，否则会导致error
			return nil
		}
		err = producer.Publish(topic, []byte(message)) // 发布消息
		return err
	}
	return fmt.Errorf("producer is nil", err)
}
