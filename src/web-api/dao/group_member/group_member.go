/**
* @Author: yanKoo
* @Date: 2019/3/18 11:34
* @Description:
 */
package group_member

import (
	"web-api/engine/db"
	"web-api/logger"
	"web-api/model"
)



func SelectDevicesByGroupId(gid int) ([]*model.User, error) {
	stmtOut, err := db.DBHandler.Prepare(`SELECT id, imei, name,nick_name, passwd, user_type, d_type, cid, create_time, last_login_time, change_time 
									FROM user WHERE id IN (SELECT uid FROM group_member WHERE gid = ?) AND user_type = 1`)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := stmtOut.Close(); err != nil {
			logger.Debugln("Statement close fail")
		}
	}()

	var res []*model.User
	rows, err := stmtOut.Query(gid)
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		var id, accountId, userType int
		var iMei, userName, nickName, pwd, d_type string
		var cTime, llTime, changeTime string
		if err := rows.Scan(&id, &iMei, &userName, &nickName, &pwd, &userType, &d_type, &accountId, &cTime, &llTime, &changeTime); err != nil {
			return res, err
		}

		d := &model.User{
			Id: id, IMei: iMei,
			UserName: userName, NickName: nickName, //PassWord: pwd,
			AccountId: accountId,
			UserType:  userType,
			DeviceType:d_type,
			//Status:    status, ActiveStatus: aStatus, BindStatus: bindStatus,
			CreateTime: cTime, LLTime: llTime, ChangeTime: changeTime,
		}
		res = append(res, d)
	}

	return res, nil
}

func SelectDeviceIdsByGroupId(gid int) ([]int, error) {
	stmtOut, err := db.DBHandler.Prepare("SELECT uid FROM group_member WHERE gid = ?")
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := stmtOut.Close(); err != nil {
			logger.Debugln("statement close fail.")
		}
	}()

	var res []int
	rows, err := stmtOut.Query(gid)
	if err != nil {
		return res, err
	}

	for rows.Next() {
		var id int
		if err := rows.Scan(&id); err != nil {
			return res, err
		}
		res = append(res, id)
	}

	return res, nil
}
