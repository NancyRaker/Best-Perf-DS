/*
@Time : 2019/5/14 11:49
@Author : yanKoo
@File : user_cache_test
@Software: GoLand
@Description:
*/
package user_cache

import (
	"github.com/gomodule/redigo/redis"
	"bo-server/dao/pub"
	"bo-server/engine/cache"
	"testing"
)

func testGetGroupMemDataFromCache(t *testing.T) {
	online, err := redis.Int(cache.GetRedisClient().Do("GET", pub.GetRedisKey(int32(7))))
	if err != nil {
		t.Error("get user online with err ", err.Error())
	}
	t.Log(online)
}

func TestAddUserDataInCache(t *testing.T) {
	err := AddUserDataInCache(17, []interface{}{
		pub.USER_Id, 17,
		pub.IMEI, "355172100001534",
		pub.USER_NAME, "355172100001534",
		pub.NICK_NAME, "vss2",
		pub.USER_TYPE, 1,
		pub.LOCK_GID, 2,
		pub.ONLINE, pub.USER_OFFLINE, // 加载数据默认全部离线
	}, cache.GetRedisClient())
	t.Log(err)
}
