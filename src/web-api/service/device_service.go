/*
@Time : 2019/8/22 17:37 
@Author : yanKoo
@File : device
@Software: GoLand
@Description:
*/
package service

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/gomodule/redigo/redis"
	tlc "web-api/dao/location"
	"web-api/dao/pub"
	tu "web-api/dao/user"
	"web-api/engine/cache"
	"web-api/logger"
	"web-api/model"
)


func getAllDeviceBelongAccount(input interface{}, errOut chan interface{}) *[]*model.Device {
	if input == nil {
		errOut <- errors.New("getDevicesAccount input is nil")
		return nil
	}
	accountId := input.(int)
	// 3. 获取所有用户设备

	deviceAll, err := tu.SelectUserByAccountId(accountId)
	if err != nil {
		logger.Debugf("Error in GetGroups: %s", err)
		errOut <- gin.H{
			"error":      "get devices DB error",
			"error_code": "009",
		}
		return nil
	}

	// 5. 如果有位置信息，就带上位置信息
	err = tlc.GetUserLocationInCache(&deviceAll, cache.GetRedisClient())
	if err != nil {
		logger.Debugf("GetAccountInfo GetUserLocationInCache error : %s", err)
		errOut <- model.ErrorInternalServerError
		return nil
	}
	return &deviceAll
}

// 获取群组在线信息 // 目前只用来获取stat
func getDevicesStatus(devices *[]*model.Device,  statMap *map[int32]bool ) error {
	if devices == nil {
		return errors.New("getDevicesAccount input is nil")
	}
	var (
		rd = cache.GetRedisClient()
	)
	if rd == nil {
		err := errors.New("GetAccountInfo getDevicesStatus rd is nil")
		return err
	}
	defer rd.Close()

	var totalUId []int
	for _, device := range *devices{
		totalUId = append(totalUId, device.Id)
	}

	// uid在线状态
	for _, uId := range totalUId {
		_ = rd.Send("GET", pub.GetRedisKey(int32(uId)))
	}
	_ = rd.Flush()
	for _, id := range totalUId {
		source, err := redis.Bytes(rd.Receive())
		s := &model.SessionInfo{}
		_ = json.Unmarshal(source, s)
		// TODO 在线状态
		if err != nil {
		} else {
			(*statMap)[int32(id)] = true
		}
	}

	return nil
}

// 根据设备在线状态排序
func sortDeviceByOnlineStatus(devices *[]*model.Device, statusMap map[int32]bool) {
	var i, j = 0, len(*devices)-1

	for i <= j {
		if statusMap[int32((*devices)[i].Id)] {
			(*devices)[i].Online = pub.USER_ONLINE
			i++
		} else {
			(*devices)[i].Online = pub.USER_OFFLINE
			(*devices)[i], (*devices)[j] = (*devices)[j], (*devices)[i]
			j--
		}
	}
}

// 获取调度员名下所有的设备
func GetDevicesForDispatchers(ctx context.Context, cancelFunc context.CancelFunc, aId int, errOut chan interface{}) (*[]*model.Device, map[int32]bool, error) {
	var (
		statusMap = make(map[int32]bool)
	)
	// 1. 获取所有用户设备
	devices := getAllDeviceBelongAccount(aId, errOut)

	// 2. 获取在线状态
	if err := getDevicesStatus(devices, &statusMap); err != nil {
		return nil, nil, errors.New("GetAccountInfo fail with :" + err.Error())
	}

	// 3.根据设备在线状态排序
	sortDeviceByOnlineStatus(devices, statusMap)
	// 如果调度员在线，也把他删除，让web处理调度员在线不在线的人数的那个字段
	if statusMap[int32(aId)] {
		delete(statusMap, int32(aId))
	}
	return devices, statusMap, nil

}
