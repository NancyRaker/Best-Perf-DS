/*
@Time : 2019/10/30 17:29 
@Author : yanKoo
@File : ex——key
@Software: GoLand
@Description:
*/
package main

import (
	"bo-server/api/server/im"
	"bo-server/engine/cache"
	"github.com/gomodule/redigo/redis"
	"log"
	"time"
	"unsafe"
)

type PSubscribeCallback func(pattern, channel, message string)

type PSubscriber struct {
	client redis.PubSubConn
	cbMap  map[string]PSubscribeCallback
}

func (c *PSubscriber) PConnect() {
	rd := cache.GetRedisClient()
	if rd == nil {
		return
	}
	//defer rd.Close()

	c.client = redis.PubSubConn{Conn: rd}
	c.cbMap = make(map[string]PSubscribeCallback)

	go func() {
		for {
			//log.Println("wait...")
			switch res := c.client.Receive().(type) {
			case redis.Message:
				pattern := (*string)(unsafe.Pointer(&res.Pattern))
				channel := (*string)(unsafe.Pointer(&res.Channel))
				message := (*string)(unsafe.Pointer(&res.Data))
				c.cbMap[*channel](*pattern, *channel, *message)
			case redis.Subscription:
				log.Printf("%s: %s %d\n", res.Channel, res.Kind, res.Count)
			case error:
				log.Printf("error handle: %+v", res)
				continue
			}
		}
	}()

}
func (c *PSubscriber) Psubscribe(channel interface{}, cb PSubscribeCallback) {
	err := c.client.PSubscribe(channel)
	if err != nil {
		log.Println("redis Subscribe error: ", err)
	}

	c.cbMap[channel.(string)] = cb
}

func TestPubCallback(patter, chann, msg string) {
	log.Println("TestPubCallback patter : "+patter+" channel : ", chann, " message : ", msg)
}

func main2() {
	log.Println("===========main start============")
	var psub PSubscriber
	psub.PConnect()
	psub.Psubscribe("__keyevent@0__:expired", TestPubCallback)
	for {
		time.Sleep(1 * time.Second)
	}
}

func main()  {
	//im.syncStreamWithRedis()
}