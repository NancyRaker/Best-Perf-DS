/*
@Time : 2019/7/3 11:05 
@Author : yanKoo
@File : group_member_test
@Software: GoLand
@Description:
*/
package group_member

import (
	"testing"
	"web-api/engine/cache"
)

//func TestSelectDevicesByGroupId(t *testing.T) {
//	res, err := SelectDevicesByGroupId(49)
//	t.Logf("%d, %+v, %+v", len(res), err, res)
//}

func TestSelectDevicesByGroupId2(t *testing.T) {
	for i := 0; i < 1000; i++ {
		cache.GetRedisClient()
	}
}