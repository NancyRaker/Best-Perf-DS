/**
* @Author: yanKoo
* @Date: 2019/3/15 14:11
* @Description: 调试阶段redis存储 session
 */
package session

import (
	"encoding/json"
	"errors"
	"github.com/gomodule/redigo/redis"
	cfgGs "bo-server/conf"
	"bo-server/dao/pub"
	"bo-server/engine/cache"
	"bo-server/logger"
	"bo-server/model"
)

// 删除缓存中的session
func DeleteSession(id int32, redisCli redis.Conn) error {
	logger.Debugf("start set user offline state")
	if redisCli == nil {
		return errors.New("redis conn is nil")
	}
	defer redisCli.Close()

	if _, err := redisCli.Do("DEL", pub.GetRedisKey(id)); err != nil {
		return errors.New("UpdateOnlineInCache  failed with error:" + err.Error())
	}
	return nil
}

// 插入缓存中的session
func InsertSession(sInfo *model.SessionInfo) error {
	rdsCli := cache.GetRedisClient()
	if rdsCli == nil {
		return errors.New("redis connection is nil")
	}
	defer rdsCli.Close()

	value, err := json.Marshal(*sInfo)
	if err != nil {
		return err
	}

	_, err = rdsCli.Do("SET", pub.GetRedisKey(sInfo.Id), value, "ex", 60*60*3)
	if err != nil {
		return err
	}
	return nil
}

// 判断是否存在session
func ExistSession(sid string, value int32) bool {
	logger.Debugf("sid, value :%s, %+v", sid, value)
	rd := cache.GetRedisClient()
	if rd == nil {
		logger.Debugf("CheckSession with error %+v", errors.New("redis conn is null"))
		return false
	}
	defer rd.Close()

	ifExist, err := redis.Bool(rd.Do("EXISTS", pub.GetRedisKey(value)))
	if err != nil {
		logger.Debugf("CheckSession with error %+v", err)
		return false
	}

	return ifExist
}

// 校验session是否合法
func CheckSession(sid string, id int32) (bool, error) {
	logger.Debugf("sid, id :%s, %d", sid, id)
	rd := cache.GetRedisClient()
	if rd == nil {
		err := errors.New("redis conn is null")
		logger.Debugf("CheckSession with error %+v", err)
		return false, err
	}
	defer rd.Close()

	sInfo, ok, err := GetSessionValue(id)
	if err != nil {
		return false, errors.New("get session error: " + err.Error())
	}

	if ok && sInfo != nil && sInfo.SessionId == sid {
		return true, nil
	}
	return false, nil
}

// 获取session,没有错误只有两种返回，有和没有
func GetSessionValue(value int32) (*model.SessionInfo, bool, error) {
	logger.Debugln("getRedisKey(value, key):", pub.GetRedisKey(value))
	rd := cache.GetRedisClient()
	if rd == nil {
		err := errors.New("redis conn is null")
		logger.Debugf("CheckSession with error %+v", err)
		return nil, false, err
	}
	defer rd.Close()

	if resBytes, err := redis.Bytes(rd.Do("GET", pub.GetRedisKey(value))); err != nil {
		logger.Debugf("CheckSession redis.Bytes with error %+v", err)
		return nil, false, nil
	} else {
		res := &model.SessionInfo{}
		if err := json.Unmarshal(resBytes, res); err != nil {
			logger.Debugln("json err")
			return nil, false, err
		}
		return res, true, nil
	}
}

// 更新缓存中的session
func UpdateSessionInCache(sInfo *model.SessionInfo) error {
	rdsCli := cache.GetRedisClient()
	if rdsCli == nil {
		return errors.New("redis connection is nil")
	}
	defer rdsCli.Close()

	value, err := json.Marshal(*sInfo)
	if err != nil {
		return err
	}

	_, err = rdsCli.Do("SET", pub.GetRedisKey(sInfo.Id), value, "ex", cfgGs.ExpireTime)
	if err != nil {
		return err
	}
	return nil
}
