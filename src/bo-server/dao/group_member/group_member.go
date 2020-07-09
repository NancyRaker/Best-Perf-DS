/**
* @Author: yanKoo
* @Date: 2019/3/18 11:34
* @Description:
 */
package group_member

import (
	"database/sql"
	"github.com/smartwalle/dbs"
	pb "bo-server/api/proto"
	"bo-server/engine/db"
	"bo-server/logger"
	"bo-server/model"
)

func SelectDevicesByGroupId(gid int) ([]*model.User, error) {
	stmtOut, err := db.DBHandler.Prepare(`SELECT id, imei, name, passwd, user_type, cid, create_time, last_login_time, change_time 
									FROM user WHERE id IN (SELECT uid FROM group_member WHERE gid = ?) AND user_type = 1`)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := stmtOut.Close(); err != nil {
			logger.Error("Statement close fail")
		}
	}()

	var res []*model.User
	rows, err := stmtOut.Query(gid)
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		var id, accountId, userType int
		var iMei, userName, pwd string
		var cTime, llTime, changeTime string
		if err := rows.Scan(&id, &iMei, &userName, &pwd, &userType, &accountId, &cTime, &llTime, &changeTime); err != nil {
			return res, err
		}

		d := &model.User{
			Id: id, IMei: iMei,
			UserName:  userName, //PassWord: pwd,
			AccountId: accountId,
			UserType:  userType,
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
			logger.Error("statement close fail.")
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

// 更新群组管理员
func UpdateGroupManager(gmr *pb.UpdGManagerReq, db *sql.DB) error {
	for _, v := range gmr.UId {
		var ub = dbs.NewUpdateBuilder()
		ub.Table("group_member")
		ub.SET("role_type", gmr.RoleType)
		// 首先更新组管理员
		ub.Where("gid = ? AND uid = ?", gmr.GId, v)
		if _, err := ub.Exec(db); err != nil {
			logger.Error("update group name error：", err)
			return err
		}
	}
	return nil
}
