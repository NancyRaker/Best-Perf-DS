/*
@Time : 2019/4/3 14:35
@Author : yanKoo
@File : user_cache
@Software: GoLand
@Description:
*/
package user_cache

import (
	"database/sql"
	"strconv"
	"web-api/dao/pub"
	"web-api/engine/cache"
	"web-api/engine/db"
	"web-api/model"
	pb "web-api/proto/talk_cloud"
	//"encoding/json"
	"errors"
	"fmt"
	"github.com/gomodule/redigo/redis"
	"web-api/logger"
)

// Before when you change these constants
const (
	GRP_MEM_KEY_FMT    = "grp:%d:mem"
	GRP_DATA_KEY_FMT   = "grp:%d:data"
	USR_DATA_KEY_FMT   = "usr:%d:data"
	USR_STATUS_KEY_FMT = "usr:%d:stat"
	USR_GROUP_KEY_FMT  = "usr:%d:grps"

	USER_OFFLINE = 1 // 用户离线
	USER_ONLINE  = 2 // 用户在线
)

func MakeUserDataKey(uid int32) string {
	return fmt.Sprintf(USR_DATA_KEY_FMT, uid)
}


func MakeUserGroupKey(uid int32) string {
	return fmt.Sprintf(USR_GROUP_KEY_FMT, uid)
}

// 获取单个成员信息
func GetUserFromCache(uId int32) (*pb.DeviceInfo, error) {
	rd := cache.GetRedisClient()
	if rd == nil {
		return nil, errors.New("redis conn is nil")
	}
	defer rd.Close()

	user := &pb.DeviceInfo{}
	value, err := redis.Values(rd.Do("HMGET", pub.MakeUserDataKey(uId),
		pub.USER_Id, pub.IMEI, pub.NICK_NAME, pub.LOCK_GID, pub.ACCOUNT_ID, pub.DEVICE_TYPE))
	if err != nil {
		fmt.Println("hmget failed", err.Error())
	}
	logger.Debugf("Get %d user info value string  : %s from cache ", uId, value)

	var (
		valueStr string
		resStr   = make([]string, 0)
		online   int32
	)
	for _, v := range value {
		if v != nil {
			valueStr = string(v.([]byte))
			resStr = append(resStr, valueStr)
		} else {
			break // redis找不到，去数据库加载
		}
	}
	session, err := pub.GetSession(uId)
	if err != nil {
		logger.Debugf("get user online with err  with error: %+v", err.Error())
		online = pub.USER_OFFLINE
	} else {
		online = session.Online
	}
	logger.Debugf("Get %d user info : %v from cache", uId, resStr)
	if value != nil &&
		value[0] != nil && value[3] != nil && value[4] != nil { // 只要任意一个字段为空就是没有这个数据
		uid, _ := strconv.Atoi(resStr[0])
		lockGId, _ := strconv.Atoi(resStr[3])
		accountId, _ := strconv.Atoi(resStr[4])

		user.Id = int32(uid)
		user.Imei = resStr[1]
		user.NickName = resStr[2]
		user.Online = int32(online)
		user.LockGroupId = int32(lockGId)
		user.AccountId = int32(accountId)
		user.DeviceType = resStr[5]

	} else {
		logger.Debugf("can't find user %d from redis", int(uId))
		UpdateUserFromDBToRedis(user, int(uId))
	}

	logger.Debugf("# %d user info  : %+v from cache ", uId, user)
	return user, nil
}

func UpdateUserFromDBToRedis(user *pb.DeviceInfo, v int) {
	res, err := selectUserByKey(v)
	if err != nil {
		logger.Debugf("GetGroupMemDataFromCache UpdateUserFromDBToRedis selectUserByKey has error: %v", err)
		return
	}
	user.Id = int32(res.Id)
	user.Imei = res.IMei
	user.NickName = res.NickName
	user.Online = USER_OFFLINE
	user.LockGroupId = int32(res.LockGroupId)
	// 增加到缓存
	if err := AddUserDataInCache(&pb.Member{
		Id:          user.Id,
		IMei:        user.Imei,
		NickName:    user.NickName,
		Online:      user.Online,
		LockGroupId: user.LockGroupId,
	}, cache.GetRedisClient()); err != nil {
		logger.Debugln("UpdateUserFromDBToRedis Add user information to cache with error: ", err)
	}
}

// 通过关键词查找用户名
func selectUserByKey(key interface{}) (*model.User, error) {
	var stmtOut *sql.Stmt
	var err error
	switch t := key.(type) {
	case int:
		stmtOut, err = db.DBHandler.Prepare("SELECT id, name, nick_name, passwd, imei, user_type, pid, cid, lock_gid, create_time, last_login_time, change_time FROM `user` WHERE id = ?")
	case string:
		stmtOut, err = db.DBHandler.Prepare("SELECT id, name, nick_name, passwd, imei, user_type, pid, cid, lock_gid, create_time, last_login_time, change_time  FROM `user` WHERE name = ?")
	default:
		_ = t
		return nil, err
	}
	if err != nil {
		logger.Debugf("%s", err)
		return nil, err
	}
	defer func() {
		if err := stmtOut.Close(); err != nil {
			logger.Debugln("Statement close fail")
		}
	}()

	var (
		id, userType, cId, lockGId                                    int
		pId, userName, nickName, pwd, iMei, cTime, llTime, changeTime string
	)
	err = stmtOut.QueryRow(key).Scan(&id, &userName, &nickName, &pwd, &iMei, &userType, &pId, &cId, &lockGId, &cTime, &llTime, &changeTime)
	if err != nil {
		return nil, err
	}

	res := &model.User{
		Id:          id,
		IMei:        iMei,
		UserName:    userName,
		PassWord:    pwd,
		NickName:    nickName,
		UserType:    userType,
		ParentId:    pId,
		AccountId:   cId,
		LockGroupId: lockGId,
		CreateTime:  cTime,
		LLTime:      llTime,
		ChangeTime:  changeTime,
	}

	return res, nil
}

// 一个用户添加进组 可以在加载数据的时候用
func AddUserForSingleGroupCache(uId, gid int32, rd redis.Conn) error {
	if rd == nil {
		return errors.New("redis conn is nil")
	}
	defer rd.Close()

	res, err := rd.Do("SADD", MakeUserGroupKey(int32(uId)), gid)
	if err != nil {
		return err
	}
	logger.Debugln(res)
	return nil
}

// 一个用户在多个组， 用来更新，获取群组列表之后，去缓存中获取群组列表
func AddUsersGroupInCache(uid []int32, gId int32, rd redis.Conn) error {
	if rd == nil {
		return errors.New("redis conn is nil")
	}
	defer rd.Close()

	_ = rd.Send("MULTI")
	for _, v := range uid {
		_ = rd.Send("SADD", MakeUserGroupKey(int32(v)), gId)
	}
	_, err := rd.Do("EXEC")
	if err != nil {
		return err
	}
	return nil
}

// 移除单个用户
func RemoveUserForSingleGroupCache(uId, gid int32, rd redis.Conn) error {
	if rd == nil {
		return errors.New("redis conn is nil")
	}
	defer rd.Close()

	res, err := rd.Do("SREM", MakeUserGroupKey(int32(uId)), gid)
	if err != nil {
		return err
	}
	logger.Debugln(res)
	return nil
}

// 一个用户在多个组， 用来更新，获取群组列表之后，去缓存中获取群组列表
func AddUserInGroupToCache(gl *pb.GroupListRsp, rd redis.Conn) error {
	if rd == nil {
		return errors.New("redis conn is nil")
	}
	defer rd.Close()

	_ = rd.Send("MULTI")
	for _, v := range gl.GroupList {
		_ = rd.Send("SADD", MakeUserGroupKey(int32(gl.Uid)), v.Gid)
	}
	_, err := rd.Do("EXEC")
	if err != nil {
		return err
	}
	return nil
}

// 向缓存添加用户信息数据
func AddUserDataInCache(m *pb.Member, redisCli redis.Conn) error {
	if redisCli == nil {
		return errors.New("redis conn is nil")
	}
	defer redisCli.Close()
	//log.Log.Debugf(">>>>> start AddUserDataInCache")
	if _, err := redisCli.Do("HMSET", MakeUserDataKey(m.Id),
		"id", m.Id, "imei", m.IMei, "username", m.UserName, "nickname", m.NickName, "online", m.Online, "lock_gid", m.LockGroupId); err != nil {
		//log.Log.Debugf("AddUserDataInCache HMSET error: %+v",err)
		return errors.New("hSet failed with error: " + err.Error())
	}
	//log.Log.Debugf(">>>>> done AddUserDataInCache")
	return nil
}

