/*
@Time : 2019/7/1 15:04 
@Author : yanKoo
@File : login_record
@Software: GoLand
@Description:
*/
package login_record

import (
	"database/sql"
	"bo-server/engine/db"
	"bo-server/logger"
)

type LoginSaveData struct {
	UId        int32
	Ip         string
	LoginTime  int64
	LoginAddr  string
	AppVersion string
}

// 保存登录信息
func SaveDeviceLoginInfo(loginInfo *LoginSaveData) error {
	var (
		stmtIns *sql.Stmt
		err     error
	)

	if db.DBHandler == nil {
		logger.Debugln("SaveDeviceLoginInfo db.DBHandler is nil")
		return err
	}

	if stmtIns, err = db.DBHandler.Prepare("INSERT INTO login_record (uid, addr, im_addr, app_v, login_time) VALUES(?, ?, ?, ?, ?)"); err != nil {
		logger.Debugf("GetServerAddr stmtOut error :", err)
		return err
	}
	defer stmtIns.Close()

	if _, err = stmtIns.Exec(loginInfo.UId, loginInfo.Ip, loginInfo.LoginAddr, loginInfo.AppVersion, loginInfo.LoginTime); err != nil {
		logger.Debugf("GetServerAddr stmtOut.Query error :", err)
		return err
	}
	return nil
}
