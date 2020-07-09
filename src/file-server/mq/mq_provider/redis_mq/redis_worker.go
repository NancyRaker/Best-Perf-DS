/*
@Time : 2019/9/24 10:25 
@Author : yanKoo
@File : recv_message
@Software: GoLand
@Description:
*/
package redis_mq

import (
	cfgGs "file-server/conf"
	"file-server/engine/cache"
	"file-server/logger"
	"github.com/gomodule/redigo/redis"
	"time"
)

type redisMQWorker struct {
	messages chan interface{}
}

const CHANNEL = "WebNotifyJanus"

func (*redisMQWorker) SendMsg(msg interface{}) {
	c := cache.GetRedisClient()
	if c == nil {
		logger.Debugln("redis is nil")
		return
	}
	defer c.Close()

	var value string
	switch msg.(type) {
	case string:
		value = msg.(string)
	default:
		// TODO 序列化
	}
	_, _ = c.Do("PUBLISH", CHANNEL, value)
}

func (rw *redisMQWorker) ListenMsg() {
	tick := time.NewTicker(time.Millisecond * time.Duration(cfgGs.PttWaitTime))
	for {
		select {
		case <-tick.C:
			func() {
				redisCli := cache.GetRedisClient()
				if redisCli == nil {
					return
				}
				defer redisCli.Close()

				value, err := redis.String(redisCli.Do("lpop", cfgGs.PttMsgKey))
				err = redisCli.Err()
				if err != nil {
					logger.Debugf("ptt lpop with error: %s", err.Error())
				}
				if value != "" {
					rw.messages <- value
					logger.Debugf("Get ptt msg from redis: %s", value)
				}
			}()
		}
	}
}

func NewRedisMQWorker(messages chan interface{}) *redisMQWorker {
	return &redisMQWorker{messages: messages}
}
/*
{"uid": "62", "m_type": "ptt", "md5": "md5", "grp_id": "193", "timestamp": "1568968731", "file_path": "https://dev.yunptt.com:82/group1/M00/00/00/wKhkBl2Jd8SAVwq-AABVgIc4lZw066.mp3"}
*/