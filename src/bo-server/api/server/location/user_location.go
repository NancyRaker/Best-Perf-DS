/*
@Time : 2019/4/11 8:53
@Author : yanKoo
@File : app_user_location_impl
@Software: GoLand
@Description: protoc -I. -I%GOPATH%/src -ID:\GoWorks\src\github.com\grpc-ecosystem\grpc-gateway\third_party\googleapis --go_out=plugins=grpc:. talk_cloud_location.proto
              protoc -I. -I%GOPATH%/src -ID:\GoWorks\src\github.com\grpc-ecosystem\grpc-gateway\third_party\googleapis --grpc-gateway_out=logtostderr=true:. talk_cloud_location.proto
*/
package location

import (
	"context"
	"errors"
	pb "bo-server/api/proto"
	"bo-server/api/server/app"
	"bo-server/dao/pub"
	"bo-server/engine/cache"
	tl "bo-server/dao/location"
	twc "bo-server/dao/wifi_cache"
	"bo-server/engine/db"
	"bo-server/logger"
	"bo-server/utils"
	"net/http"
)

const (
	GSP_DATA_REPORT   = 1
	SOS_DATA_REPORT   = 2
	SOS_CANCLE_REPORT = 3

	MYSQL_STORE_SUCCESS = 1
	MYSQL_STORE_FAIL    = 0

	REDIS_UPDATE_SUCCESS = 1
	REDIS_UPDATE_FAIL    = 0

	WIFI_MODE = 1 // wifi定位
)

type TalkCloudLocationServiceImpl struct {
}

func (tcs *TalkCloudLocationServiceImpl) GetGpsData(ctx context.Context, req *pb.GPSHttpReq) (*pb.GPSHttpResp, error) {
	logger.Debugf("GetGpsData req : %v", req.Uid)
	res, _, err := tl.GetUserLocationInCache(req.Uid, cache.GetRedisClient())
	if err != nil {
		logger.Debugf("GetGpsData error: %+v", err)
	}
	return res, nil
}

type worker struct {
	dataStoreState   int
	updateRedisState int
	wifiIsValid      bool
}

// 处理上报数据
func (tcs *TalkCloudLocationServiceImpl) ReportGPSData(ctx context.Context, req *pb.ReportDataReq) (*pb.ReportDataResp, error) {
	// TODO 暂时认为存储到mysql之后，GPS数据一定更新成功, 更新不成，就去mysql查询最新一条数据出来
	var (
		gpsWorker = &worker{wifiIsValid: false}
		cacheInfo = &tl.CacheLocation{UpdateInfo: make([]interface{}, 0)}

		wifi  *pb.Wifi
		bssId string
	)
	//log.log.Debugf("================>req.LocationInfo: %+v", req.LocationInfo)
	logger.Debugf("ReportGPSData receiver type: %d data: %+v", req.DataType, req)

	// 0. sos cancel sos 没有数据也能上报
	if req.DeviceInfo == nil {
		return &pb.ReportDataResp{Res: &pb.Result{Msg: "req.DeviceInfo is not correct", Code: http.StatusUnprocessableEntity}}, errors.New("device info is null")
	}
	if req.DataType == SOS_DATA_REPORT {
		reportSosMsg(req)
		return &pb.ReportDataResp{Res: &pb.Result{Msg: "Receive SOS data success", Code: 200}}, nil
	}

	if req.DataType == SOS_CANCLE_REPORT {
		reportSosMsg(req)
		return &pb.ReportDataResp{Res: &pb.Result{Msg: "Receive SOS cancel data success", Code: 200}}, nil
	}

	// 1. 首先对数据进行参数校验
	match, err := preCheckData(req, gpsWorker)
	if !match {
		logger.Errorf("receiver data is invalid with:%v", err)
		return &pb.ReportDataResp{Res: &pb.Result{Msg: "params is not correct", Code: http.StatusUnprocessableEntity}}, err
	}
	cacheInfo.Id = req.DeviceInfo.Id

	// 2.0 TODO 根据上传的数据，采用对应的定位方式
	// 2.0.1 首先使用gps定位
	if req.LocationInfo.GpsInfo != nil {
		cacheInfo.UpdateInfo = append(cacheInfo.UpdateInfo,
			pub.LOCAL_TIME, req.LocationInfo.GpsInfo.LocalTime,
			pub.LATITUDE, req.LocationInfo.GpsInfo.Latitude,
			pub.LONGITUDE, req.LocationInfo.GpsInfo.Longitude,
			pub.SPEED_GPS, req.LocationInfo.GpsInfo.Speed,
			pub.COURSE, req.LocationInfo.GpsInfo.Course,
			pub.BATTERY, req.DeviceInfo.Battery, pub.WIFI_INFO, "")
		goto CONT
	}

	// 2.0.2 如果gps数据为空，就使用wifi信息进行处理
	if req.LocationInfo.GpsInfo == nil && req.LocationInfo.WifiInfo != nil {
		// pre: 1.先获取wifi对应的info 先去缓存查询，再去mysql查询
		logger.Debugf("start wifi query location %+v", req.LocationInfo.WifiInfo)
		bssId = twc.FindMostMatchWifiBssId(req.LocationInfo.WifiInfo)
		wifi, err = twc.GetWifiInfoFromCache(bssId)
		if err != nil && wifi != nil {
			logger.Debugf("GET WifiInfoFromCache with error:%+v", err)
		}

		if wifi == nil { // 说明没有web导入
			logger.Debugln("GET WifiInfoFromCache wifi info is nil")
		} else {
			cacheInfo.UpdateInfo = append(cacheInfo.UpdateInfo,
				pub.LATITUDE, wifi.Latitude,
				pub.LONGITUDE, wifi.Longitude,
				pub.WIFI_INFO, wifi.Des)
		}
		goto CONT
	}

	// 2.0.3 如果gps、wifi都为空，就使用蓝牙信息定位？ TODO
	if req.LocationInfo.GpsInfo == nil && req.LocationInfo.WifiInfo == nil &&
		req.LocationInfo.BtInfo != nil {
		//cacheInfo = &tl.CacheLocation{}
		goto CONT
	}
	// 2.0.5 全部为空那就返回
	err = errors.New("must have one location info at least")
	logger.Errorf("receiver data is invalid with:%v", err)
	return &pb.ReportDataResp{Res: &pb.Result{Msg: "params is not correct", Code: http.StatusUnprocessableEntity}}, err

CONT:
	// 3.0 存储到mysql中
	storeReportData(req, gpsWorker)

	// 3.1 更新缓存中GPS数据
	updateGPSDataByReq(cacheInfo, gpsWorker)

	if gpsWorker.updateRedisState == REDIS_UPDATE_FAIL {
		// 保证redis数据库里的数据一定是mysql中最新的那条记录
		updateGPSDataByMysql(req)
	}

	/*if req.DataType == SOS_DATA_REPORT {
		reportSosMsg(req)
	}

	if req.DataType == SOS_CANCLE_REPORT {
		reportSosMsg(req)
		return &pb.ReportDataResp{Res: &pb.Result{Msg: "Receive SOS cancel data success", Code: 200}}, nil
	}*/

	return &pb.ReportDataResp{Res: &pb.Result{Msg: "Receive data success", Code: 200}}, nil
}

// 存储到mysql上报数据
func storeReportData(req *pb.ReportDataReq, gw *worker) {
	if err := tl.InsertLocationData(req, db.DBHandler); err != nil {
		logger.Debugf("store data to mysql fail with error: %v", err)
		gw.dataStoreState = MYSQL_STORE_FAIL
	} else {
		gw.dataStoreState = MYSQL_STORE_SUCCESS
	}

}

// 更新缓存中的gps数据
func updateGPSDataByReq(cacheLocation *tl.CacheLocation, gw *worker) {
	mysqlState := gw.dataStoreState
	//logger.Debugf("cache receive gps data from app: %+v", cacheLocation)
	if mysqlState == MYSQL_STORE_SUCCESS {
		// 更新数据
		if err := tl.UpdateUserLocationInCache(cacheLocation, cache.GetRedisClient(), gw.wifiIsValid); err != nil {
			logger.Debugf("redis update data fail with error: %v", err)
			gw.updateRedisState = REDIS_UPDATE_FAIL
		} else {
			gw.updateRedisState = REDIS_UPDATE_SUCCESS
		}
	}

	// 暂时 如果mysql插入失败，就扔掉这个数据
	if mysqlState == MYSQL_STORE_FAIL {

	}
}

// 如果缓存更新失败，就去数据库里查询再来更新，估计不会出这样的问题
func updateGPSDataByMysql(req *pb.ReportDataReq) {
	// 去数据库查询数据

	// 更新缓存
}

// 校验数据合法性
func preCheckData(req *pb.ReportDataReq, w *worker) (bool, error) {
	if req == nil || req.IMei == "" || req.LocationInfo == nil || req.DeviceInfo == nil || req.DataType < 1 || req.DataType > 3 {
		return false, errors.New("param is invalid")
	}

	var wifiInfo, bssId = false, false
	// 1. imei 校验
	if !utils.CheckIMei(req.IMei) {
		return false, errors.New("IMei is invalid")
	}

	// 2. 设备位置信息校验
	if req.LocationInfo.GpsInfo == nil {
		logger.Debugln("req.LocationInfo.GpsInfo is invalid")
		//return false, errors.New("req.LocationInfo.GpsInfo is invalid")
	}

	if req.LocationInfo.GpsInfo != nil && (req.LocationInfo.GpsInfo.Latitude == 0 || req.LocationInfo.GpsInfo.Longitude == 0 ||
		req.LocationInfo.GpsInfo.LocalTime == 0 || req.LocationInfo.GpsInfo.Speed < 0 ||
		req.LocationInfo.GpsInfo.Course < 0) {
		logger.Debugln("detail info for req.LocationInfo.GpsInfo is invalid")
		//return false, errors.New("detail info for req.LocationInfo.GpsInfo is invalid")
	}

	/*if req.LocationInfo.BSInfo == nil {
		return false ,errors.New("req.LocationInfo.BSInfo is invalid")
	}
	if req.LocationInfo.BtInfo == nil {
		return false ,errors.New("req.LocationInfo.BtInfo is invalid")
	}*/
	if req.LocationInfo.WifiInfo == nil || len(req.LocationInfo.WifiInfo) < 1 {
		logger.Debugln("req.LocationInfo.WifiInfo is invalid")
		// 就让后面的wifi处理时，跳过，不要去访问 录入的wifi信息
		//return false, errors.New("req.LocationInfo.WifiInfo is invalid")
	} else {
		wifiInfo = true
	}

	for _, wifi := range req.LocationInfo.WifiInfo {
		if wifi.BssId == "" {
			logger.Debugln("req.LocationInfo.WifiInfo wifi.BssId is invalid")
			//return false, errors.New("req.LocationInfo.WifiInfo wifi.BssId is invalid")
			bssId = false
			break
		} else {
			bssId = true
		}
	}

	w.wifiIsValid = wifiInfo && bssId
	// 3.设备信息校验
	if req.DeviceInfo.Id <= 0 || req.DeviceInfo.Battery < 0 || req.DeviceInfo.DeviceType == 0 {
		return false, errors.New("req.DeviceInfo is invalid")
	}

	return true, nil
}

// 正常是要把这个消息写进rabbitMQ，暂时直接调用, TODO 当时为什么不直接调用而是去生成一个连接？
func reportSosMsg(req *pb.ReportDataReq) {
	logger.Debugf("start report sos with req:%+v", req)
	res, err := app.TalkCloudServiceImpl{}.ImSosPublish(context.Background(), req)
	if err != nil {
		logger.Errorf("sos msg publish error: %+v", err)
	}
	logger.Debugf("sos done!!! res: %+v", res)
}
