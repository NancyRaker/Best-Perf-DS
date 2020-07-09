/*
@Time : 2019/4/3 14:35
@Author : yanKoo
@File : user_cache
@Software: GoLand
@Description:
*/
package user_cache

import (
	pb "bo-server/api/proto"
	"bo-server/dao/pub"
	"bo-server/engine/cache"
	"bo-server/engine/db"
	"bo-server/logger"
	"bo-server/model"
	"bo-server/utils"
	"database/sql"
	"encoding/json"
	"strconv"
	//"encoding/json"
	"errors"
	"fmt"
	"github.com/gomodule/redigo/redis"
)

// 从缓存中获取群组列表
func GetUserIncludedInGroups(uId int32, rd redis.Conn) ([]int64, error) {
	if rd == nil {
		return nil, errors.New("redis conn is nil")
	}
	defer rd.Close()

	logger.Debugln("start get group list from redis")
	// 1.0 获取该设备在哪几个群组
	gIds, err := redis.Int64s(rd.Do("SMEMBERS", pub.MakeUserGroupKey(int32(uId))))
	if err != nil {
		return nil, err
	}
	logger.Debugf("the # %d device in :%+v", uId, gIds)

	if 0 == len(gIds) {
		logger.Debugf("user is not in any group\n")
		return nil, sql.ErrNoRows
	}

	return gIds, nil
}

// 获取单个群组中的用户列表信息
func GetGroupMemDataFromCache(gid int32, userMap map[int32]*pb.DeviceInfo, rd redis.Conn) ([]*pb.DeviceInfo, error) {
	if rd == nil {
		return nil, fmt.Errorf("rd is nil")
	}
	defer rd.Close()

	res := make([]*pb.DeviceInfo, 0)
	resOffline := make([]*pb.DeviceInfo, 0)
	key := pub.MakeGroupMemKey(gid)
	uids, err := redis.Int64s(rd.Do("SMEMBERS", key))
	if err != nil {
		return nil, fmt.Errorf("get members from %s error: %s", key, err)
	}

	logger.Debugf("group %d have user id : %+v in cache", gid, uids)

	sz := len(uids)
	if 0 == sz {
		logger.Debugf("group is not has any user\n")
		return nil, nil
	}

	memKeys := make([]interface{}, 0)
	for i := 0; i < sz; i++ {
		memKeys = append(memKeys)
	}

	// 获取缓存中某个群成员信息
	for _, v := range uids {
		uId := int32(v)
		if user, ok := userMap[int32(uId)]; ok {
			// 在线离线顺序
			if user.Online == pub.USER_ONLINE {
				res = append(res, user)
			} else {
				resOffline = append(resOffline, user)
			}
		} else {
			//getUserData(int32(uId), gid, rd, res, resOffline, userMap)
			user := &pb.DeviceInfo{}
			value, err := redis.Values(rd.Do("HMGET", pub.MakeUserDataKey(uId),
				"id", "imei", "nickname", "lock_gid"))
			if err != nil {
				logger.Debugln("hmget err:", err.Error())
			}
			//log.log.Debugf("Get group %d user info value string  : %s from cache ", gid, value)
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
			logger.Debugf("Get group %d user info : %+v from cache", gid, resStr)
			if value != nil {
				if value[0] != nil { // 只要任意一个字段为空就是没有这个数据
					uid, _ := strconv.Atoi(resStr[0])
					lockGId, _ := strconv.Atoi(resStr[3])

					user.Id = int32(uid)
					user.Imei = resStr[1]
					user.NickName = resStr[2]
					user.Online = online
					user.LockGroupId = int32(lockGId)
				} else {
					logger.Debugf("can't find user %d from redis", int(uId))
					UpdateUserFromDBToRedis(user, int(uId))
				}

				// 在线离线顺序
				if user.Online == pub.USER_ONLINE {
					res = append(res, user)
				} else {
					resOffline = append(resOffline, user)
				}
				userMap[uId] = user
			}
		}
	}
	res = append(res, resOffline...)
	return res, nil
}

func getUserData(uId, gid int32, rd redis.Conn, res, resOffline []*pb.DeviceInfo, userMap map[int32]*pb.DeviceInfo) {
	if rd == nil {
		return
	}
	defer rd.Close()
	user := &pb.DeviceInfo{}
	value, err := redis.Values(rd.Do("HMGET", pub.MakeUserDataKey(uId),
		"id", "imei", "nickname", "lock_gid"))
	if err != nil {
		fmt.Println("hmget err:", err.Error())
	}
	//log.log.Debugf("Get group %d user info value string  : %s from cache ", gid, value)
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
	logger.Debugf("Get group %d user info : %+v from cache", gid, resStr)
	if value != nil {
		if value[0] != nil && value[3] != nil { // 只要任意一个字段为空就是没有这个数据
			uid, _ := strconv.Atoi(resStr[0])
			lockGId, _ := strconv.Atoi(resStr[3])

			user.Id = int32(uid)
			user.Imei = resStr[1]
			user.NickName = resStr[2]
			user.Online = online
			user.LockGroupId = int32(lockGId)
		} else {
			logger.Debugf("can't find user %d from redis", int(uId))
			UpdateUserFromDBToRedis(user, int(uId))
		}

		// 在线离线顺序
		if user.Online == pub.USER_ONLINE {
			res = append(res, user)
		} else {
			resOffline = append(resOffline, user)
		}
		userMap[uId] = user
	}
}

func UpdateUserFromDBToRedis(user *pb.DeviceInfo, v int) {
	res, err := selectUserByKey(v)
	if err != nil {
		logger.Debugf("GetGroupMemDataFromCache UpdateUserFromDBToRedis selectUserByKey has error: %v", err)
		return
	}
	// 增加到缓存
	if err := AddUserDataInCache(int32(res.Id), []interface{}{
		pub.USER_Id, int32(res.Id),
		pub.IMEI, res.IMei,
		pub.USER_NAME, res.UserName,
		pub.NICK_NAME, res.NickName,
		pub.ONLINE, pub.USER_OFFLINE,
		pub.LOCK_GID, int32(res.LockGroupId),
	}, cache.GetRedisClient()); err != nil {
		logger.Error("UpdateUserFromDBToRedis Add user information to cache with error: ", err)
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
			logger.Error("Statement close fail")
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

	res, err := rd.Do("SADD", pub.MakeUserGroupKey(int32(uId)), gid)
	if err != nil {
		return err
	}
	logger.Error(res)
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
		_ = rd.Send("SADD", pub.MakeUserGroupKey(int32(v)), gId)
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

	res, err := rd.Do("SREM", pub.MakeUserGroupKey(int32(uId)), gid)
	if err != nil {
		return err
	}
	logger.Error(res)
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
		_ = rd.Send("SADD", pub.MakeUserGroupKey(int32(gl.Uid)), v.Gid)
	}
	_, err := rd.Do("EXEC")
	if err != nil {
		return err
	}
	return nil
}

// 向缓存添加用户信息数据
func AddUserDataInCache(uId int32, UpdateValuePair []interface{}, redisCli redis.Conn) error {
	if redisCli == nil {
		return errors.New("redis conn is nil")
	}
	defer redisCli.Close()

	value := []interface{}{pub.MakeUserDataKey(uId)}

	if _, err := redisCli.Do("HMSET", append(value, UpdateValuePair...)...); err != nil {
		return errors.New("hSet failed with error: " + err.Error())
	}
	return nil
}

// 更新用户所在默认用户组，就是更新用户data
func UpdateUserInfoInCache(uId int32, key, value interface{}, redisCli redis.Conn) error {
	if redisCli == nil {
		return errors.New("redis conn is nil")
	}
	defer redisCli.Close()

	if _, err := redisCli.Do("HSET", pub.MakeUserDataKey(uId), key, value); err != nil {
		return errors.New("UpdateUserInfoInCache hSet failed with error:" + err.Error())
	}
	return nil
}

// 获取一批用户的基本信息 TODO 这样取的值不对，会报空针
func GetUserBaseInfo(uIds []int32) []*pb.DeviceInfo {
	var (
		rd  = cache.GetRedisClient()
		res []*pb.DeviceInfo
	)
	if rd == nil {
		return res
	}
	defer rd.Close()

	for _, uId := range uIds {
		_ = rd.Send("HMGET", pub.MakeUserDataKey(int32(uId)),
			pub.USER_Id, pub.IMEI, pub.NICK_NAME, pub.LOCK_GID, pub.USER_TYPE, pub.DEVICE_TYPE, pub.BATTERY,
			// 通话质量类数据
			pub.USER_STATUS, pub.IN_LINK_QUALITY, pub.IN_MEDIA_LINK_QUALITY, pub.OUT_LINK_QUALITY, pub.OUT_MEDIA_LINK_QUALITY, pub.RTT)
	}
	_ = rd.Flush()
	for _, uId := range uIds {
		source, err := redis.Strings(rd.Receive())
		if err != nil {
			logger.Debugf("from redis # %d users source: %+v with err: %+v", uId, source, err)
			continue
		}
		logger.Debugf("from redis # %d users source: %+v with err: %+v", uId, source, err)
		user := &pb.DeviceInfo{}
		if source != nil {
			uid, _ := strconv.Atoi(source[0])
			user.Id = int32(uid)
			user.Imei = source[1]
			user.NickName = source[2]
			lockGId, _ := strconv.Atoi(source[3])
			user.LockGroupId = int32(lockGId)
			userType, _ := strconv.Atoi(source[4])
			user.UserType = int32(userType)
			user.DeviceType = source[5]
			// 心跳上传更新的电量
			battery, _ := strconv.Atoi(source[6])
			user.Battery = int32(battery)
			if len(source) > 7 {
				user.UserStatus = utils.StringToINT32(source[7])
				user.InLinkQuality = utils.StringToINT32(source[8])
				user.InMediaLinkQuality = utils.StringToINT32(source[9])
				user.OutLinkQuality = utils.StringToINT32(source[10])
				user.OutMediaLinkQuality = utils.StringToINT32(source[11])
				user.Rtt = utils.StringToINT32(source[12])
			}
		}
		res = append(res, user)
	}
	return res
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

// 获取群组在线信息 // 目前只用来获取stat
func QueryDevicesStatus(devices *[]*pb.DeviceInfo, statMap *map[int32]bool) error {
	if devices == nil {
		return errors.New("getDevicesAccount input is nil")
	}
	var (
		rd = cache.GetRedisClient()
	)
	if rd == nil {
		err := errors.New("GetAccountInfo getDevicesStatus rd is nil")
		return err
	}
	defer rd.Close()

	var totalUId []int32
	for _, device := range *devices {
		totalUId = append(totalUId, device.Id)
	}

	// uid在线状态
	for _, deviceInfo := range *devices {
		_ = rd.Send("GET", pub.GetRedisKey(deviceInfo.Id))
	}
	_ = rd.Flush()
	for _, deviceInfo := range *devices {
		source, err := redis.Bytes(rd.Receive())
		s := &model.SessionInfo{}
		_ = json.Unmarshal(source, s)
		// TODO 在线状态
		if err != nil {
			deviceInfo.Online = pub.USER_OFFLINE
		} else {
			(*statMap)[deviceInfo.Id] = true
			deviceInfo.Online = pub.USER_ONLINE
		}
	}

	return nil
}
