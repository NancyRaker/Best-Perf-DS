/*
@Time : 2019/10/22 10:21 
@Author : yanKoo
@File : nfc_device_clock
@Software: GoLand
@Description:
*/
package nfc_device_clock

import (
	pb "bo-server/api/proto"
	"bo-server/engine/db"
	"bo-server/logger"
	"errors"
)

// 保存设备打卡时间
func SaveNFCDeviceClockRecord(req *pb.ClockData) error {
	if db.DBHandler == nil {
		return errors.New("db conn is nil")
	}
	stmtInsG, err := db.DBHandler.Prepare(`INSERT INTO nfc_device_clock (uid, tag_id, clock_time, time_zone) VALUES (?, ?, ?, ?)`)
	if err != nil {
		return err
	}
	defer stmtInsG.Close()

	_, err = stmtInsG.Exec(req.DeviceId, req.TagId, req.ClockTime, req.TimeZone)
	if err != nil {
		logger.Error("Insert  error : ", err)
		return err
	}

	return nil
}

// TODO 查询设备打卡时间
func QueryNFCDeviceClockRecord(req *pb.ClockStatusReq) ([]*pb.ClockData, error) {
	return nil, nil
}
