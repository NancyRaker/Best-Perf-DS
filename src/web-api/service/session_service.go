/**
* @Author: yanKoo
* @Date: 2019/3/11 16:51
* @Description:
 */
package service

import (
	"encoding/json"
	"errors"
	"github.com/gomodule/redigo/redis"
	"net/http"
	s "web-api/dao/session"
	"web-api/engine/cache"
	"web-api/logger"
	"web-api/model"
	"web-api/utils"
)

var HEADER_FIELD_SESSION = "Authorization"

func GetSessionId(r *http.Request) string {
	return r.Header.Get(HEADER_FIELD_SESSION)
}

// Check if the current user has the permission
// Use session id to do the check
func ValidateAccountSession(r *http.Request, id int) bool {
	sid := GetSessionId(r)
	if len(sid) == 0 {
		return false
	}

	ifExist, err := IsExistsSession(sid, id)
	if err != nil {
		logger.Debugf("validateAccountSession err: %v", err)
		return false
	}
	return ifExist
}

// 判断session 是否存在
func IsExistsSession(sid string, value int) (bool, error) {
	ifExist, err := s.CheckSession(sid, value)
	if err != nil {
		return false, err
	}
	return ifExist, nil
}

// 删除session
func DeleteSessionInfo(session string, aInfo *model.Account) error {
	if err := s.DeleteSession(session, aInfo); err != nil {
		return err
	}
	return nil
}

// 获取session,没有错误只有两种返回，有和没有
func GetSessionValue(value int32) (*model.ImSessionInfo, bool, error) {
	logger.Debugln("getRedisKey(value, key):", utils.GetRedisKey(value))
	rd := cache.GetRedisClient()
	if rd == nil {
		err := errors.New("redis conn is null")
		logger.Debugf("CheckSession with error %+v", err)
		return nil, false, err
	}
	defer rd.Close()

	if resBytes, err := redis.Bytes(rd.Do("GET", utils.GetRedisKey(value))); err != nil {
		logger.Debugf("CheckSession redis.Bytes with error %+v", err)
		return nil, false, nil
	} else {
		res := &model.ImSessionInfo{}
		if err := json.Unmarshal(resBytes, res); err != nil {
			logger.Debugln("json err")
			return nil, false, err
		}
		return res, true, nil
	}
}
