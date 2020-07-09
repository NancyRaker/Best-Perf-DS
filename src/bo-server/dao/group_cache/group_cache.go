/**
 * Copyright (c) 2019. All rights reserved.
 * some functions deal with cache data of the groups
 * Author: tesion
 * Date: March 29th 2019
 */
package group_cache

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gomodule/redigo/redis"
	pb "bo-server/api/proto"
	"bo-server/dao/group"
	tlc "bo-server/dao/location"
	"bo-server/dao/pub"
	"bo-server/dao/user_cache"
	"bo-server/engine/cache"
	"bo-server/logger"
	"bo-server/model"
	"bo-server/utils"
	"strconv"
)

// check the key whether exists or not
func IsKeyExists(key string, rd redis.Conn) (bool, error) {
	if rd == nil {
		return false, fmt.Errorf("rd is null")
	}

	reply, err := redis.Int(rd.Do("EXISTS", key))
	if err != nil {
		logger.Debugf("check key(%s) exists error: %s\n", key, err)
		return false, nil
	}

	if 0 == reply {
		return false, nil
	}

	return true, nil
}

// add new group data to cache
func AddGroupCache(ur []*pb.DeviceInfo, gInfo *pb.GroupInfo, rd redis.Conn) error {
	if rd == nil {
		return errors.New("rd is null")
	}
	defer rd.Close()
	grpData, err := json.Marshal(pb.GroupInfo{
		Gid:          gInfo.Gid,
		GroupName:    gInfo.GroupName,
		GroupManager: gInfo.GroupManager,
		GroupOwner:   gInfo.GroupOwner})
	if err != nil {
		logger.Debugf("json marshal error: %s\n", err)
		return err
	}

	grpKey := pub.MakeGroupDataKey(gInfo.Gid)
	memKey := pub.MakeGroupMemKey(gInfo.Gid)

	_ = rd.Send("MULTI")
	for _, v := range ur {
		_ = rd.Send("SADD", memKey, v.Id)
	}
	_ = rd.Send("SET", grpKey, grpData)

	_, err = rd.Do("EXEC")
	if err != nil {
		logger.Debugf("add group to cache error: %s\n", err)
		return err
	}
	//log.log.Debugf("reply:%T", reply)

	return nil
}

// add new member to the group
func AddGroupSingleMemCache(gid, uid int32, rd redis.Conn) error {
	if rd == nil {
		return fmt.Errorf("rd is nil")
	}
	defer rd.Close()

	key := pub.MakeGroupMemKey(gid)
	_, err := rd.Do("SADD", key, uid)
	if err != nil {
		return fmt.Errorf("add new group key(%s) error: %s", key, err)
	}

	return nil
}

// remove new member to the group
func RemoveGroupSingleMemCache(gid, uid int32, rd redis.Conn) error {
	if rd == nil {
		return fmt.Errorf("rd is nil")
	}
	defer rd.Close()

	key := pub.MakeGroupMemKey(gid)
	_, err := rd.Do("SREM", key, uid)
	if err != nil {
		return fmt.Errorf("remove new group key(%s) error: %s", key, err)
	}

	return nil
}

// 往群组成员里加成员
func AddGroupMemsInCache(gid int32, uids []int32, rd redis.Conn) error {
	if rd == nil {
		return fmt.Errorf("rd is nil")
	}
	defer rd.Close()

	key := pub.MakeGroupMemKey(gid)
	_ = rd.Send("MULTI")
	for _, v := range uids {
		_ = rd.Send("SADD", key, v)
	}
	_, err := rd.Do("EXEC")
	if err != nil {
		return fmt.Errorf("add new group key(%s) error: %s", key, err)
	}

	return nil
}

func GetGroupMem(gid int32, rd redis.Conn) ([]int64, error) {
	if rd == nil {
		return nil, fmt.Errorf("rd is nil")
	}
	key := pub.MakeGroupMemKey(gid)
	uids, err := redis.Int64s(rd.Do("SMEMBERS", key))
	if err != nil {
		return nil, fmt.Errorf("get members from %s error: %s", key, err)
	}

	return uids, nil
}

// 更新群组的信息
func AddGroupInCache(gl *pb.GroupListRsp, rd redis.Conn) error {
	if rd == nil {
		return errors.New("rd is null")
	}
	defer rd.Close()

	_ = rd.Send("MULTI")
	for _, v := range gl.GroupList {
		grpData, err := json.Marshal(&pb.GroupInfo{
			Gid: v.Gid, GroupName: v.GroupName, GroupManager: v.GroupManager, GroupOwner: v.GroupOwner})
		if err != nil {
			logger.Debugf("json marshal error: %s\n", err)
			return err
		}
		grpKey := pub.MakeGroupDataKey(v.Gid)
		_ = rd.Send("SET", grpKey, grpData)
	}

	if _, err := rd.Do("EXEC"); err != nil {
		logger.Debugf("Add group data to cache error: %s\n", err)
		return err
	}

	return nil
}

func AddGroupAndUserInCache(gl *model.GroupList, rd redis.Conn) error {
	if rd == nil {
		return errors.New("rd is null")
	}
	defer rd.Close()
	logger.Error("start add create group to cache")

	grpData, err := json.Marshal(&pb.GroupInfo{Gid: int32(gl.GroupInfo.Id), GroupName: gl.GroupInfo.GroupName, GroupManager: make([]int32, 0),
		GroupOwner: int32(gl.GroupInfo.AccountId)})
	if err != nil {
		logger.Debugf("json marshal error: %s\n", err)
		return err
	}

	grpKey := pub.MakeGroupDataKey(int32(gl.GroupInfo.Id))
	memKey := pub.MakeGroupMemKey(int32(gl.GroupInfo.Id))

	// TODO redis 错误处理
	_ = rd.Send("MULTI")

	// 1. 新建一个组，涉及到的是一个组加入了很多个成员，就有一个groupDataKey值和一个memberKey，
	for _, v := range gl.DeviceInfo {
		_ = rd.Send("SADD", memKey, v.(map[string]interface{})["id"])

	}
	// 2.更新每一个userGroups的key里面的组数
	for _, v := range gl.DeviceIds {
		_ = rd.Send("SADD", pub.MakeUserGroupKey(int32(v)), gl.GroupInfo.Id)
	}
	_ = rd.Send("SET", grpKey, grpData)

	_, err = rd.Do("EXEC")
	if err != nil {
		logger.Debugf("Add group to cache error: %s\n", err)
		return err
	}

	return nil
}

// web 创建群组的时候更新缓存
func WebAddGroupAndUserInCache(gl *pb.WebCreateGroupReq, rd redis.Conn) error {
	if rd == nil {
		return errors.New("rd is null")
	}
	defer rd.Close()
	logger.Error("start add create group to cache")

	grpData, err := json.Marshal(&pb.GroupInfo{Gid: int32(gl.GroupInfo.Id), Status: gl.GroupInfo.Status,
		GroupName: gl.GroupInfo.GroupName,
		GroupManager: make([]int32, 0),
		GroupOwner: int32(gl.GroupInfo.AccountId)})
	if err != nil {
		logger.Debugf("json marshal error: %s\n", err)
		return err
	}

	grpKey := pub.MakeGroupDataKey(int32(gl.GroupInfo.Id))
	memKey := pub.MakeGroupMemKey(int32(gl.GroupInfo.Id))

	// TODO redis 错误处理
	_ = rd.Send("MULTI")

	// 1. 新建一个组，涉及到的是一个组加入了很多个成员，就有一个groupDataKey值和一个memberKey，
	for _, v := range gl.DeviceIds {
		_ = rd.Send("SADD", memKey, v)

	}
	// 2.更新每一个userGroups的key里面的组数
	for _, v := range gl.DeviceIds {
		_ = rd.Send("SADD", pub.MakeUserGroupKey(int32(v)), gl.GroupInfo.Id)
	}
	_ = rd.Send("SET", grpKey, grpData)

	_, err = rd.Do("EXEC")
	if err != nil {
		logger.Debugf("Add group to cache error: %s\n", err)
		return err
	}

	return nil
}

func UpdateGroupAndUserMemberInCache(remove, add []int32, gId int32) error {
	rd := cache.GetRedisClient()
	if rd == nil {
		return errors.New("rd is null")
	}
	defer rd.Close()
	logger.Error("start add create group to cache")

	memKey := pub.MakeGroupMemKey(int32(gId))

	// TODO redis 错误处理
	_ = rd.Send("MULTI")

	// 1. 新建一个组，涉及到的是一个组加入了很多个成员，就有一个groupDataKey值和一个memberKey，
	for _, v := range add {
		_ = rd.Send("SADD", memKey, v)
		_ = rd.Send("SADD", pub.MakeUserGroupKey(int32(v)), gId)
	}

	for _, v := range remove {
		_ = rd.Send("SREM", memKey, v)
		// 2.更新每一个userGroups的key里面的组数
		_ = rd.Send("SREM", pub.MakeUserGroupKey(int32(v)), gId)
	}

	if _, err := rd.Do("EXEC"); err != nil {
		logger.Debugf("Add group to cache error: %s\n", err)
		return err
	}

	return nil
}

func GetGroupListInfos(gIds []int64, rd redis.Conn) (groupList []*pb.GroupInfo, err error) {
	if rd == nil {
		return nil, errors.New("redis is nil")
	}
	defer rd.Close()

	// 1.0 获取所有的缓存中所有的群组信息
	var (
		gDataKeys = make([]interface{}, 0)
		gMemKeys  = make([]interface{}, 0)
	)

	if gIds == nil || len(gIds) == 0 {
		return nil, errors.New("can't get nil group id")
	}

	for _, v := range gIds {
		gDataKeys = append(gDataKeys, pub.MakeGroupDataKey(int32(v)))
		gMemKeys = append(gMemKeys, pub.MakeGroupMemKey(int32(v)))
	}
	groups, err := redis.Strings(rd.Do("MGET", gDataKeys...))

	// 2.1 去重, 找出需要查询的用户
	var (
		groupIdMaps  = make(map[int32][]int)
		uIdsMap      = make(map[int32]bool)
		totalUId     []int32
		userInfoMap  = make(map[int32]*pb.DeviceInfo)
		userInfoList = make([]*pb.DeviceInfo, 0)
		statMap      = make(map[int32]bool)
	)
	// 2.2 批量查询群组里面有多少人
	for _, key := range gMemKeys {
		_ = rd.Send("SMEMBERS", key)
	}
	_ = rd.Flush()
	for i := range gMemKeys {
		ids, err := redis.Ints(rd.Receive())
		groupIdMaps[int32(gIds[i])] = ids
		for _, id := range ids {
			uIdsMap[int32(id)] = true
		}
		logger.Debugf("from redis # %d user id: %+v with err: %+v", i, ids, err)
	}
	//log.log.Debugf("from redis groupIdMaps: %+v", groupIdMaps)

	// 2.3 计算所有群组存在过的设备
	for id, ok := range uIdsMap {
		if ok {
			totalUId = append(totalUId, id)
		}
	}

	//log.log.Debugf("from redis groups: %+v", groups)
	//log.log.Debugf("from redis totalUId: %+v", totalUId)

	// 3.0 批量查询uid的信息和在线状态
	for _, uId := range totalUId {
		_ = rd.Send("HMGET", pub.MakeUserDataKey(int32(uId)),
			pub.USER_Id, pub.IMEI, pub.NICK_NAME, pub.LOCK_GID, pub.USER_TYPE, pub.DEVICE_TYPE)
	}
	_ = rd.Flush()
	for _, uId := range totalUId {
		source, err := redis.Strings(rd.Receive())
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

			user.Online = pub.USER_OFFLINE // 1是不在线
			user.DeviceType = source[5]
		}
		userInfoMap[uId] = user
		userInfoList = append(userInfoList, user)
		logger.Debugf("from redis # %d users source: %+v with err: %+v", uId, source, err)
	}
	// 3.0.1 获取设备的经纬度等位置信息
	if err := tlc.GetUsersLocationFromCache(&userInfoList, cache.GetRedisClient()); err != nil {
		logger.Debugf("GetGroupListInfos GetUsersLocationFromCache err: %+v", err)
	}
	// 3.1 uid在线状态
	for _, uId := range totalUId {
		_ = rd.Send("GET", pub.GetRedisKey(int32(uId)))
	}
	_ = rd.Flush()
	for _, id := range totalUId {
		source, err := redis.Bytes(rd.Receive())
		s := &model.SessionInfo{}
		_ = json.Unmarshal(source, s)
		// TODO 在线状态
		if err != nil {
			statMap[int32(id)] = false
		} else {
			statMap[int32(id)] = true
		}
	}
	//log.log.Printf("current map :%+v", statMap)

	for _, v := range groups {
		gInfo := &pb.GroupInfo{}
		logger.Debugf("find group name %+v", v)
		if v == "" {
			// TODO 会报空针
			continue
		}
		err = json.Unmarshal([]byte(v), gInfo)
		logger.Debugf("Get Group info from cache: %+v", gInfo)
		if err != nil {
			logger.Debugf("json parse user data(%s) error: %s\n", string(v), err)
			return nil, err
		}

		var groupManagerMap = make(map[int32]bool)
		for _, managerId := range gInfo.GroupManager {
			groupManagerMap[managerId] = true
		}

		var offlineList = make([]*pb.DeviceInfo, 0)
		for _, idInt := range groupIdMaps[gInfo.Gid] {
			id := int32(idInt)
			user := userInfoMap[id]
			if groupManagerMap[id] {
				user.GrpRole = group.GROUP_MANAGER
			} else if id == gInfo.GroupOwner {
				user.GrpRole = group.GROUP_OWNER
			} else {
				user.GrpRole = group.GROUP_MEMBER
			}

			// 增加状态信息
			if statMap[id] {
				user.Online = pub.USER_ONLINE
				if user.LockGroupId == gInfo.Gid {
					var newUser = &pb.DeviceInfo{}
					_ = utils.DeepCopy(newUser, user)
					newUser.Online = pub.USER_JANUS_ONLINE
					gInfo.UserList = append(gInfo.UserList, newUser)
				} else {
					gInfo.UserList = append(gInfo.UserList, user)

				}
			} else {
				offlineList = append(offlineList, user)
			}
		}
		gInfo.UserList = append(gInfo.UserList, offlineList...)
		groupList = append(groupList, gInfo)
	}
	return
}

func sortByMemberType(gInfo *pb.GroupInfo, user *[]*pb.DeviceInfo) {

}

// 获取单个群组 @Deprecated
func GetGroupInfoFromCache(gId int32, rd redis.Conn) (*pb.GroupInfo, error) {
	if rd == nil {
		return nil, errors.New("rd is null")
	}
	defer rd.Close()
	key := pub.MakeGroupDataKey(int32(gId))

	g, err := redis.String(rd.Do("GET", key))
	gInfo := &pb.GroupInfo{}
	err = json.Unmarshal([]byte(g), gInfo)

	logger.Debugf("get group info: %+v", gInfo)
	if err != nil {
		logger.Debugf("json parse user data(%s) error: %s\n", string(g), err)
		return nil, err
	}
	// 获取每个群组中的userList
	var userMap = make(map[int32]*pb.DeviceInfo)
	gInfo.UserList, err = user_cache.GetGroupMemDataFromCache(gId, userMap, rd)
	if err != nil {
		logger.Debugf("get user from group (%s) error: %s\n", string(gId), err)
		return nil, err
	}
	return gInfo, nil
}

func DelGroupAll(gId int32) error {
	var (
		rd  redis.Conn
		err error
	)

	if rd = cache.GetRedisClient(); rd == nil {
		return errors.New("can't get redis conn")
	}
	defer rd.Close()

	_, err = rd.Do("DEL", pub.MakeGroupMemKey(gId), pub.MakeGroupDataKey(gId))
	if err != nil {
		logger.Error("DelGroupAll with error : ", err)
		return errors.New("can't remove redis conn")
	}
	return nil
}

// 更新群组的信息
func AddSingleGroupInCache(gi *pb.GroupInfo, rd redis.Conn) error {
	if rd == nil {
		return errors.New("rd is null")
	}
	defer rd.Close()

	logger.Debugf("Will add single group data: %+v into cache", gi)
	grpData, err := json.Marshal(gi)
	if err != nil {
		logger.Debugf("json marshal error: %s\n", err)
		return err
	}
	grpKey := pub.MakeGroupDataKey(gi.Gid)

	if _, err = rd.Do("SET", grpKey, grpData); err != nil {
		logger.Debugf("Add group data to cache error: %s\n", err)
		return err
	}

	return nil
}

func GetGroupData(gId int32, rd redis.Conn) (*pb.GroupInfo, error) {
	if rd == nil {
		return nil, errors.New("rd is null")
	}
	defer rd.Close()
	key := pub.MakeGroupDataKey(int32(gId))

	g, err := redis.String(rd.Do("GET", key))
	logger.Debugf("get group info: %+v", g)
	gInfo := &pb.GroupInfo{}
	err = json.Unmarshal([]byte(g), gInfo)
	if err != nil {
		logger.Debugf("json parse user data(%s) error: %+v", string(g), err)
		return nil, err
	}
	logger.Debugf("Get group data from cache", gInfo)

	return gInfo, nil
}

func UpdateGroupData(groupInfo *pb.GroupInfo, rd redis.Conn) error {
	if rd == nil {
		return errors.New("rd is null")
	}
	defer rd.Close()
	key := pub.MakeGroupDataKey(int32(groupInfo.Gid))

	g, err := redis.String(rd.Do("GET", key))
	logger.Debugf("get group info: %+v", g)
	gInfo := &pb.GroupInfo{}
	err = json.Unmarshal([]byte(g), gInfo)
	logger.Debugf("Get group data from cache", gInfo)
	if err != nil {
		logger.Debugf("json parse user data(%s) error: %s\n", string(g), err)
		return err
	}

	gInfo.GroupName = groupInfo.GroupName
	grpData, err := json.Marshal(gInfo)
	if err != nil {
		logger.Debugf("json marshal error: %s\n", err)
		return err
	}
	_, _ = rd.Do("SET", key, grpData)

	return nil
}
