/*
@Time : 2019/9/18 14:26 
@Author : yanKoo
@File : listener
@Software: GoLand
@Description:从redis/mq获取ptt消息转发
*/
package listener

import (
	"file-server/logger"
	"file-server/model"
	"file-server/mq/mq_receiver"
)

/**
 * 从消息队列中获取设备的对讲语音消息
 */
type pttFileListener struct {
	fileChan chan *model.InterphoneMsg
}

func NewFileListener(fileChan chan *model.InterphoneMsg) *pttFileListener {
	return &pttFileListener{fileChan: fileChan}
}

// 从redis获取ptt消息
func (pttFileListener) receiverPttMsg(pttD chan string) {
	/*tick := time.NewTicker(time.Millisecond * time.Duration(conf.PttWaitTime))
	for {
		select {
		case <-tick.C:
			func() {
				redisCli := cache.GetRedisClient()
				if redisCli == nil {
					return
				}
				defer redisCli.Close()

				value, err := redis.String(redisCli.Do("lpop", conf.PttMsgKey))
				err = redisCli.Err()
				if err != nil {
					log.Log.Debugf("ptt lpop with error: %s", err.Error())
				}
				if value != "" {
					pttD <- value
					log.Log.Debugf("Get ptt msg from redis: %s", value)
				}
			}()
		}
	}
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
		FilePath:"https://dev.yunptt.com:82/group1/M00/15/D3/wKhkBl2JdBeAIaRRAABVgIc4lZw079.mp3",
		Timestamp:"1564556272",
	})


	time.Sleep(time.Second*30)
	for i:=0; i < 10000; i++ {
		msgByte ,_ := json.Marshal(msgs[rand.Intn(2)])
		pttD <- string(msgByte)
	}

	time.Sleep(time.Minute*10)
	for i:=0; i < 10000; i++ {
		msgByte ,_ := json.Marshal(msgs[0])
		pttD <- string(msgByte)
	}

	time.Sleep(time.Minute*30)
	for i:=0; i < 10000; i++ {
		msgByte ,_ := json.Marshal(msgs[0])
		pttD <- string(msgByte)
	}*/
}

// 从kafka获取ptt消息
func (pttFileListener) receiverPttMsgFromMQ(pttD chan *model.InterphoneMsg) {
	for {
		select {
		case msg:=<-mq_receiver.PttMessage():
			pttD <- msg
		}
	}
}

// 实现推送的client中的dispatcher方法
func (pfl pttFileListener) Listen() {
	// 消息队列获取对讲音频信息传递给task_runner
	pttC := make(chan *model.InterphoneMsg, 100)
	go pfl.receiverPttMsgFromMQ(pttC)

	var msgQueue []*model.InterphoneMsg
	var dispatcherChan = pfl.createPttDispatcher()
	for {
		var activeChan chan *model.InterphoneMsg
		var activeMsg *model.InterphoneMsg
		if len(msgQueue) > 0 {
			activeChan = dispatcherChan
			activeMsg = msgQueue[0]
		}
		select {
		case t := <-pttC:
			logger.Debugf("receive ptt msg: %+v", t)
			msgQueue = append(msgQueue, t)
		case activeChan <- activeMsg:
			msgQueue = msgQueue[1:]
		}
	}
}

// 创建一个ptt消息的分发器
func (pfl pttFileListener) createPttDispatcher() chan *model.InterphoneMsg {
	messageC := make(chan *model.InterphoneMsg)
	go pfl.pttMidHandler(messageC)
	return messageC
}

// json反序列化 现在不需要了
func (pfl pttFileListener) pttMidHandler(c chan *model.InterphoneMsg) {
	for {
		pttMsg := <-c
		logger.Debugf("Will send Ptt msg: %s", pttMsg)
		if pttMsg.Uid != "" && pttMsg.GId != "" {
			// 往chan里面插入ptt消息
			pfl.fileChan <- pttMsg
		}
	}
}
