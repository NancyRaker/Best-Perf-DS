/*
@Time : 2019/9/19 14:24 
@Author : yanKoo
@File : mq_message
@Software: GoLand
@Description:
*/
package main

import (
	"encoding/json"
	"state-server/model"
	"state-server/pub/cache"
	"fmt"
	"math/rand"
	"sync"
)

/*
https://dev.yunptt.com:83/yankooo/serverv1.0.1/raw/master/login_io.png
https://dev.yunptt.com:82/group1/M00/00/E7/CgAABF1BO_CAaq9zAABVgFYIa8M126.mp3
"{\"uid\": \"62\", \"m_type\": \"ptt\", \"md5\": \"md5\", \"grp_id\": \"395\", \"timestamp\": \"1564556272\", \"file_path\": \"https://dev.yunptt.com:83/yankooo/serverv1.0.1/raw/master/login_io.png\"}"
*/

func main(){
	addMsgToRedisMQ()
}

func addMsgToRedisMQ() {
	var msgs []model.FileSourceInfo

	msgs = append(msgs, model.FileSourceInfo{
		Uid:"62",
		MsgType:"ptt",
		Md5:"md5",
		GId:"395",
		FilePath:"https://dev.yunptt.com:83/yankooo/serverv1.0.1/raw/master/login_io.png",
		Timestamp:"1564556272",
	})

	msgs = append(msgs, model.FileSourceInfo{
		Uid:"62",
		MsgType:"ptt",
		Md5:"md5",
		GId:"3999",
		FilePath:"https://dev.yunptt.com:82/group1/M00/00/BB/wKhkBl2ERxeAY12MAABVgIc4lZw940.mp3",
		Timestamp:"1564556272",
	})


	var wg sync.WaitGroup
	for i:=0; i < 1; i++ {
		wg.Add(1)
		go func(wgg *sync.WaitGroup) {
			defer wgg.Done()
			rd := cache.GetRedisClient()
			defer rd.Close()
			msgByte ,_ := json.Marshal(msgs[rand.Intn(2)])
			_, err := rd.Do("rpush", "janusFileStorage", msgByte)
			if err != nil {
				fmt.Println(err)
			}
		}(&wg)
	}
	wg.Wait()
	fmt.Println("send done")
}