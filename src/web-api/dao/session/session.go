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
	"strconv"
	"web-api/engine/cache"
	"web-api/logger"
	"web-api/model"
)

const (
	ttl = 60 * 60 * 3 // session 过期时间
)

func getRedisKey(key int) string {
	return "web:" + strconv.Itoa(key) + ":acc"
}

// 删除缓存中的session
func DeleteSession(s string, aInfo *model.Account) error {
	rdsCli := cache.GetRedisClient()
	if rdsCli == nil {
		return errors.New("redis connection is nil")
	}
	defer rdsCli.Close()

	sInfo, err := GetSessionValue(s, aInfo.Id)
	if err != nil {
		return errors.New("delete session error: " + err.Error())
	}

	_ = rdsCli.Send("MULTI")
	//_ = rdsCli.Send("DEL", getRedisKey(aInfo.Username, s))
	_ = rdsCli.Send("DEL", getRedisKey(sInfo.AccountId))
	_, err = rdsCli.Do("EXEC")
	if err != nil {
		logger.Debugln("delete session error: ", err)
		return err
	}

	return nil
}

// 更新缓存中的session
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
	//_ = rdsCli.Send("MULTI")
	//_ = rdsCli.Send("SET", getRedisKey(sInfo.UserName, sInfo.SessionID), value, "ex", 60*60*3)
	_, err = rdsCli.Do("SET", getRedisKey(sInfo.AccountId), value, "ex", ttl)
	//_, err = rdsCli.Do("EXEC")
	if err != nil {
		return err
	}
	return nil
}

// 判断是否存在session
func ExistSession(sid string, value int) bool {
	logger.Debugf("sid, value :%s, %+v", sid, value)
	rd := cache.GetRedisClient()
	if rd == nil {
		logger.Debugf("CheckSession with error %+v", errors.New("redis conn is null"))
		return false
	}
	defer rd.Close()

	ifExist, err := redis.Bool(rd.Do("EXISTS", getRedisKey(value)))
	if err != nil {
		logger.Debugf("CheckSession with error %+v", err)
		return false
	}

	return ifExist
}

// 校验session是否合法
func CheckSession(sid string, value int) (bool, error) {
	logger.Debugf("sid, value :%s, %+v", sid, value)
	rd := cache.GetRedisClient()
	if rd == nil {
		logger.Debugf("CheckSession with error %+v", errors.New("redis conn is null"))
	}
	defer rd.Close()

	sInfo, err := GetSessionValue(sid, value)
	if err != nil {
		return false, errors.New("Get SessionValue error: " + err.Error())
	}

	if sInfo.SessionID == sid {
		logger.Debugf("reset session %+v", sInfo)
		go InsertSession(sInfo)
		return true, nil
	}
	return false, nil
}

// 获取session
func GetSessionValue(sId string, value int) (*model.SessionInfo, error) {
	logger.Debugln("getRedisKey(value, key):", getRedisKey(value))
	rd := cache.GetRedisClient()
	if rd == nil {
		logger.Debugf("CheckSession with error %+v", errors.New("redis conn is null"))
	}
	defer rd.Close()

	if resBytes, err := redis.Bytes(rd.Do("GET", getRedisKey(value))); err != nil {
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
