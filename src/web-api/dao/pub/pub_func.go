/*
@Time : 2019/9/7 17:09 
@Author : yanKoo
@File : pub_func
@Software: GoLand
@Description:
*/
package pub

import (
	"encoding/json"
	"errors"
	"github.com/gomodule/redigo/redis"
	"web-api/engine/cache"
	"web-api/logger"
	"web-api/model"
)

// 获取session,没有错误只有两种返回，有和没有
func GetSession(id int32) (*model.SessionInfo, error) {
	logger.Debugln("getRedisKey(id, key):", GetRedisKey(id))
	rd := cache.GetRedisClient()
	if rd == nil {
		err := errors.New("redis conn is null")
		logger.Debugf("CheckSession with error %+v", err)
		return nil, err
	}
	defer rd.Close()

	if resBytes, err := redis.Bytes(rd.Do("GET", GetRedisKey(id))); err != nil {
		return nil, err
	} else {
		res := &model.SessionInfo{}
		if err := json.Unmarshal(resBytes, res); err != nil {
			logger.Debugln("json err")
			return nil, err
		}
		return res, nil
	}
}

