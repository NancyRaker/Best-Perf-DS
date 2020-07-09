/**
 * Copyright (c) 2019. All rights reserved.
 * some functions deal with cache data of the groups
 * Author: tesion
 * Date: March 29th 2019
 */
package service

import (
	"encoding/json"
	"errors"
	"github.com/gomodule/redigo/redis"
	"strconv"
	"web-api/dao/pub"
	"web-api/logger"
	"web-api/model"
	pb "web-api/proto/talk_cloud"
	"web-api/utils"
)

func GetGroupListInfos(accountId int, gIds []int32, rd redis.Conn) ([]*model.GroupListNode, map[int32]bool, error) {
	if rd == nil {
		return nil, nil, errors.New("redis conn is nil")
	}
	defer rd.Close()

	// 1.0 获取所有的缓存中所有的群组信息
	var (
		groupList []*model.GroupListNode
		gDataKeys = make([]interface{}, 0)
		gMemKeys  = make([]interface{}, 0)
		statMap   = make(map[int32]bool)
		err       error
	)

	for _, v := range gIds {
		gDataKeys = append(gDataKeys, pub.MakeGroupDataKey(v))
		gMemKeys = append(gMemKeys, pub.MakeGroupMemKey(v))
	}
	groups, err := redis.Strings(rd.Do("MGET", gDataKeys...))

	// 2.1 去重, 找出需要查询的用户
	var (
		groupIdMaps = make(map[int32][]int)
		uIdsMap     = make(map[int32]bool)
		totalUId    []int32
		userInfoMap = make(map[int32]*model.User)
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
	//log.Log.Debugf("from redis groupIdMaps: %+v", groupIdMaps)

	// 2.3 计算所有群组存在过的设备
	for id, ok := range uIdsMap {
		if ok {
			totalUId = append(totalUId, id)
		}
	}

	logger.Debugf("from redis groups: %+v", groups)
	logger.Debugf("from redis totalUId: %+v", totalUId)

	// 3.0 批量查询uid的信息和在线状态
	for _, uId := range totalUId {
		_ = rd.Send("HMGET", pub.MakeUserDataKey(int32(uId)),
			pub.USER_Id, pub.IMEI, pub.NICK_NAME, pub.LOCK_GID,
			pub.USER_NAME, pub.USER_TYPE, pub.ACCOUNT_ID, pub.DEVICE_TYPE)
	}
	_ = rd.Flush()
	for _, uId := range totalUId {
		source, err := redis.Strings(rd.Receive())
		user := &model.User{}
		if source != nil {
			id, _ := strconv.Atoi(source[0])
			user.Id = id
			user.IMei = source[1]
			user.NickName = source[2]
			lockGId, _ := strconv.Atoi(source[3])
			user.LockGroupId = lockGId
			user.Online = pub.USER_OFFLINE // 2是在线
			//if devicesMap[id] == nil { // 因为缓存里可能有非设备用户在群组里，而这个用户又不在调度员名下，这样就会报空针
			//	//continue  // 会出现手机用户的群组出现在调度员名下
			//} else {
				user.UserName = source[4] //devicesMap[id].UserName
				userType, _ := strconv.Atoi(source[5])
				user.UserType = userType // devicesMap[id].UserType
				accountId, _ := strconv.Atoi(source[6])
				user.AccountId = accountId// devicesMap[id].AccountId
				user.DeviceType = source[7] // devicesMap[id].DeviceType
			//}

		}
		userInfoMap[uId] = user
		logger.Debugf("from redis # %d users source: %+v with err: %+v", uId, source, err)
	}

	var bad = make([]int32, 0)
	// 3.1 uid在线状态
	for _, uId := range totalUId {
		//if devicesMap[int(uId)] == nil {
		//	bad = append(bad, uId)
		//}
		_ = rd.Send("GET", pub.GetRedisKey(int32(uId)))
	}
	_ = rd.Flush()
	for _, id := range totalUId {
		source, err := redis.Bytes(rd.Receive())
		s := &model.SessionInfo{}
		_ = json.Unmarshal(source, s)
		// TODO 在线状态
		if err != nil {
		} else {
			statMap[int32(id)] = true
		}
	}
	logger.Debugf("current map :%+v", statMap)
	logger.Debugf("bad id :%+v", bad)

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
			return nil, nil, err
		}
		var groupListNode = &model.GroupListNode{}
		groupListNode.GroupInfo = &model.GroupInfo{
			Id:        int(gInfo.Gid),
			GroupName: gInfo.GroupName,
			AccountId: accountId,
			//Status:"",
			//CTime:
			GroupManager: gInfo.GroupManager,
			GroupOwner:   gInfo.GroupOwner,
		}

		var (
			head        = make([]*model.User, 0) // 只用来存放调度员节点
			offlineList = make([]*model.User, 0)
			onlineList  = make([]*model.User, 0)
			OnlineNum   = 0
			janusNum    = 0
			GroupTotal  = 0
		)
		for _, id := range groupIdMaps[gInfo.Gid] {
			GroupTotal++
			user := userInfoMap[int32(id)]
			if key, ok := statMap[int32(id)]; ok && key {
				OnlineNum++
				user.Online = pub.USER_ONLINE
				if id == accountId {
					OnlineNum--  // 调度员 let web deal
					head = append(head, user)
					continue
				}
				if user.LockGroupId == int(gInfo.Gid) {
					janusNum++
					newUser := &model.User{}
					_ = utils.DeepCopy(newUser, user)
					newUser.Online = pub.USER_JANUS_ONLINE
					groupListNode.DeviceInfo = append(groupListNode.DeviceInfo, newUser)
				} else {
					onlineList = append(onlineList, user)
				}
			} else {
				if id == accountId {
					head = append(head, user)
					continue
				}
				offlineList = append(offlineList, user)
			}
		}

		// 根据群主、管理员、普通成员排序
		sortByMemberType(groupListNode.GroupInfo, &groupListNode.DeviceInfo)
		sortByMemberType(groupListNode.GroupInfo, &offlineList)
		sortByMemberType(groupListNode.GroupInfo, &onlineList)
		sortByMemberType(groupListNode.GroupInfo, &head)

		groupListNode.DeviceInfo = append(groupListNode.DeviceInfo, onlineList...)
		groupListNode.DeviceInfo = append(groupListNode.DeviceInfo, offlineList...)
		head = append(head, groupListNode.DeviceInfo...)

		groupListNode.DeviceInfo = nil
		groupListNode.DeviceInfo = append(groupListNode.DeviceInfo, head...)

		groupListNode.GroupInfo.OnlineNum = OnlineNum
		groupListNode.GroupInfo.JanusNum = janusNum

		groupList = append(groupList, groupListNode)
	}
	return groupList, statMap, nil
}

// 根据群主、管理员、普通成员排序
func sortByMemberType(groupInfo *model.GroupInfo, users *[]*model.User) {
	var (
		i, j, k    = 0, 0, len(*users)-1 // 管理员：[i, j)
		managerMap = make(map[int32]bool)
	)
	for _, id := range groupInfo.GroupManager {
		managerMap[id] = true
	}

	for j <= k {
		if int32((*users)[j].Id) == groupInfo.GroupOwner {
			(*users)[j].GroupType = pub.GROUP_OWNER
			swap(users, i, j)
			i++
			j++
		} else if managerMap[int32((*users)[j].Id)] { // 如果是管理员
			(*users)[j].GroupType = pub.GROUP_MANAGER
			j++
		} else { // 普通成员
			(*users)[j].GroupType = pub.GROUP_MEMBER
			swap(users, j, k)
			k--
		}
	}
}

func swap(users *[]*model.User, i, j int) {
	(*users)[i], (*users)[j] = (*users)[j], (*users)[i]
}
