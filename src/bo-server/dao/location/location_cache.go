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
	"fmt"
	"github.com/gomodule/redigo/redis"
	pb "bo-server/api/proto"
	"bo-server/dao/pub"
	"bo-server/logger"
	"net/http"
	"strconv"
)

type CacheLocation struct {
	LocalTime  uint64  `json:"local_time"`
	Longitude  float64 `json:"longitude"`
	Latitude   float64 `json:"latitude"`
	Speed      float32 `json:"speed"`
	Course     float32 `json:"course"`
	Battery    int32   `json:"battery"`
	WifiDes    string  `json:"wifi_des"`
	Id         int32   `json:"id"`
	UpdateInfo []interface{}
}

func UpdateUserLocationInCache(cacheLocation *CacheLocation, redisCli redis.Conn, _ bool) error {
	// 向缓存添加用户信息数据
	if redisCli == nil {
		return errors.New("redis conn is nil")
	}
	defer redisCli.Close()

	if cacheLocation == nil || len(cacheLocation.UpdateInfo) == 0 || (len(cacheLocation.UpdateInfo)%2) != 0 {
		return errors.New("cacheLocation is nil")
	}
	if len(cacheLocation.UpdateInfo) == 0 || (len(cacheLocation.UpdateInfo)%2) != 0 {
		return errors.New("cacheLocation is invalid")
	}

	temp := []interface{}{pub.MakeUserDataKey(cacheLocation.Id)}
	temp = append(temp, cacheLocation.UpdateInfo...)

	fmt.Printf("%+v", temp)
	if _, err := redisCli.Do("HMSET", temp...); err != nil {
		return errors.New("HMSET failed with error: " + err.Error())
	}
	return nil
}

// TODO 给web端推送数据 @Deprecated
func GetUserLocationInCache(uId int32, rd redis.Conn) (*pb.GPSHttpResp, *pb.GPS, error) {
	if rd == nil {
		return nil, nil, errors.New("redis conn is nil")
	}
	defer rd.Close()

	value, err := redis.Values(rd.Do("HMGET", pub.MakeUserDataKey(uId),
		pub.LONGITUDE, pub.LATITUDE, pub.SPEED_GPS, pub.COURSE, pub.LOCAL_TIME, pub.BATTERY, pub.WIFI_INFO))
	if err != nil {
		logger.Errorf("hmget failed: %+v", err.Error())
	}
	//log.log.Debugf("Get group %d user info value string  : %s from cache ", gid, value)

	var (
		valueStr    string
		gpsDataResp *pb.GPSHttpResp
		gpsData     *pb.GPS

		valueNum int
		battery  int
		wifiInfo string
	)
	resStr := make([]string, 0)
	for _, v := range value {
		if v != nil {
			valueStr = string(v.([]byte))
			valueNum++
			resStr = append(resStr, valueStr)
		} else {
			break // redis找不到，去数据库加载
		}
	}
	logger.Debugf("Get user %d  gps info : %v from cache", uId, resStr)

	var (
		lon, lat, speed, course float64
		localT                  uint64
	)
	if valueNum > 0 { // 只要任意一个字段为空就是没有这个数据
		lon, err = strconv.ParseFloat(resStr[0], 128)
		if err != nil {
			logger.Debugf("convent lon error:%v", err)
		}
		valueNum--

		if valueNum > 0 {
			lat, err = strconv.ParseFloat(resStr[1], 128)
			if err != nil {
				logger.Debugf("convent lat error:%v", err)
			}
			valueNum--
		}

		if valueNum > 0 {
			speed, err = strconv.ParseFloat(resStr[2], 64)
			if err != nil {
				logger.Debugf("convent speed error:%v", err)
			}
			valueNum--
		}

		if valueNum > 0 {
			course, err = strconv.ParseFloat(resStr[3], 64)
			if err != nil {
				logger.Debugf("convent course error:%v", err)
			}
			valueNum--
		}

		if valueNum > 0 {
			localT, err = strconv.ParseUint(resStr[4], 10, 64)
			if err != nil {
				logger.Debugf("convent lTime error:%v", err)
			}
			valueNum--
		}

		if valueNum > 0 {
			battery, err = strconv.Atoi(resStr[5])
			if err != nil {
				logger.Debugf("convent course error:%v", err)
			}
			valueNum--
		}

		if valueNum > 0 {
			wifiInfo = resStr[6]
			if err != nil {
				logger.Debugf("convent course error:%v", err)
			}
			// TODO 返回给web展示
			//log.log.Debug(wifiInfo)
		}

		gpsData = &pb.GPS{
			LocalTime: localT,
			Longitude: lon,
			Latitude:  lat,
			Speed:     float32(speed),
			Course:    float32(course),
		}
		gpsDataResp = &pb.GPSHttpResp{
			Uid:        uId,
			Res:        &pb.Result{Msg: "", Code: http.StatusOK},
			GpsInfo:    gpsData,
			WifiInfos:  &pb.Wifi{Des: wifiInfo},
			DeviceInfo: &pb.Device{Battery: int32(battery)},
		}
	} else {
		// 去数据库查找，返回空
	}
	return gpsDataResp, gpsData, nil
}

// 给web端位置数据
func GetUsersLocationFromCache(deviceInfos *[]*pb.DeviceInfo, rd redis.Conn) (err error) {
	if rd == nil {
		return errors.New("redis conn is nil")
	}
	defer rd.Close()

	for _, deviceInfo := range *deviceInfos {
		_ = rd.Send("HMGET", pub.MakeUserDataKey(int32(deviceInfo.Id)),
			pub.LONGITUDE, pub.LATITUDE, pub.SPEED_GPS, pub.COURSE, pub.LOCAL_TIME, pub.BATTERY, pub.WIFI_INFO)
	}
	_ = rd.Flush()
	for _, deviceInfo := range *deviceInfos {
		source, err := redis.Strings(rd.Receive())
		//log.log.Debugf("from redis # %d gps info :%+v with err: %+v", deviceInfo.Id, source, err)
		if err != nil {
			logger.Debugf("from redis # %d gps info :%+v with err: %+v", deviceInfo.Id, source, err)
			continue
		}
		parseGpsData(deviceInfo, source)
	}
	return
}

func parseGpsData(device *pb.DeviceInfo, resStr []string) {
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
	//log.log.Debugf("parseGpsData resStr: %+v", resStr)
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
			//log.log.Debug(wifiInfo)
		}

		gpsData := &pb.GPS{
			LocalTime: localT,
			Longitude: lon,
			Latitude:  lat,
			Speed:     float32(speed),
			Course:    float32(course),
		}

		device.Longitude = lon
		device.Latitude = lat
		device.Course = gpsData.Course
		device.Speed = gpsData.Speed
		device.LocalTime = gpsData.LocalTime
		device.WifiDescription = wifiInfo
		if device.Battery == 0 {
			device.Battery = int32(battery)
		}
	} else {
		// 去数据库查找，返回空
	}
}
