/*
@Time : 2019/10/21 18:28 
@Author : yanKoo
@File : nfc_report
@Software: GoLand
@Description: 调度员设置发送报告邮箱和是否发送月报周报时间
*/
package nfc_report

import (
	pb "bo-server/api/proto"
	"bo-server/engine/db"
	"bo-server/logger"
	"database/sql"
	"errors"
)

// 保存调度员应该发送月报和周报的时间
func SaveDeviceClockInfo(req *pb.ReportInfoReq) error {
	if db.DBHandler == nil {
		return errors.New("db conn is nil")
	}
	stmtInsG, err := db.DBHandler.Prepare(`INSERT INTO nfc_report (a_id, report_email, month_time, day_time) VALUES (?, ?, ?, ?)`)
	if err != nil {
		return err
	}
	defer stmtInsG.Close()

	_, err = stmtInsG.Exec(req.DetailParam.AccountId, req.DetailParam.ReportEmail, req.DetailParam.MonthTime, req.DetailParam.DayTime)
	if err != nil {
		logger.Error("Insert  error : ", err)
		return err
	}

	return nil
}


func DeleteDeviceClockInfo(req *pb.ReportInfoReq) error {
	if db.DBHandler == nil {
		return  errors.New("db conn is nil")
	}
	stmtInsG, err := db.DBHandler.Prepare(`DELETE from nfc_report where a_id = ?`)
	if err != nil {
		return err
	}
	defer stmtInsG.Close()

	_ = stmtInsG.QueryRow(req.DetailParam.AccountId)
	if err != nil {
		return  err
	}

	return  nil
}

func UpdateDeviceClockInfo(req *pb.ReportInfoReq) error {
	if db.DBHandler == nil {
		return errors.New("db conn is nil")
	}
	stmtInsG, err := db.DBHandler.Prepare(`UPDATE nfc_report SET report_email = ?, month_time = ?, day_time = ? WHERE a_id = ? `)
	if err != nil {
		return err
	}
	defer stmtInsG.Close()

	_, err = stmtInsG.Exec(req.DetailParam.ReportEmail, req.DetailParam.MonthTime, req.DetailParam.DayTime, req.DetailParam.AccountId)
	if err != nil {
		logger.Error("update  error : ", err)
		return err
	}

	return nil
}


func QueryDeviceClockInfo(req *pb.ReportSetInfoReq) (*pb.ReportSetDetail, error) {
	if db.DBHandler == nil {
		return nil, errors.New("db conn is nil")
	}

	var err error
	stmtOut, err := db.DBHandler.Prepare(`SELECT a_id, report_email, month_time, day_time FROM nfc_report WHERE a_id = ? `)
	if err != nil {
		return nil, err
	}
	defer stmtOut.Close()


	var res = &pb.ReportSetDetail{}
	if err = stmtOut.QueryRow(req.AccountId).Scan(
		&res.AccountId, &res.ReportEmail, &res.MonthTime, &res.DayTime); err != nil && err != sql.ErrNoRows{
		return nil, err
	} else {
		if err == sql.ErrNoRows {
			return nil, nil
		}
	}
	return res, nil
}