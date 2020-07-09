/*
@Time : 2019/6/17 16:59 
@Author : yanKoo
@File : wifi
@Software: GoLand
@Description:
*/
package wifi

import (
	"database/sql"
	"errors"
	"github.com/smartwalle/dbs"
	pb "bo-server/api/proto"
	"bo-server/engine/db"
)

//导入wifi信息
func SaveWifiInfo(wifiObjs *pb.WifiInfoReq) error {
	if db.DBHandler == nil {
		return errors.New("sql db conn is nil")
	}

	var (
		res sql.Result
		err error
	)
	for i, wifi := range wifiObjs.Wifis {
		ib := dbs.NewInsertBuilder()
		ib.Table("wifi_info").Columns("bssid", "info", "cid", "lon", "lat")
		ib.Values(wifi.BssId, wifi.Des, wifiObjs.AccountId, wifi.Longitude, wifi.Latitude)
		if res, err = ib.Exec(db.DBHandler); err != nil {
			return err
		}
		id, err := res.LastInsertId()
		if err != nil {
			return err
		}
		wifiObjs.Wifis[i].Id = int32(id)
	}

	return nil
}

//查询wifi信息
func SelectWifiByBssId(bssId int64) (string, error) {
	return "", nil
}

// 删除wifi信息
func DelWifiInfo(wifiObjs *pb.WifiInfoReq) error {
	if db.DBHandler == nil {
		return errors.New("sql db conn is nil")
	}

	var (
		err     error
		delStat *sql.Stmt
	)

	delStat, err = db.DBHandler.Prepare("DELETE FROM `wifi_info` WHERE id = ?")
	if err != nil {
		return err
	}
	defer delStat.Close()
	for _, wifi := range wifiObjs.Wifis {
		if _, err = delStat.Exec(wifi.Id); err != nil {
			return err
		}
	}

	return nil
}

func UpdateWifiInfo(wifiObjs *pb.WifiInfoReq) error {
	if db.DBHandler == nil {
		return errors.New("sql db conn is nil")
	}

	var (
		updStmt *sql.Stmt
		err error
	)
	if updStmt, err = db.DBHandler.Prepare("UPDATE `wifi_info` SET `bssid` = ?, `info` = ?, `lon` = ?, `lat`= ? WHERE id = ?"); err != nil {
		return err
	}
	defer updStmt.Close()

	for _, w := range wifiObjs.Wifis {
		if _, err := updStmt.Exec(w.BssId, w.Des, w.Longitude,w.Latitude, w.Id); err != nil {
			return err
		}
	}
	return nil
}

func SelectWifibyKey(key interface{}) (*pb.Wifi, error) {
	if db.DBHandler == nil {
		return nil,
		errors.New("sql db conn is nil")
	}

	var (
		wifi = &pb.Wifi{}
		outStmt *sql.Stmt
		err error
	)

	switch key.(type) {
	case int32:
		outStmt, err = db.DBHandler.Prepare("SELECT bssid, info, lon, lat, id FROM wifi_info WHERE id = ?")
	case string:
		outStmt, err = db.DBHandler.Prepare("SELECT bssid, info, lon, lat, id FROM wifi_info WHERE bssid = ?")
	}

	if err = outStmt.QueryRow(key).Scan(wifi.BssId, wifi.Des, wifi.Longitude, wifi.Latitude, wifi.Id); err != nil {
		return nil, err
	}

	return wifi, nil
}