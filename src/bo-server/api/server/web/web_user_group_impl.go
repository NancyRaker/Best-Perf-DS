/*
@Time : 2019/3/28 15:33
@Author : yanKoo
@File : TalkCloudRegisterImpl
@Software: GoLand
@Description: 目前主要供web端调用 protoc -I . talk_cloud_web.proto --go_out=plugins=grpc:.
*/
package web

import (
	pb "bo-server/api/proto"
	"bo-server/api/server/im"
	tg "bo-server/dao/group"        // table group
	tgc "bo-server/dao/group_cache" // table group cache
	tgm "bo-server/dao/group_member"
	tuc "bo-server/dao/user_cache"
	"bo-server/engine/cache"
	"bo-server/engine/db"
	"bo-server/internal/event_hub"
	"bo-server/internal/mq/mq_sender"
	"bo-server/logger"
	"bo-server/model"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"time"
)

type WebServiceServerImpl struct {
}

// web创建群组
func (wssu *WebServiceServerImpl) WebCreateGroup(ctx context.Context, groupReq *pb.WebCreateGroupReq) (*pb.GroupInfo, error) {
	logger.Debugf("Create group is name: %s. the GroupInfos: %+v", groupReq.GroupInfo.GroupName, groupReq)
	if groupReq == nil || groupReq.GroupInfo == nil || groupReq.GroupInfo.AccountId == 0 { //|| len(groupReq.DeviceIds) == 0 {
		return nil, errors.New("param is invalid")
	}
	if groupReq.GroupInfo.Status == tg.TEMP_GROUP {
		// 判断该调度员是否有创建临时组
		var err error
		groupReq.GroupInfo.Id = 0
		if groupReq.GroupInfo.Id, err = tg.CheckDuplicateCreateGroup(groupReq.GroupInfo.AccountId); err != nil {
			return nil, errors.New("CheckDuplicateCreateGroup is has error: " + err.Error())
		} else {
			if groupReq.GroupInfo.Id != 0 {
				// 更新群成员
				var (
					addMemberId    = make([]int32, 0) // 增加设备的id
					removeMemberId = make([]int32, 0) // 减少设备的id
				)
				// 1.获取该群组下有多少人
				if olders, err := tgc.GetGroupMem(groupReq.GroupInfo.Id, cache.GetRedisClient()); err != nil {
					return nil, errors.New("GetGroupMem is has error: " + err.Error())
				} else {
					//addMemberId, removeMemberId = splitGroupMember(olders, groupReq.DeviceIds, groupReq.GroupInfo.AccountId)
					addMemberId, removeMemberId = modifyGroupMember(olders, groupReq.DeviceIds, int64(groupReq.GroupInfo.AccountId))
				}
				// 2. 更新群成员
				// 2.1 找出删除了哪些人，增加了哪些人，调用更新成员接口
				if _, err := wssu.UpdateGroup(ctx, &pb.UpdateGroupReq{
					GroupInfo:         groupReq.GroupInfo,
					AddDeviceInfos:    addMemberId,
					RemoveDeviceInfos: removeMemberId}); err != nil {
					return nil, errors.New("wssu.UpdateGroup is has error: " + err.Error())
				}
				// 3. 返回
				return &pb.GroupInfo{Gid: groupReq.GroupInfo.Id, Status: groupReq.GroupInfo.Status}, nil
			}
		}
	}

	// 第一次创建临时组
	if gid, err := tg.WebCreateGroup(groupReq, tg.CREATE_GROUP_BY_DISPATCHER); err != nil {
		logger.Error("create group error :", err)
		return nil, err
	} else {
		groupReq.GroupInfo.Id = int32(gid)
	}

	// TODO grpc来通知janus创建房间
	if err := notifyJanusCreateRoom(groupReq); err != nil {
		return nil, err
	}

	// 群组信息和群组成员id增加到缓存
	if err := tgc.WebAddGroupAndUserInCache(groupReq, cache.GetRedisClient()); err != nil {
		logger.Debugf("CreateGroup AddGroupAndUserInCache error: %v", err)
	}

	// 增加所创建群所含成员也要加进缓存,因为每个成员都新加了一个群组,还要把每个人的信息也加入缓存
	for _, v := range groupReq.DeviceIds {
		if err := tgc.AddGroupSingleMemCache(int32(groupReq.GroupInfo.Id), v, cache.GetRedisClient()); err != nil {
			logger.Debugf("CreateGroup AddGroupAndUserInCache error: %v", err)
		}

		if err := tuc.AddUserForSingleGroupCache(v, int32(groupReq.GroupInfo.Id), cache.GetRedisClient()); err != nil {
			logger.Error("CreateGroup add group member into single group into cache error:", err)
		}

		u := &pb.DeviceInfo{}
		if v != groupReq.GroupInfo.AccountId { // TODO 有问题， 为什么要写这一步，暂时放着
			tuc.UpdateUserFromDBToRedis(u, int(v))
		}

		if groupReq.GroupInfo.Status == tg.TEMP_GROUP {
			_, _ = im.ServerWorker{NotifyType: im.TEMP_GROUP_CREATE_NOTIFY}.WebJanusTempGroupImPublish(context.TODO(), v, groupReq.GroupInfo.Id)
		} else { // im.ServerWorker{}.WebJanusImPublish
			_, _ = im.ServerWorker{NotifyType: im.WEB_JANUS_NOTIFY}.WebJanusImPublish(context.TODO(), v, groupReq.GroupInfo.Id)
		}
	}

	// web创建群组的时候，自己加进缓存
	if err := tgc.AddGroupSingleMemCache(int32(groupReq.GroupInfo.Id), int32(groupReq.GroupInfo.AccountId), cache.GetRedisClient()); err != nil {
		logger.Debugf("CreateGroup AddGroupAndUserInCache error: %v", err)
	}
	if err := tuc.AddUserForSingleGroupCache(int32(groupReq.GroupInfo.AccountId), int32(groupReq.GroupInfo.Id), cache.GetRedisClient()); err != nil {
		logger.Error("CreateGroup add group member into single group into cache error:", err)
	}

	return &pb.GroupInfo{Gid: int32(groupReq.GroupInfo.Id), Status: groupReq.GroupInfo.Status}, nil
}

// 把oldMembers中除了accountId的全部删除
// newMembers正常情况下不存在accountId
func modifyGroupMember(oldMembers []int64, newMembers []int32, accountId int64) ([]int32, []int32) {
	var removeMembers []int32
	for _, m := range oldMembers {
		if m != accountId {
			removeMembers = append(removeMembers, int32(m))
		}
	}

	return newMembers, removeMembers
}

func notifyJanusCreateRoom(groupReq *pb.WebCreateGroupReq) error {
	// TODO 往消息队列里塞数据发送给janus创建房间
	msgCreateGroup := &model.GrpcNotifyJanus{
		MsgType: mq_sender.SignalType,
		Ip:      "127.0.0.1",
		SignalMsg: &model.NotifyJanus{
			SignalType: mq_sender.CreateGroupReq,
			CreateGroupReq: &model.CreateGroupReq{
				GId:          int(groupReq.GroupInfo.Id),
				GroupName:    groupReq.GroupInfo.GroupName,
				DispatcherId: int(groupReq.GroupInfo.AccountId),
			},
		},
	}
	if msg, err := json.Marshal(msgCreateGroup); err != nil {
		return errors.New("janus create Group msg json marshal is fail")
	} else {
		mq_sender.SendMsg(&model.MsgObject{Msg: msg})
	}

	// TODO 阻塞一段时间，等待获取到eventhub中创建房间又返回
	janusRoomIsOk := event_hub.QueryEvent(event_hub.CREATE_GROUP_RESP, event_hub.QueryParams{
		GroupId: int(groupReq.GroupInfo.Id),
		Timeout: time.Millisecond * 500, // 暂时设置500ms超时
	})
	if !janusRoomIsOk {
		// TODO 失败就删除房间
		_ = tg.DeleteGroup(&model.GroupInfo{Id: int(groupReq.GroupInfo.Id)})
		return errors.New("janus create Group is fail or timeout")
	}

	return nil
}

// add: newer中有而older中没有
// no: newer中有older中也有
// rem: newer中没有而older中有
func splitGroupMember(olders []int64, newers []int32, accountId int32) (addMem []int32, revMem []int32) {
	var (
		olderMap = make(map[int32]bool)
		newerMap = make(map[int32]bool)
	)

	logger.Debugf("older: %+v", olders)
	logger.Debugf("newers: %+v", newers)
	logger.Debugf("dispatcher: %+d", accountId)

	for _, v := range newers {
		newerMap[int32(v)] = true
	}

	for _, v := range olders {
		id := int32(v)
		if !newerMap[id] && id != accountId { // 调度员不能移除
			revMem = append(revMem, id)
		}
		olderMap[id] = true
	}

	for id := range newerMap {
		if !olderMap[id] {
			addMem = append(addMem, id)
		}
	}

	logger.Debugf("add:%+v", addMem)
	logger.Debugf("rev:%+v", revMem)
	return
}

// 更新组成员
func (wssu *WebServiceServerImpl) UpdateGroup(ctx context.Context, req *pb.UpdateGroupReq) (*pb.UpdateGroupResp, error) {
	logger.Debugf("UpdateGroup enter update group with %+v, add device:%+v, rev device:%+v", req, req.AddDeviceInfos, req.RemoveDeviceInfos)
	// TODO 校验设备是否还在调度员名下

	var addDevicesNotify, removeDevicesNotify []int32
	// 更新群成员
	if req.GroupInfo != nil && req.GroupInfo.Status == tg.TEMP_GROUP {
		var removeDevices []int32
		// 1.获取该群组下有多少人
		if olders, err := tgc.GetGroupMem(req.GroupInfo.Id, cache.GetRedisClient()); err != nil {
			return nil, errors.New("GetGroupMem is has error: " + err.Error())
		} else {
			// 临时组每个人都要通知到
			addDevicesNotify, removeDevicesNotify = modifyNotifyMember(modifyGroupMember(olders, req.AddDeviceInfos, int64(req.GroupInfo.AccountId)))

			if req.AddDeviceInfos != nil { // web判断加入多少设备
				req.AddDeviceInfos, removeDevices = splitGroupMember(olders, req.AddDeviceInfos, req.GroupInfo.AccountId)
				req.RemoveDeviceInfos = append(req.RemoveDeviceInfos, removeDevices...)
			}
			// 如果removeNotify里面，addDeviceNotify也有，就不通知以及被移除了
			removeDevicesNotify = req.RemoveDeviceInfos
			logger.Debugf("UpdateGroup exec update group with add device:%+v, rev device:%+v， add notify: %+v, rem notify: %+v",
				req.AddDeviceInfos, req.RemoveDeviceInfos, addDevicesNotify, removeDevicesNotify)

		}
	} else if req.GroupInfo != nil && req.GroupInfo.Status == tg.NORMAL_GROUP {
		addDevicesNotify, removeDevicesNotify = req.AddDeviceInfos, req.RemoveDeviceInfos
	}

	logger.Debugf("UpdateGroup exec update group with add device:%+v, rev device:%+v， add notify: %+v, rem notify: %+v",
		req.AddDeviceInfos, req.RemoveDeviceInfos, addDevicesNotify, removeDevicesNotify)
	if err := tg.UpdateGroupMember(req.RemoveDeviceInfos, req.AddDeviceInfos, req.GroupInfo.Id); err != nil {
		logger.Error("create group error :", err)
		return &pb.UpdateGroupResp{ResultMsg: &pb.Result{Msg: "Update group unsuccessful, please try again later", Code: 422}}, err
	}

	// 增加到缓存
	if err := tgc.UpdateGroupAndUserMemberInCache(req.RemoveDeviceInfos, req.AddDeviceInfos, req.GroupInfo.Id); err != nil {
		logger.Error("insert cache error")
		return &pb.UpdateGroupResp{ResultMsg: &pb.Result{Msg: "Update group unsuccessful, please try again later", Code: 422}}, err
	}

	for _, id := range addDevicesNotify {
		if req.GroupInfo.Status == tg.TEMP_GROUP {
			_, _ = im.ServerWorker{NotifyType: im.TEMP_GROUP_CREATE_NOTIFY}.WebJanusTempGroupImPublish(context.TODO(), id, req.GroupInfo.Id)
		} else {
			_, _ = im.ServerWorker{NotifyType: im.WEB_JANUS_NOTIFY}.WebJanusImPublish(context.TODO(), id, req.GroupInfo.Id)
		}
	}

	for _, id := range removeDevicesNotify {
		if req.GroupInfo.Status == tg.TEMP_GROUP {
			_, _ = im.ServerWorker{NotifyType: im.TEMP_GROUP_REM_NOTIFY}.WebJanusTempGroupImPublish(context.TODO(), id, req.GroupInfo.Id)
		} else {
			_, _ = im.ServerWorker{NotifyType: im.NORMAL_GROUP_REM_NOTIFY}.WebJanusImPublish(context.TODO(), id, req.GroupInfo.Id)
		}
	}

	return &pb.UpdateGroupResp{ResultMsg: &pb.Result{Msg: "Update group successful.", Code: 200}}, nil
}

func modifyNotifyMember(addNotify []int32, revNotify[]int32) ([]int32, []int32) {
	var addNotifyMap = make(map[int32]bool)
	for _, id := range addNotify {
		addNotifyMap[id]= true
	}

	var removeResult []int32
	for i := range revNotify {
		if !addNotifyMap[revNotify[i]] {
			removeResult = append(removeResult, revNotify[i])
		}
	}

	return addNotify, removeResult
}

// 删除组
func (wssu *WebServiceServerImpl) DeleteGroup(ctx context.Context, req *pb.GroupDelReq) (*pb.GroupDelRsp, error) {
	var comm = im.Common{}
	return comm.RemoveGroup(ctx, req, im.WEB_JANUS_NOTIFY)
}

// 更新群组管理员
func (wssu *WebServiceServerImpl) UpdateGroupManager(ctx context.Context, req *pb.UpdGManagerReq) (*pb.UpdGManagerResp, error) {
	// TODO 判断管理员人数是否超标
	var (
		gi         *pb.GroupInfo
		err        error
		managerMap = make(map[int32]bool, 0)
	)
	// 目前只允许修改群管理员
	if !((req.RoleType == tg.GROUP_MANAGER) || (req.RoleType == tg.GROUP_MEMBER)) || len(req.UId) == 0 {
		err = errors.New("param is invalid")
		logger.Debugf("Update GroupManager fail , error: %s", err)
		goto Err
	}

	// 1. 更新数据库
	if err = tgm.UpdateGroupManager(req, db.DBHandler); err != nil {
		logger.Debugf("Update GroupManager UpdateGroupManager fail , error: %s", err)
		goto Err
	}

	// 2. 更新缓存
	if gi, err = tgc.GetGroupData(req.GId, cache.GetRedisClient()); err != nil {
		logger.Debugf("Update GroupManager GetGroupData fail , error: %s", err)
		goto Err
	}

	for _, m := range gi.GroupManager {
		managerMap[m] = true
	}

	for _, m := range req.UId {
		if req.RoleType == tg.GROUP_MANAGER {
			// 是增加管理员
			managerMap[m] = true
		} else if req.RoleType == tg.GROUP_MEMBER {
			managerMap[m] = false
		}
	}

	gi.GroupManager = make([]int32, 0)
	for id, manager := range managerMap {
		if manager {
			gi.GroupManager = append(gi.GroupManager, id)
		}
	}

	// TODO 群主必须是管理员之一？
	//gi.GroupManager = append(gi.GroupManager, gi.GroupOwner)

	if err = tgc.AddSingleGroupInCache(gi, cache.GetRedisClient()); err != nil {
		goto Err
	}

	// 通知janus更新调度员
	go func() {
		msgObj := model.GrpcNotifyJanus{
			MsgType: mq_sender.SignalType,
			Ip:      "127.0.0.1",
			SignalMsg: &model.NotifyJanus{
				SignalType: mq_sender.AlterParticipantReq,
				AlterParticipantReq: &model.AlterParticipantReq{
					Gid:          int(gi.Gid),
					DispatcherId: int(gi.GroupOwner),
					Uid:          int(req.UId[0]),
					Role:         int(req.RoleType),
				},
			},
		}
		if msg, err := json.Marshal(msgObj); err != nil {
			logger.Debugln("janus Alter Participant  msg json marshal is fail")
		} else {
			//TODO 如果janus改变角色失败，应该回滚数据，实现事务操作
			mq_sender.SendMsg(&model.MsgObject{Msg: msg})
		}

		janusOpsIsOk := event_hub.QueryEvent(event_hub.ALTER_PARTICIPANT_RESP, event_hub.QueryParams{
			AlterParticipantReq: msgObj.SignalMsg.AlterParticipantReq,
			Timeout:             time.Millisecond * 500, // 暂时设置500ms超时
		})
		logger.Debugf("UpdateGroupManager janus res: %+v", janusOpsIsOk)
	}()
	return &pb.UpdGManagerResp{Res: &pb.Result{Msg: "Update GroupManager successful ", Code: 200}}, nil

Err:
	return &pb.UpdGManagerResp{Res: &pb.Result{Msg: "Update GroupManager error ", Code: 500},}, err
}

// 更新群组信息（群组名称等）
func (wssu *WebServiceServerImpl) UpdateGroupInfo(ctx context.Context, req *pb.GroupInfo) (*pb.Result, error) {
	//  校验参数
	if req.Gid <= 0 || req.GroupName == "" {
		return nil, errors.New("invalid params")
	}

	// 更新数据库
	if err := tg.UpdateGroup(req, db.DBHandler); err != nil {
		logger.Debugf("update group fail , error: %s", err)
		return nil, errors.New("db ops error: " + err.Error())
	}

	// 更新缓存
	if err := tgc.UpdateGroupData(req, cache.GetRedisClient()); err != nil {
		logger.Debugf("update group update redis fail , error: %s", err)
		return nil, errors.New("redis ops error: " + err.Error())
	}

	groupMembers, err := tgc.GetGroupMem(req.Gid, cache.GetRedisClient())
	if err != nil {
		logger.Debugf("update group get group members redis fail , error: %s", err)
		return nil, errors.New("db ops error: " + err.Error())
	}

	// 发送通知通知app更新
	for _, id := range groupMembers {
		_, _ = im.ServerWorker{NotifyType: im.WEB_JANUS_NOTIFY}.WebJanusImPublish(context.TODO(), int32(id), req.Gid)
	}

	return &pb.Result{Code: http.StatusOK}, nil
}
