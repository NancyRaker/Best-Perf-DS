/*
@Time : 2019/4/12 16:24
@Author : yanKoo
@File : location
@Software: GoLand
@Description: 存储设备发过来的的数据
*/
package location

import (
	"errors"
	"github.com/gomodule/redigo/redis"
	"strconv"
	"web-api/dao/pub"
	"web-api/logger"
	"web-api/model"
	pb "web-api/proto/talk_cloud"
)

const (
	LOCAL_TIME = "local_time"
	LONGITUDE  = "lon"
	LATITUDE   = "lat"
	SPEED_GPS  = "speed"
	COURSE     = "course"
	BATTERY    = "battery"
	WIFI_INFO  = "wifi_info"
)

// 给web端位置数据
func GetUserLocationInCache(uIds *[]*model.Device, rd redis.Conn) (err error) {
	if rd == nil {
		return errors.New("redis conn is nil")
	}
	defer rd.Close()

	for _, uId := range *uIds {
		_ = rd.Send("HMGET", pub.MakeUserDataKey(int32(uId.Id)),
			LONGITUDE, LATITUDE, SPEED_GPS, COURSE, LOCAL_TIME, BATTERY, WIFI_INFO)
	}
	_ = rd.Flush()
	for _, uId := range *uIds {
		source, err := redis.Strings(rd.Receive())
		parseGpsData(uId, source)
		logger.Debugf("from redis # %d gps info :%+v with err: %+v", uId, source, err)
	}
	return
}

func parseGpsData(device *model.Device, resStr []string) {
	var (
		valueNum                int
		wifiInfo                string
		err                     error
		battery                 int
		lon, lat, speed, course float64
		localT                  uint64
	)
	if resStr != nil {
		valueNum = len(resStr)
	}
	if valueNum > 0 && resStr[0] != "" { // 只要任意一个字段为空就是没有这个数据
		lon, err = strconv.ParseFloat(resStr[0], 128)
		if err != nil {
			logger.Debugf("convent lon error:%v", err)
		}
		valueNum--

		if valueNum > 0 && resStr[1] != "" {
			lat, err = strconv.ParseFloat(resStr[1], 128)
			if err != nil {
				logger.Debugf("convent lat error:%v", err)
			}
			valueNum--
		}

		if valueNum > 0 && resStr[2] != ""{
			speed, err = strconv.ParseFloat(resStr[2], 64)
			if err != nil {
				logger.Debugf("convent speed error:%v", err)
			}
			valueNum--
		}

		if valueNum > 0 && resStr[3] != ""{
			course, err = strconv.ParseFloat(resStr[3], 64)
			if err != nil {
				logger.Debugf("convent course error:%v", err)
			}
			valueNum--
		}

		if valueNum > 0 && resStr[4] != ""{
			localT, err = strconv.ParseUint(resStr[4], 10, 64)
			if err != nil {
				logger.Debugf("convent lTime error:%v", err)
			}
			valueNum--
		}

		if valueNum > 0 && resStr[5] != ""{
			battery, err = strconv.Atoi(resStr[5])
			if err != nil {
				logger.Debugf("convent %d battery error:%v", battery, err)
			}
			valueNum--
		}

		if valueNum > 0 && resStr[6] != "" {
			wifiInfo = resStr[6]
			if err != nil {
				logger.Debugf("convent course error:%v", err)
			}
			// TODO 返回给web展示
			//log.Log.Debug(wifiInfo)
		}

		gpsData := &pb.GPS{
			LocalTime: localT,
			Longitude: lon,
			Latitude:  lat,
			Speed:     float32(speed),
			Course:    float32(course),
		}

		device.GPSData = &model.GPS{
			Lng: lon,
			Lat: lat,
		}
		device.Course = gpsData.Course
		device.Speed = gpsData.Speed
		device.LocalTime = gpsData.LocalTime
		device.WifiDes = wifiInfo
	} else {
		// 去数据库查找，返回空
	}
}
