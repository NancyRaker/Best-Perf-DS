/**
* @Author: yanKoo
* @Date: 2019/3/16 15:29
* @Description:
 */
package device

import (
	"database/sql"
	"errors"
	"github.com/smartwalle/dbs"
	pb "bo-server/api/proto"
	cfgComm "bo-server/conf"
	"bo-server/engine/db"
	"bo-server/logger"
	"bo-server/model"
	"time"
)

// 批量导入设备
func ImportDevice(devices *[]*model.User) error {
	if db.DBHandler == nil {
		return errors.New("db conn is nil")
	}
	stmtInsG, err := db.DBHandler.Prepare(`INSERT INTO user (
                  imei, name, passwd, cid, pid, nick_name, user_type, last_login_time, create_time, change_time, d_type, active_time, sale_time) 
                  VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`)
	if err != nil {
		return err
	}
	defer stmtInsG.Close()

	for _, device := range *devices {
		// 一律以上海时区为准
		time.Local, _ = time.LoadLocation("Asia/Shanghai")
		t := time.Now()
		ctime := t.Format(cfgComm.TimeLayout)
		insGroupRes, err := stmtInsG.Exec(device.IMei, device.UserName, device.PassWord, device.AccountId, device.ParentId,
			device.NickName, 1, ctime, ctime, ctime, device.DeviceType, device.ActiveTime, device.SaleTime)
		if err != nil {
			logger.Error("Insert Group error : ", err)
			return err
		}
		if insGroupRes != nil {
			deviceId, _ := insGroupRes.LastInsertId()
			device.Id = int(deviceId)
		}
	}

	return nil
}

// 批量转移设备
func MultiUpdateDevice(req *pb.TransDevices) error {
	var ub = dbs.NewUpdateBuilder()
	ub.Table("user")
	ub.SET("cid", req.ReceiverId).SET( "lock_gid", 0)  // 把默认组置为零
	logger.Debugf("multi imeis: %+v", req.Imeis)
	ub.Where(dbs.IN("imei", req.Imeis))

	if _, err := ub.Exec(db.DBHandler); err != nil {
		logger.Error("multi update device error :", err)
		return err
	}
	return nil
}

func UpdateDeviceInfo(accountDevices *pb.DeviceUpdate) error {
	var ub = dbs.NewUpdateBuilder()
	ub.Table("user")
	if accountDevices.NickName != "" {
		ub.SET("nick_name", accountDevices.NickName)
	}
	if accountDevices.StartLog != "" {
		ub.SET("start_log", accountDevices.StartLog)
	}
	ub.Where("id = ?", accountDevices.Id)

	if _, err := ub.Exec(db.DBHandler); err != nil {
		logger.Error("update device info error :", err)
		return err
	}
	return nil
}

func SelectDeviceByImei(u *model.User) (int32, error) {
	stmtOut, err := db.DBHandler.Prepare("SELECT id FROM `user` WHERE imei = ? LIMIT 1")
	if err != nil {
		return -1, err
	}
	defer stmtOut.Close()

	var id int32
	if err = stmtOut.QueryRow(u.IMei).Scan(&id); err != nil && err != sql.ErrNoRows {
		return -1, err
	}

	if err == sql.ErrNoRows {
		return 0, nil
	}

	return id, nil
}
