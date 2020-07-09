/*
@Time : 2019/11/28 14:42 
@Author : yanKoo
@File : device
@Software: GoLand
@Description:
*/
package user_cache

import (
	"errors"
	"fmt"
	"state-server/public/cache"
)

const (
	SLEEP_STATE      = "sleeping"
	USR_DATA_KEY_FMT = "usr:%d:state_data"
)

func makeUserDataKey(uid int32) string {
	return fmt.Sprintf(USR_DATA_KEY_FMT, uid)
}

func ModifyUserState(uId int32, UpdateValuePair []interface{}) error {
	redisCli := cache.GetRedisClient()
	if redisCli == nil {
		return errors.New("redis conn is nil")
	}
	defer redisCli.Close()

	value := []interface{}{makeUserDataKey(uId)}

	if _, err := redisCli.Do("HMSET", append(value, UpdateValuePair...)...); err != nil {
		return errors.New("hSet failed with error: " + err.Error())
	}
	return nil
}
