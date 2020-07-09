/*
@Time : 2019/4/12 16:24
@Author : yanKoo
@File : location
@Software: GoLand
@Description: 存储设备发过来的的数据
*/
package location

import (
	pb "bo-server/api/proto"
	"bo-server/conf"
	"bo-server/engine/db"
	"bo-server/logger"
	"bo-server/utils"
	"database/sql"
	"fmt"
	"github.com/smartwalle/dbs"
	"strconv"
	"strings"
	"time"
)

const LIMIT = 1000 // 默认每次只查询1000条数据返回给web前端做轨迹回放
/*
id bigint(12) NULL记录id
uid bigint(12) NOT NULL设备/用户的id
local_time timestamp NULLGPS数据定位时间
lng varchar(64) NOT NULL经度
lat varchar(64) NOT NULL纬度
cse_sp varchar(128) NULL航向，速度

country varchar(255) NULL基站所在国家
operator varchar(255) NULL基站的运营商所在地区
region varchar(255) NULL基站所在的地区
bs_sth int(12) NULL基站的信号强度
wifi_sth int(12) NULLwifi的信号强度
bt_sth int(12) NULL蓝牙的信号强度
create_time timestamp NOT NULL记录存入数据库的时间
*/
// 插入GPS数据
func InsertLocationData(req *pb.ReportDataReq, db *sql.DB) error {
	logger.Debugf("receive gps data from app: %+v", req)
	ib := dbs.NewInsertBuilder()
	ib.Table("location")
	ib.SET("uid", req.DeviceInfo.Id)
	if req.LocationInfo.GpsInfo != nil {
		ib.SET("local_time", strconv.FormatUint(req.LocationInfo.GpsInfo.LocalTime, 10)) //utils.ConvertTimeUnix(req.LocationInfo.GpsInfo.LocalTime))
		ib.SET("lng", req.LocationInfo.GpsInfo.Longitude)
		ib.SET("lat", req.LocationInfo.GpsInfo.Latitude)
		ib.SET("cse_sp", PackCourseSpeed(req.LocationInfo.GpsInfo.Course, req.LocationInfo.GpsInfo.Speed))
	}
	if req.LocationInfo.BSInfo != nil {
		ib.SET("country", req.LocationInfo.BSInfo.Country)
		ib.SET("operator", req.LocationInfo.BSInfo.Operator)
		ib.SET("lac", req.LocationInfo.BSInfo.Lac)
		ib.SET("cid", req.LocationInfo.BSInfo.Cid)
		ib.SET("bs_sth",
			utils.FormatStrength(req.LocationInfo.BSInfo.FirstBs, req.LocationInfo.BSInfo.SecondBs,
				req.LocationInfo.BSInfo.ThirdBs, req.LocationInfo.BSInfo.FourthBs))
		/*ib.SET("bt_sth",
			utils.FormatStrength(req.LocationInfo.BtInfo.FirstBt, req.LocationInfo.BtInfo.SecondBt,
				req.LocationInfo.BtInfo.ThirdBt, req.LocationInfo.BtInfo.FourthBt))*/
	}
	if req.LocationInfo.WifiInfo != nil {
		var wifisStr string
		for _, wifi := range req.LocationInfo.WifiInfo {
			wifiStr := utils.FormatWifiInfo(wifi.BssId, wifi.Level)
			wifiStr += "|"
			wifisStr += wifiStr
		}
		ib.SET("wifi_sth", wifisStr[:len(wifisStr)-1])
	}
	if _, err := ib.Exec(db); err != nil {
		return err
	}
	return nil
}

// 打包航向和速度
func PackCourseSpeed(course, speed float32) string {
	return strconv.FormatFloat(float64(course), 'f', -1, 32) + "," +
		strconv.FormatFloat(float64(speed), 'f', -1, 32)
}

// 解析出航向和速度
func ParseCourseSpeed(cseSpeed string) (int32, float32) {
	strs := strings.Split(cseSpeed, ",")

	course, err := strconv.ParseInt(strs[0], 10, 32)
	if err != nil {
		logger.Error("parse course fail with error: ", err)
	}

	speed, err := strconv.ParseFloat(strs[1], 64)
	if err != nil {
		logger.Error("parse speed fail with error: ", err)
	}

	return int32(course), float32(speed)
}

// 根据起始时间戳返回limit条数据
func SelectTraceDataByLocalTimes(req *pb.GpsForTraceReq) ([]*pb.TraceInfo, error) {
	var (
		stmtOut   *sql.Stmt
		rows      *sql.Rows
		err       error
		traceData = make([]*pb.TraceInfo, 0)
	)

	if db.DBHandler == nil {
		logger.Error("db conn is nil")
		return nil, fmt.Errorf("db is nil")
	}

	if stmtOut, err = db.DBHandler.Prepare(`SELECT lng, lat, create_time FROM location
             WHERE UNIX_TIMESTAMP(create_time) > ? AND UNIX_TIMESTAMP(create_time) <= ?  AND lng IS NOT NULL AND uid = ? ORDER BY create_time ASC LIMIT ?`); err != nil {
		logger.Errorf("SelectTraceDataByLocalTime db.DBHandler.Prepare fail with error: %+v", err)
		return nil, err
	}

	if rows, err = stmtOut.Query(req.TimesStampStart, req.TimesStampEnd, req.Id, LIMIT); err != nil {
		logger.Errorf("SelectTraceDataByLocalTime stmtOut.Query fail with error: %+v", err)
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var (
			traceInfo = &pb.TraceInfo{}
			localtime time.Time
			//t string
		)
		if err = rows.Scan(&traceInfo.Lng, &traceInfo.Lat, &localtime); err != nil {
			logger.Errorf("SelectTraceDataByLocalTime rows.Scan fail with error: %+v", err)
			return nil, err
		}
		fmt.Println(localtime.Format(conf.TimeLayout))
		stamp, _ := time.ParseInLocation(conf.TimeLayout, localtime.Format(conf.TimeLayout), time.Local)
		traceInfo.Timestamp = strconv.FormatInt(stamp.Unix(), 10)
		traceData = append(traceData, traceInfo)
	}

	return traceData, nil
}

// 根据起始时间戳返回limit条数据
func SelectTraceDataByLocalTime(req *pb.GpsForTraceReq) ([]*pb.TraceInfo, error) {
	var (
		stmtOut   *sql.Stmt
		rows      *sql.Rows
		err       error
		traceData = make([]*pb.TraceInfo, 0)
	)

	if db.DBHandler == nil {
		logger.Error("db conn is nil")
		return nil, fmt.Errorf("db is nil")
	}

	/*if stmtOut, err = db.DBHandler.Prepare(`SELECT lng, lat, create_time FROM location
             WHERE UNIX_TIMESTAMP(create_time) > ? AND UNIX_TIMESTAMP(create_time) <= ?  AND lng IS NOT NULL AND uid = ? ORDER BY create_time ASC LIMIT ?`); err != nil {
		logger.Errorf("SelectTraceDataByLocalTime db.DBHandler.Prepare fail with error: %+v", err)
		return nil, err
	}*/

	if stmtOut, err = db.DBHandler.Prepare(`SELECT lng, lat, local_time FROM location
             WHERE local_time > ? AND local_time <= ?  AND lng IS NOT NULL AND uid = ? ORDER BY local_time ASC LIMIT ?`); err != nil {
		logger.Errorf("SelectTraceDataByLocalTime db.DBHandler.Prepare fail with error: %+v", err)
		return nil, err
	}

	if rows, err = stmtOut.Query(req.TimesStampStart, req.TimesStampEnd, req.Id, LIMIT); err != nil {
		logger.Errorf("SelectTraceDataByLocalTime stmtOut.Query fail with error: %+v", err)
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var (
			traceInfo = &pb.TraceInfo{}
		)
		if err = rows.Scan(&traceInfo.Lng, &traceInfo.Lat, &traceInfo.Timestamp); err != nil {
			logger.Errorf("SelectTraceDataByLocalTime rows.Scan fail with error: %+v", err)
			return nil, err
		}

		traceData = append(traceData, traceInfo)
	}

	return traceData, nil
}

func GetTimeStampIntervalTotal(req *pb.GpsForTraceReq) (int64, error) {
	var (
		stmtOut *sql.Stmt
		err     error
		total   int64
	)

	if db.DBHandler == nil {
		logger.Error("db conn is nil")
		return -1, fmt.Errorf("db is nil")
	}

	/*if stmtOut, err = db.DBHandler.Prepare(`SELECT COUNT(lng) FROM location
             WHERE UNIX_TIMESTAMP(create_time) > ? AND UNIX_TIMESTAMP(create_time) <= ? AND lng IS NOT NULL AND uid = ?`); err != nil {
		logger.Errorf("GetTimeStampIntervalTotal db.DBHandler.Prepare fail with error: %+v", err)
		return -1, err
	}*/

	if stmtOut, err = db.DBHandler.Prepare(`SELECT COUNT(lng) FROM location
		 WHERE local_time > ? AND local_time <= ? AND lng IS NOT NULL AND uid = ?`); err != nil {
		logger.Errorf("GetTimeStampIntervalTotal db.DBHandler.Prepare fail with error: %+v", err)
		return -1, err
	}

	if err = stmtOut.QueryRow(req.TimesStampStart, req.TimesStampEnd, req.Id).Scan(&total); err != nil {
		logger.Errorf("GetTimeStampIntervalTotal stmtOut.Query fail with error: %+v", err)
		return -1, err
	}

	return total, nil
}
