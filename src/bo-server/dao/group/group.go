/**
 * Copyrights (c) 2019. All rights reserved.
 * Group handlers
 * Author: tesion
 * Date: March 26 2019
 */
package group

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gomodule/redigo/redis"
	"github.com/smartwalle/dbs"
	pb "bo-server/api/proto"
	"bo-server/dao/pub"
	"bo-server/dao/user_friend"
	tfd "bo-server/dao/user_friend"
	"bo-server/engine/cache"
	"bo-server/engine/db"
	"bo-server/logger"
	"bo-server/model"
)

const (
	GROUP_MEMBER  = 1
	GROUP_MANAGER = 2
	GROUP_OWNER   = 3

	TEMP_GROUP = 2 // 临时组
	NORMAL_GROUP = 1 // 普通组

	CREATE_GROUP_BY_DISPATCHER = 1
	CREATE_GROUP_BY_USER       = 0
)

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

	sql := fmt.Sprintf("DELETE FROM user_group WHERE id=%d", gid)
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

// 获取该用户在哪几个群组
func GetGroupListFromDB(uid int32, db *sql.DB) (*pb.GroupListRsp, *map[int32]string, error) {
	if db == nil {
		return nil, nil, errors.New("db is nil")
	}

	stmtOut, err := db.Prepare("SELECT g.id, g.`stat`, g.group_name FROM user_group AS g INNER JOIN group_member AS gm ON g.id = gm.gid WHERE gm.uid = ?")

	rows, err := stmtOut.Query(uid)
	if err != nil {
		logger.Debugf("GetGroupListFromDB fail with err: %+v", err)
		return nil, nil, err
	}
	defer rows.Close()

	groups := &pb.GroupListRsp{Uid: uid, GroupList: nil}

	gMap := make(map[int32]string, 0)
	for rows.Next() {
		group := &pb.GroupInfo{}
		err = rows.Scan(&group.Gid, &group.Status, &group.GroupName)
		if err != nil {
			return nil, nil, err
		}

		// 获取群组的群主
		res, err := GetGroupUserId(group.Gid, GROUP_OWNER)
		if err != nil {
			return nil, nil, err
		}

		if res != nil && len(res) != 0 {
			group.GroupOwner = res[0]
		}

		// 获取群组管理员
		group.GroupManager, err = GetGroupUserId(group.Gid, GROUP_MANAGER)

		// 当前用户是否在组
		group.IsExist = true

		// 获取群组中有哪些人
		gMembers, err := GetGruopMembers(group.Gid, db)
		if err != nil {
			return nil, nil, err
		}
		// 查找当前用户好友，然后再群成员里面打标签
		_, fMap, err := user_friend.GetFriendReqList(uid, db)
		if err != nil {
			return nil, nil, err
		}

		for _, v := range gMembers {
			if _, ok := (*fMap)[v.Id]; ok {
				v.IsFriend = true
			}
		}
		group.UserList = gMembers
		gMap[group.Gid] = group.GroupName

		groups.GroupList = append(groups.GroupList, group)
		if err != nil {
			return nil, nil, err
		}
	}

	return groups, &gMap, nil
}

// 去mysql数据库获取群组的群组id
func GetGroupUserId(gid int32, roleType int32) ([]int32, error) {
	var (
		stmtOut *sql.Stmt
		rows    *sql.Rows
		err     error
		res     = make([]int32, 0)
		id      int32
	)
	db := db.DBHandler
	if db == nil {
		err = errors.New("db conn is nil")
		logger.Error(err)
		return res, err
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
	for rows.Next() {
		err = rows.Scan(&id)
		if err != nil {
			return nil, err
		}
		res = append(res, id)
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

// 查找当前群组所有的成员信息（在线信息去redis获取！！！）
func GetGruopMembers(gid int32, db *sql.DB) ([]*pb.DeviceInfo, error) {
	if db == nil {
		return nil, fmt.Errorf("db is nil")
	}
	stmtOut, err := db.Prepare(`SELECT u.id, u.imei, u.nick_name, u.user_type, u.lock_gid FROM user AS u 
										INNER JOIN group_member AS gm ON gm.uid=u.id WHERE gm.gid= ? AND gm.stat = 1`)

	rows, err := stmtOut.Query(gid)
	if err != nil {
		logger.Debugf("GetGruopMembers fail with error: %+v", err)
		return nil, err
	}
	defer rows.Close()

	grpMems := make([]*pb.DeviceInfo, 0)
	grpMemsOffline := make([]*pb.DeviceInfo, 0)

	for rows.Next() {
		gm := new(pb.DeviceInfo)
		err = rows.Scan(&gm.Id, &gm.Imei, &gm.NickName, &gm.UserType, &gm.LockGroupId)
		if err != nil {
			return nil, err
		}
		gm.IsFriend = false
		// 群成员的在线状态去redis取
		res, err := getSessionValue(gm.Id)
		if err != nil {
			gm.Online = pub.USER_OFFLINE
		} else {
			gm.Online = res.Online
		}
		if gm.Online == pub.USER_ONLINE {
			grpMems = append(grpMems, gm)
		} else {
			grpMemsOffline = append(grpMemsOffline, gm)
		}
	}
	grpMems = append(grpMems, grpMemsOffline...)
	return grpMems, nil
}

// 获取用户状态
func getSessionValue(value int32) (*model.SessionInfo, error) {
	logger.Debugln("getRedisKey(value, key):", pub.GetRedisKey(value))
	rd := cache.GetRedisClient()
	if rd == nil {
		err := errors.New("redis conn is null")
		logger.Debugf("CheckSession with error %+v", err)
		return nil, err
	}
	defer rd.Close()

	if resBytes, err := redis.Bytes(rd.Do("GET", pub.GetRedisKey(value))); err != nil {
		return nil, err
	} else {
		res := &model.SessionInfo{}
		if err := json.Unmarshal(resBytes, res); err != nil {
			logger.Debugln("json err")
			return nil, err
		}
		return res, nil
	}
}

// 查看
func SelectGroupsByAccountId(aid int) ([]*model.GroupInfo, error) {
	stmtOut, err := db.DBHandler.Prepare("SELECT id, group_name, stat, create_time FROM user_group WHERE account_id = ? AND stat = 1")
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := stmtOut.Close(); err != nil {
			logger.Error("statement close fail.")
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

// web创建群组
func WebCreateGroup(gl *pb.WebCreateGroupReq, userType int) (int64, error) {
	stmtInsG, err := db.DBHandler.Prepare("INSERT INTO user_group (group_name, account_id, user_type, stat) VALUES (?, ?, ?, ?)")
	if err != nil {
		return -1, err
	}
	defer stmtInsG.Close()

	insGroupRes, err := stmtInsG.Exec(gl.GroupInfo.GroupName, gl.GroupInfo.AccountId, userType, gl.GroupInfo.Status)
	if err != nil {
		logger.Error("Insert Group error : ", err)
		return -1, err
	}

	var (
		groupId int64
	)

	if insGroupRes != nil {
		groupId, _ = insGroupRes.LastInsertId()
	}

	var ib = dbs.NewInsertBuilder()
	ib.Table("group_member")
	ib.Columns("gid", "uid", "role_type")
	// 如果是1就是web用户 range 每个设备的id
	for _, v := range gl.DeviceIds {
		ib.Values(groupId, v, GROUP_MEMBER)
	}
	ib.Values(groupId, gl.GroupInfo.AccountId, GROUP_OWNER)

	stmtInsGD, value, err := ib.ToSQL()
	if err != nil {
		logger.Error("Error in ib ToSQL", err)
		// 2019/11/28 group_member操作失败，删除user_group
		_ = RemoveGroup( gl.GroupInfo.Id, db.DBHandler)
		return -1, err
	}

	_, err = db.DBHandler.Exec(stmtInsGD, value...)
	if err != nil {
		logger.Error("Error in insert group device", err)
		_ = RemoveGroup( gl.GroupInfo.Id, db.DBHandler)
		return -1, err
	}

	return groupId, nil
}

// 查找群组
func SelectGroupByKey(key interface{}) (*model.GroupInfo, error) {
	var stmtOut *sql.Stmt
	var err error
	switch t := key.(type) {
	case int:
		stmtOut, err = db.DBHandler.Prepare("SELECT id, group_name, account_id, stat, create_time FROM user_group WHERE id = ?")
	case string:
		stmtOut, err = db.DBHandler.Prepare("SELECT id, group_name, account_id, stat, create_time FROM user_group WHERE group_name = ?")
	default:
		_ = t
	}
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := stmtOut.Close(); err != nil {
			logger.Error("Statement close fail.")
		}
	}()

	var gid, accountId int
	var status, cTime, gName string
	if err := stmtOut.QueryRow(key).Scan(&gid, &gName, &accountId, &status, &cTime); err != nil {
		return nil, err
	}

	g := &model.GroupInfo{Id: gid, AccountId: accountId, GroupName: gName, Status: status, CTime: cTime}

	return g, nil
}

// 更新群组
func UpdateGroup(info *pb.GroupInfo, db *sql.DB) error {
	// 首先更新组，有更新group表，然后就去群里有几个设备，更新准备使用第三方库 目前web只用更新群的名字
	var ub = dbs.NewUpdateBuilder()
	ub.Table("user_group")
	ub.SET("group_name", info.GroupName)
	ub.Where("id = ? ", info.Gid)
	if _, err := ub.Exec(db); err != nil {
		logger.Error("update group name error：", err)
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
	defer stmtUpd.Close()

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

// 检查临时组
func CheckDuplicateCreateGroup(accountId int32) (int32, error) {
	stmtOut, err := db.DBHandler.Prepare("SELECT id FROM user_group WHERE stat = ? AND account_id = ?")
	if err != nil {
		logger.Debugf("DB error :%s", err)
		return -1, err
	}
	defer func() {
		if err := stmtOut.Close(); err != nil {
			logger.Error("Statement close fail")
		}
	}()

	var res int32
	if err = stmtOut.QueryRow(TEMP_GROUP, accountId).Scan(&res); err != nil && err != sql.ErrNoRows {
		logger.Debugf("err: %s", err)
		return -1, err
	}
	if err == sql.ErrNoRows {
		return 0, nil
	}
	return res, nil
}

// 更新组成员
func UpdateGroupMemberSS(gl *model.GroupList, userType int) (int64, error) {
	tx, err := db.DBHandler.Begin()
	if err != nil {
		return -1, err
	}
	// 如果是更新，opsType就是true
	var updGMAff, groupDeviceAff int64
	stmtUpdDG := "DELETE FROM group_member WHERE gid = ?"
	if err != nil {
		return -1, err
	}

	//log.log.Error(gl.GroupInfo.Id)
	if updDGRes, err := tx.Exec(stmtUpdDG, gl.GroupInfo.Id); err != nil {
		return -1, err
	} else {
		if updDGRes != nil {
			updGMAff, err = updDGRes.RowsAffected()
		}
	}

	var ib = dbs.NewInsertBuilder()
	ib.Table("group_member")
	ib.Columns("gid", "uid", "role_type")
	// 如果是1就是web用户 range 每个设备的id
	if userType == 1 {
		for _, v := range gl.DeviceInfo {
			logger.Debugf("%T", v)
			ib.Values(gl.GroupInfo.Id, (v.(map[string]interface{}))["id"], GROUP_MEMBER)
		}
		ib.Values(gl.GroupInfo.Id, gl.GroupInfo.AccountId, GROUP_OWNER)
	} else {
		for index, v := range gl.DeviceIds {
			if index == 0 { // 默认把创建群组的切片第一个作为管理员
				ib.Values(gl.GroupInfo.Id, v, GROUP_OWNER)
			}
			ib.Values(gl.GroupInfo.Id, v, GROUP_MEMBER)
		}
	}

	stmtInsGD, value, err := ib.ToSQL()
	if err != nil {
		logger.Error("Error in ib ToSQL", err)
		return -1, err
	}

	insGroupDeviceRes, err := tx.Exec(stmtInsGD, value...)
	if err != nil {
		logger.Error("Error in insert group device", err)
		return -1, err
	}

	if insGroupDeviceRes != nil {
		groupDeviceAff, _ = insGroupDeviceRes.RowsAffected()
	}

	logger.Error(updGMAff, groupDeviceAff, len(gl.DeviceIds)+1, len(gl.DeviceInfo)+1)

	if (updGMAff == updGMAff) && (groupDeviceAff == int64(len(gl.DeviceInfo)+1) || groupDeviceAff == int64(len(gl.DeviceIds)+1)) {
		if err := tx.Commit(); err != nil {
			logger.Error("tx commit")
			return -1, err
		}
	} else {
		logger.Error("rollback")
		if err := tx.Rollback(); err != nil {
			return -1, err
		}
		return -1, errors.New("rollback")
	}

	return int64(gl.GroupInfo.Id), nil
}

// 更新组成员
func UpdateGroupMember(removeDevice, addDevice []int32, gId int32) error {
	// 如果移除群成员切片不为空，就移除成员
	//if removeDevice == nil || len(removeDevice) < 0 || addDevice == nil || len(addDevice) < 0{
	//	return errors.New("removeDevice or addDevice is null")
	//}
	if len(removeDevice) > 0 {
		for _, id := range removeDevice {
			stmtDel, err := db.DBHandler.Prepare("DELETE FROM group_member WHERE gid = ? and uid = ?")
			if err != nil {
				return err
			}
			defer stmtDel.Close()

			//log.log.Error(gl.GroupInfo.Id)
			if _, err := stmtDel.Exec(gId, id); err != nil {
				return err
			}
			stmtDel.Close()
		}
	}
	if len(addDevice) > 0 {
		var ib = dbs.NewInsertBuilder()
		ib.Table("group_member")
		ib.Columns("gid", "uid", "role_type")

		for _, id := range addDevice {
			ib.Values(gId, id, GROUP_MEMBER)
		}

		if _, err := ib.Exec(db.DBHandler); err != nil {
			logger.Errorf("insert device error with : %+v", err)
			return err
		}
	}

	return nil
}

// 获取单个群组信息
func GetGroupInfoFromDB(gId, uId int32) (*pb.GroupInfo, error) {
	// 说明是直接去数据库模糊搜索的，就去数据库获取这个群组的信息和成员 TODO
	gInfo := &pb.GroupInfo{}
	g, err := SelectGroupByKey(int(gId))
	if err != nil {
		logger.Debugf("selete group  error: %v", err)
	}
	gInfo.Gid, gInfo.GroupName = int32(g.Id), g.GroupName
	if err != nil {
		return nil, err
	}

	// 获取群组的群主
	res, err := GetGroupUserId(gInfo.Gid, GROUP_OWNER)
	if err != nil {
		return nil, err
	}
	gInfo.GroupOwner = res[0]

	// 获取群组管理员
	gInfo.GroupManager, err = GetGroupUserId(gInfo.Gid, GROUP_MANAGER)

	// 当前用户是否在组
	gInfo.IsExist = true

	// 获取群组中有哪些人
	gMembers, err := GetGruopMembers(gInfo.Gid, db.DBHandler)
	if err != nil {
		return nil, err
	}
	// 查找当前用户好友，然后再群成员里面打标签
	_, fMap, err := tfd.GetFriendReqList(uId, db.DBHandler)
	if err != nil {
		return nil, err
	}

	for _, v := range gMembers {
		if _, ok := (*fMap)[v.Id]; ok {
			v.IsFriend = true
		}
	}
	gInfo.UserList = gMembers
	return gInfo, nil
}
