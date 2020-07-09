/**
 * Copyrights (c) 2019. All rights reserved.
 * Group handlers
 * Author: tesion
 * Date: March 26 2019
 */
package group

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/gomodule/redigo/redis"
	"github.com/smartwalle/dbs"
	"web-api/engine/db"
	"web-api/logger"
	"web-api/model"
	pb "web-api/proto/talk_cloud"
)

const (
	USER_OFFLINE = 1 // 用户离线
	USER_ONLINE  = 2 // 用户在线

	GROUP_MEMBER  = 1
	GROUP_MANAGER = 2
	GROUP_OWNER   = 3  // 群主

	USR_DATA_KEY_FMT   = "usr:%d:data"
	USR_STATUS_KEY_FMT = "usr:%d:stat"
)



func MakeUserStatusKey(uid int32) string {
	return fmt.Sprintf(USR_STATUS_KEY_FMT, uid)
}

func AddGroupMember(uid, gid int32, userType int, db *sql.DB) error {
	if db == nil {
		return fmt.Errorf("db is nil")
	}

	sql := fmt.Sprintf("INSERT INTO group_member(gid, uid, role_type) VALUES(%d, %d, %d)", gid, uid, userType)
	rows, err := db.Query(sql)
	if err != nil {
		logger.Debugf("query(%s), error(%s)", sql, err)
		return err
	}
	defer rows.Close()

	return nil
}

func RemoveGroupMember(uid, gid int32, db *sql.DB) error {
	if db == nil {
		return fmt.Errorf("db is nil")
	}

	sql := fmt.Sprintf("DELETE FROM group_member WHERE uid=%d AND gid=%d", uid, gid)
	_, err := db.Query(sql)
	if err != nil {
		logger.Debugf("query(%s), error(%s)", sql, err)
		return err
	}

	return nil
}

func RemoveGroup(gid int32, db *sql.DB) error {
	if db == nil {
		return fmt.Errorf("db is nil")
	}

	sql := fmt.Sprintf("DELETE FROM group WHERE id=%d", gid)
	_, err := db.Query(sql)
	if err != nil {
		logger.Debugf("remove group(%d) error: %s\n", gid, err)
		return err
	}

	return nil
}

func ClearGroupMember(gid int32, db *sql.DB) error {
	if db == nil {
		return fmt.Errorf("db is nil")
	}

	sql := fmt.Sprintf("DELETE FROM group_member WHERE gid=%d", gid)
	_, err := db.Query(sql)
	if err != nil {
		logger.Debugf("clear gruop(%d) user error: %s\n", gid, err)
		return err
	}

	return nil
}



// 去mysql数据库获取群组的群组id
func GetGroupManager(gid int32, db *sql.DB) (int32, error) {
	if db == nil {
		return -1, fmt.Errorf("db is nil")
	}

	stmtOut, err := db.Prepare("SELECT uid FROM group_member WHERE role_type = ? AND  gid = ? LIMIT 1")
	if err != nil {
		logger.Debugf("DB error :%s", err)
		return -1, err
	}
	defer func() {
		if err := stmtOut.Close(); err != nil {
			logger.Debugln("Statement close fail")
		}
	}()

	var res int32
	if err := stmtOut.QueryRow(GROUP_MANAGER, gid).Scan(&res); err != nil {
		logger.Debugf("GetGroupManager err: %s", err)
		return -1, nil
	}

	return res, nil
}

func SearchGroup(target string, db *sql.DB) (*pb.GroupListRsp, error) {
	if db == nil {
		return nil, fmt.Errorf("db is nil")
	}

	rows, err := db.Query("SELECT id, group_name FROM user_group WHERE group_name LIKE ?", "%"+target+"%")
	if err != nil {
		logger.Debugf("query error : %v\n", err)
	}
	defer rows.Close()

	groups := &pb.GroupListRsp{GroupList: nil}

	for rows.Next() {
		group := new(pb.GroupInfo)
		err = rows.Scan(&group.Gid, &group.GroupName)
		if err != nil {
			return nil, err
		}
		group.IsExist = false
		groups.GroupList = append(groups.GroupList, group)
	}

	return groups, nil
}

// 获取用户状态
func getUserStatusFromCache(uId int32, redisCli redis.Conn) (int32, error) {
	if redisCli == nil {
		return USER_OFFLINE, errors.New("redis conn is nil")
	}
	defer redisCli.Close()

	value, err := redis.Int(redisCli.Do("GET", MakeUserStatusKey(uId)))
	if err != nil {
		logger.Debugln("Get user online status fail with err: ", err.Error())
		return USER_OFFLINE, err
	}

	logger.Debugf("online value :%s", value)
	if value == 0 {
		return USER_OFFLINE, errors.New("no find")
	} else {
		return int32(value), nil
	}
}

// 去mysql数据库获取群组的群组id
func GetGroupUserId(gid int32, roleType int32) ([]int32, error) {
	var (
		stmtOut *sql.Stmt
		rows * sql.Rows
		err  error
		res = make([]int32, 0)
		id int32
	)
	db := db.DBHandler
	if db == nil {
		err = errors.New("db conn is nil")
		logger.Error(err)
		return res,err
	}

	if stmtOut, err = db.Prepare("SELECT uid FROM group_member WHERE role_type = ? AND  gid = ?"); err != nil {
		logger.Debugf("DB error :%s", err)
		return res, err
	}
	defer stmtOut.Close()

	if rows, err = stmtOut.Query(roleType, gid); err != nil {
		logger.Debugf("GetGroupUserId err: %s", err)
		return res, nil
	}
	logger.Debugf("GetGroupUserId rows: %+v", rows)
	for rows.Next() {
		err = rows.Scan(&id)
		if err != nil {
			return nil, err
		}
		res = append(res, id)
	}

	return res, nil
}

// 查看
func SelectGroupsByAccountId(aid int) ([]*model.GroupInfo, error) {
	stmtOut, err := db.DBHandler.Prepare("SELECT id, group_name, stat, create_time FROM user_group WHERE account_id = ? AND stat = 1")
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := stmtOut.Close(); err != nil {
			logger.Debugln("statement close fail.")
		}
	}()

	var res []*model.GroupInfo

	rows, err := stmtOut.Query(aid)
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		var gid int
		var groupName, status, cTime string
		if err := rows.Scan(&gid, &groupName, &status, &cTime); err != nil {
			return res, err
		}

		g := &model.GroupInfo{Id: gid, GroupName: groupName, AccountId: aid, Status: status, CTime: cTime}
		res = append(res, g)
	}

	return res, nil
}

// 查看
func GetGroupIdsByAccountId(aid int) ([]int32, error) {
	stmtOut, err := db.DBHandler.Prepare("SELECT id FROM user_group WHERE account_id = ? AND stat = 1")
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := stmtOut.Close(); err != nil {
			logger.Debugln("statement close fail.")
		}
	}()

	var res []int32

	rows, err := stmtOut.Query(aid)
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		var gid int32
		if err := rows.Scan(&gid); err != nil {
			return res, err
		}
		res = append(res, gid)
	}

	return res, nil
}

// 更新群组 TODO
func UpdateGroup(info *model.GroupInfo, db *sql.DB) error {
	// 首先更新组，有更新group表，然后就去群里有几个设备，更新准备使用第三方库 目前web只用更新群的名字
	var ub = dbs.NewUpdateBuilder()
	ub.Table("user_group")
	ub.SET("group_name", info.GroupName)
	ub.Where("id = ? ", info.Id)
	if _, err := ub.Exec(db); err != nil {
		logger.Debugln("update group name error：", err)
		return err
	}
	return nil
}

// 删除群组
func DeleteGroup(g *model.GroupInfo) error {
	tx, err := db.DBHandler.Begin()
	if err != nil {
		return err
	}
	stmtUpd, err := tx.Prepare("DELETE FROM user_group WHERE id = ?")
	if err != nil {
		return err
	}

	if _, err := stmtUpd.Exec(g.Id); err != nil {
		return err
	}

	stmtUpdDG, err := tx.Prepare("DELETE FROM group_member WHERE gid = ?")
	if err != nil {
		return err
	}

	if _, err := stmtUpdDG.Exec(g.Id); err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return err
	}
	return nil
}

func CheckDuplicateGName(g *model.GroupInfo) (int, error) {
	stmtOut, err := db.DBHandler.Prepare("SELECT count(id) FROM user_group WHERE group_name = ? AND account_id = ?")
	if err != nil {
		logger.Debugf("DB error :%s", err)
		return -1, err
	}
	defer func() {
		if err := stmtOut.Close(); err != nil {
			logger.Debugln("Statement close fail")
		}
	}()

	var res int
	if err := stmtOut.QueryRow(g.GroupName, g.AccountId).Scan(&res); err != nil {
		logger.Debugf("err: %s", err)
		return -1, err
	}

	return res, nil
}
