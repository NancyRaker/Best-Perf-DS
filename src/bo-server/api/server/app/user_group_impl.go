/*
@Time : 2019/5/6 13:25
@Author : yanKoo
@File : user_group_impl
@Software: GoLand
@Description:实现了群组相关的rpc调用
*/
package app

import (
	pb "bo-server/api/proto"
	"bo-server/api/server/im"
	"bo-server/dao/group"
	tg "bo-server/dao/group"        // table group
	tgc "bo-server/dao/group_cache" // table group cache
	"bo-server/dao/pub"
	tu "bo-server/dao/user"
	tuc "bo-server/dao/user_cache"
	"bo-server/engine/cache"
	"bo-server/engine/db"
	"bo-server/internal/event_hub"
	"bo-server/internal/mq/mq_sender"
	"bo-server/logger"
	"bo-server/model"
	"bo-server/utils"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
)

// 获取群组列表
func (serv *TalkCloudServiceImpl) GetGroupList(ctx context.Context, req *pb.GrpListReq) (*pb.GroupListRsp, error) {
	logger.Debugln("Get GroupList start")
	var (
		errMap   = &sync.Map{}
		wg       sync.WaitGroup
		res      = &pb.GroupListRsp{Uid: req.Uid, Res: &pb.Result{}}
		gList    = make(chan *pb.GroupListRsp, 1)
		existErr bool
		err      error
	)
	defer func() {
		close(gList)
	}()
	im.GetGroupList(int32(req.Uid), gList, errMap, &wg)
	errMap.Range(func(k, v interface{}) bool {
		err := v.(error)
		if err != nil {
			logger.Error(k, " gen error: ", err)
			existErr = true
			return false
		}
		return true
	})
	if existErr {
		err = errors.New("Internal error ")
		res.Res.Code = http.StatusInternalServerError
		return res, err
	}

	res.GroupList = (<-gList).GroupList
	res.Res.Code = http.StatusOK
	return res, nil
}

// 获取群组成员信息
func (serv *TalkCloudServiceImpl) GetGroupInfo(ctx context.Context, req *pb.GetGroupInfoReq) (*pb.GetGroupInfoResp, error) {
	// 直接去缓存获取了 TODO
	logger.Debugf("TalkCloudServiceImpl GetGroupInfo req: %+v", req)
	res, err := tgc.GetGroupListInfos([]int64{int64(req.Gid)}, cache.GetRedisClient())
	if err != nil {
		logger.Error("GetGroupInfo has error:", err)
		return &pb.GetGroupInfoResp{Res: &pb.Result{Msg: "Get group info unsuccessful, please try again later", Code: 500}}, err
	}

	resp :=  &pb.GetGroupInfoResp{
		Res: &pb.Result{
			Msg:  "Get group info successful",
			Code: http.StatusOK,
		},
	}
	if res != nil && len(res) != 0 {
		resp.GroupInfo = res[0]
	}
	return resp, nil
}

// 设置用户所在默认锁定组 (切换房间) TODO 对每一个用户设置操作都做鉴权
func (tcs *TalkCloudServiceImpl) SetLockGroupId(ctx context.Context, req *pb.SetLockGroupIdReq) (*pb.SetLockGroupIdResp, error) {
	logger.Debugf("%d Set LockGroup: %d with params: %+v", req.UId, req.GId, req)
	if !utils.CheckId(int(req.UId)) || !utils.CheckId(int(req.GId))  {
		err := errors.New("uid or gid is not valid")
		logger.Error("service SetLockGroupId error :", err)
		return &pb.SetLockGroupIdResp{Res: &pb.Result{Msg: "User id or group id is not valid, please try again later.", Code: 500}}, nil
	}

	var (
		deviceInfo *pb.DeviceInfo
		err        error
	)
	// TODO 用户id是否在组所传id中
	if deviceInfo, err = tuc.GetUserFromCache(req.UId); err != nil {

	} else {
		// 判断用户默认群组是否和之前的相同
		if deviceInfo.LockGroupId == req.GId {
			return &pb.SetLockGroupIdResp{Res: &pb.Result{Msg: "Set lock default group error, please try again later.", Code: 422}}, nil
		}
	}

	// 更新数据库
	if err := tu.SetLockGroupId(req, db.DBHandler); err != nil {
		logger.Error("service SetLockGroupId error :", err)
		return &pb.SetLockGroupIdResp{Res: &pb.Result{Msg: "Set lock default group error, please try again later.", Code: 500}}, nil
	}

	// 更新缓存
	if err := tuc.UpdateUserInfoInCache(req.UId, pub.LOCK_GID, req.GId, cache.GetRedisClient()); err != nil {
		logger.Error("service SetLockGroupId error :", err)
		// TODO 去把数据库里的群组恢复？
		return &pb.SetLockGroupIdResp{Res: &pb.Result{Msg: "Set lock default group error, please try again later.", Code: 500}}, nil
	}

	// 通知web
	go im.SendSingleNotify([]int32{deviceInfo.AccountId}, req.UId, &pb.StreamResponse{
		Uid:                    deviceInfo.AccountId,
		DataType:               im.APP_JANUS_NOTIFY,
		JanusChangeGroupNotify: &pb.JanusChangeGroupNotify{NewGroupId: req.GId, OldGroupId: deviceInfo.LockGroupId},
	})
	return &pb.SetLockGroupIdResp{Res: &pb.Result{Msg: "Set lock default group success", Code: 200}}, nil
}

// 删除组
func (serv *TalkCloudServiceImpl) RemoveGrp(ctx context.Context, req *pb.GroupDelReq) (*pb.GroupDelRsp, error) {
	var comm = im.Common{}
	return comm.RemoveGroup(ctx, req, im.APP_REMOVE_GROUP_NOTIFY)
}

// 邀请他人进组
func (serv *TalkCloudServiceImpl) InviteUserIntoGroup(ctx context.Context, req *pb.InviteUserReq) (*pb.InviteUserResp, error) {
	uIdStrs := strings.Split(req.Uids, ",")
	uIds := make([]int32, 0)
	resp := &pb.InviteUserResp{
		Res: &pb.Result{
			Msg:  "Invite user into group unsuccessful, please try again later",
			Code: http.StatusInternalServerError,
		},
	}
	for _, v := range uIdStrs {
		uId, err := strconv.Atoi(v)
		if err != nil {
			logger.Debugf("Invite user into group range uIdStrs have error: %v", err)
			return resp, nil
		}
		uIds = append(uIds, int32(uId))
	}

	logger.Debugf("%v", uIds)
	for _, v := range uIds {
		err := group.AddGroupMember(v, req.Gid, group.GROUP_MEMBER, db.DBHandler)
		if err != nil {
			logger.Debugf("Invite user into group range uIds have error: %v", err)
			return resp, nil
		}
	}
	// 添加进缓存
	// 1. 更新用户的group那个set
	if err := tuc.AddUsersGroupInCache(uIds, req.Gid, cache.GetRedisClient()); err != nil {
		logger.Error("Invite user into group AddUsersGroupInCache error: ", err)
		return resp, nil
	}
	// 2. 更新群组里有哪些用户那个set AddGroupSingleMemCache
	if err := tgc.AddGroupMemsInCache(req.Gid, uIds, cache.GetRedisClient()); err != nil {
		logger.Error("Invite user into group AddGroupMemsInCache error: ", err)
		return resp, nil
	}

	resp.Res.Code = http.StatusOK
	resp.Res.Msg = "Invite user into group successful"
	var userMap = make(map[int32]*pb.DeviceInfo)
	gMem, err := tuc.GetGroupMemDataFromCache(req.Gid, userMap, cache.GetRedisClient())
	if err != nil {
		logger.Error("Invite user into group GetGroupMemDataFromCache error: ", err)
		return resp, nil

	}
	resp.UserList = gMem
	resp.Res.Msg = "Invite user into group successful"
	resp.Res.Code = http.StatusOK
	return resp, err

}

// 主动join某个群组
func (serv *TalkCloudServiceImpl) JoinGroup(ctx context.Context, req *pb.GrpUserAddReq) (*pb.GrpUserAddRsp, error) {
	// 如果已经在群组里，就直接返回
	resp := &pb.GrpUserAddRsp{Res: &pb.Result{Msg: "Join group unsuccessful, please try again later", Code: 500}}
	_, gMap, err := group.GetGroupListFromDB(req.Uid, db.DBHandler)
	if err != nil {
		logger.Debugf("JoinGroup GetGroupListFromDB error: %+v", err)
		return resp, err
	}
	if _, ok := (*gMap)[req.Gid]; ok {
		logger.Error("User join this group already")
		return resp, err
	}

	// TODO 判断要id是不是有没有权限加群?,比如只有当前调度员名下的设备才可以加调度员名下的群组
	err = group.AddGroupMember(req.Uid, req.Gid, group.GROUP_MEMBER, db.DBHandler)
	if err != nil {
		logger.Debugf("JoinGroup AddGroupMember error: %+v", err)
		return resp, err
	}
	// 添加进缓存
	// 1. 更新用户的group那个set
	if err := tuc.AddUserForSingleGroupCache(req.Uid, req.Gid, cache.GetRedisClient()); err != nil {
		logger.Error("JoinGroup AddUserForSingleGroupCache error: ", err)
		return resp, err
	}
	// 2. 更新群组里有哪些用户那个set AddGroupSingleMemCache
	if err := tgc.AddGroupSingleMemCache(req.Gid, req.Uid, cache.GetRedisClient()); err != nil {
		logger.Error("JoinGroup AddGroupSingleMemCache error: ", err)
		return resp, err
	}
	// 3. 添加这个群组的信息进缓存，因为这个是模糊搜索的结果
	gInfo, err := tg.GetGroupInfoFromDB(req.Gid, req.Uid)
	if err != nil {
		logger.Error("JoinGroup GetGroupInfoFromDB error: ", err)
		return resp, err
	}

	//3.1 每个用户的信息
	for _, u := range gInfo.UserList {
		if err := tuc.AddUserDataInCache(u.Id, []interface{}{
			pub.USER_Id, u.Id,
			pub.IMEI, u.Imei,
			pub.USER_NAME, u.NickName,
			pub.ONLINE, u.Online,
			pub.LOCK_GID, u.LockGroupId,
		}, cache.GetRedisClient()); err != nil {
			logger.Error("Add user information to cache with error: ", err)
		}
	}

	//3.2  每一个群组拥有的成员
	if err := tgc.AddGroupCache(gInfo.UserList, gInfo, cache.GetRedisClient()); err != nil {
		logger.Error("JoinGroup AddGroupCache error: ", err)
		return resp, err
	}

	return &pb.GrpUserAddRsp{Res: &pb.Result{Msg: "Join group successful", Code: 200}}, err
}

// 通过关键字返回群组，区分在群组和不在的群组
func (serv *TalkCloudServiceImpl) SearchGroup(ctx context.Context, req *pb.GrpSearchReq) (*pb.GroupListRsp, error) {
	// 判空
	if req.Target == "" {
		return &pb.GroupListRsp{Res: &pb.Result{Code: 422, Msg: "process error, please input target"}}, errors.New("target is nil")
	}

	// 模糊查询群组 TODO 暂时这么写吧，感觉有点蠢
	groups, err := group.SearchGroup(req.Target, db.DBHandler)
	if err != nil {
		return &pb.GroupListRsp{Res: &pb.Result{Code: 500, Msg: "process error, please try again"}}, err
	}
	// 查找用户所在组
	_, gMap, err := group.GetGroupListFromDB(req.Uid, db.DBHandler)
	if err != nil {
		return &pb.GroupListRsp{Res: &pb.Result{Code: 500, Msg: "process error, please try again"}}, err
	}

	for _, v := range groups.GroupList {
		if _, ok := (*gMap)[v.Gid]; ok {
			v.IsExist = true
		}
	}
	logger.Debugf("server search group: %+v", groups)
	groups.Res = &pb.Result{Msg: "search group success", Code: 200}
	return groups, nil
}

// 移除群中某个成员
func (serv *TalkCloudServiceImpl) RemoveGrpUser(ctx context.Context, req *pb.GrpUserDelReq) (*pb.GrpUserDelRsp, error) {
	logger.Debugf("uid: %d, gid:%d", req.Uid, req.Gid)
	err := group.RemoveGroupMember(req.Uid, req.Gid, db.DBHandler)
	resp := &pb.GrpUserDelRsp{
		Res: &pb.Result{
			Msg:  "remove Group User error, please try again later.",
			Code: http.StatusInternalServerError,
		},
	}
	if err != nil {
		logger.Error("Remove Group error: ", err)
		return resp, nil
	}

	// 清空缓存
	// 1. 更新该用户在哪些组的那个set
	if err := tuc.RemoveUserForSingleGroupCache(req.Uid, req.Gid, cache.GetRedisClient()); err != nil {
		logger.Error("JoinGroup AddUserForSingleGroupCache error: ", err)
	}
	// 2. 更新群组里有哪些用户那个set AddGroupSingleMemCache
	if err := tgc.RemoveGroupSingleMemCache(req.Gid, req.Uid, cache.GetRedisClient()); err != nil {
		logger.Error("JoinGroup AddUserForSingleGroupCache error: ", err)
	}
	resp.Res.Code = http.StatusOK
	resp.Res.Msg = "remove Group User success."
	return resp, err
}

// 退出群组
func (serv *TalkCloudServiceImpl) ExitGrp(ctx context.Context, req *pb.UserExitReq) (*pb.UserExitRsp, error) {
	return nil, nil
}

// app 创建群组
func (serv *TalkCloudServiceImpl) CreateGroup(ctx context.Context, req *pb.CreateGroupReq) (*pb.GroupInfo, error) {
	logger.Debugf("App create group is name: %s. the GroupInfos: %+v", req.GroupInfo.GroupName, req)
	groupReq := &pb.WebCreateGroupReq{DeviceIds: req.DeviceIds, GroupInfo: req.GroupInfo}
	if groupReq == nil || groupReq.GroupInfo == nil || groupReq.GroupInfo.AccountId == 0 {
		return nil, errors.New("param is invalid")
	}

	// 创建组
	if gid, err := tg.WebCreateGroup(groupReq, tg.CREATE_GROUP_BY_USER); err != nil {
		logger.Error("App create group error :", err)
		return nil, err
	} else {
		groupReq.GroupInfo.Id = int32(gid)
	}

	// TODO GRPC来通知janus创建房间
	if err := notifyJanusCreateRoom(groupReq); err != nil {
		return nil, err
	}

	// 群组信息和群组成员id增加到缓存
	if err := tgc.WebAddGroupAndUserInCache(groupReq, cache.GetRedisClient()); err != nil {
		logger.Debugf("App create AddGroupAndUserInCache error: %deviceId", err)
	}

	// 增加所创建群所含成员也要加进缓存,因为每个成员都新加了一个群组,还要把每个人的信息也加入缓存
	for _, deviceId := range groupReq.DeviceIds {
		if err := tgc.AddGroupSingleMemCache(int32(groupReq.GroupInfo.Id), deviceId, cache.GetRedisClient()); err != nil {
			logger.Debugf("App create AddGroupAndUserInCache error: %deviceId", err)
		}

		if err := tuc.AddUserForSingleGroupCache(deviceId, int32(groupReq.GroupInfo.Id), cache.GetRedisClient()); err != nil {
			logger.Error("App create add group member into single group into cache error:", err)
		}
	}

	groupReq.DeviceIds = append(groupReq.DeviceIds, req.GroupInfo.AccountId)
	// 通知app进入房间 TODO 有问题，如果很多goroutine都来调用这个接口，这个全局channel可能会出现丢数据的情况，塞满之后，其他goroutine全挂掉，不过200个消费协程，这种情况应该比较少。
	im.SendSingleNotify(groupReq.DeviceIds, req.GroupInfo.AccountId, &pb.StreamResponse{
		Uid:                    req.GroupInfo.AccountId,
		DataType:               im.APP_CREATE_GROUP_NOTIFY,
		Notify: &pb.LoginOrLogoutNotify{GroupInfo: &pb.GroupInfo{Gid:groupReq.GroupInfo.Id, GroupName:"temp call group"/*groupReq.GroupInfo.GroupName*/}},
	})

	// app创建群组的时候，自己加进缓存
	if err := tgc.AddGroupSingleMemCache(int32(groupReq.GroupInfo.Id), int32(groupReq.GroupInfo.AccountId), cache.GetRedisClient()); err != nil {
		logger.Debugf("App create AddGroupAndUserInCache error: %deviceId", err)
	}
	if err := tuc.AddUserForSingleGroupCache(int32(groupReq.GroupInfo.AccountId), int32(groupReq.GroupInfo.Id), cache.GetRedisClient()); err != nil {
		logger.Error("App create add group member into single group into cache error:", err)
	}


	return &pb.GroupInfo{Gid: int32(groupReq.GroupInfo.Id), Status: groupReq.GroupInfo.Status}, nil
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
		Timeout: time.Millisecond * 1000, // 暂时设置500ms超时
	})
	if !janusRoomIsOk {
		// TODO 失败就删除房间
		_ = tg.DeleteGroup(&model.GroupInfo{Id: int(groupReq.GroupInfo.Id)})
		return errors.New("janus create Group is fail or timeout")
	}

	return nil
}
