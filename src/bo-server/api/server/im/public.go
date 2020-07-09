/*
@Time : 2019/7/30 14:45
@Author : yanKoo
@File : producer
@Software: GoLand
@Description:
*/
package im

import (
	pb "bo-server/api/proto"
	"bo-server/dao/group"
	tg "bo-server/dao/group"        // table group
	tgc "bo-server/dao/group_cache" // table group cache
	"bo-server/dao/pub"
	tuc "bo-server/dao/user_cache"
	"bo-server/engine/cache"
	"bo-server/engine/db"
	"bo-server/internal/mq/mq_sender"
	"bo-server/logger"
	"bo-server/model"
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"sync"
	"time"
)

type Common struct{}

// 删除组 TODO 修改删除组思路需要修改数据库
func (*Common) RemoveGroup(ctx context.Context, req *pb.GroupDelReq, sendType int32) (*pb.GroupDelRsp, error) {
	var (
		uIds []int64
		err  error
		resp *pb.GroupDelRsp
	)
	logger.Debugf("remove group start with %+v", req)
	// 判断当前id能不能删除这个群组
	groupInfo, err := tgc.GetGroupData(req.Gid, cache.GetRedisClient())
	if err != nil {
		return &pb.GroupDelRsp{Res:&pb.Result{Msg:err.Error(), Code: http.StatusInternalServerError}}, err
	}
	if groupInfo.GroupOwner != req.Uid {
		return &pb.GroupDelRsp{Res:&pb.Result{Msg:"this user can't delete this group", Code: http.StatusUnprocessableEntity}}, nil
	}

	// clear group user first
	if err = group.ClearGroupMember(req.Gid, db.DBHandler); err != nil {
		resp.Res = &pb.Result{Code: http.StatusInternalServerError, Msg: err.Error()}
	}

	// then remove group
	if err = group.RemoveGroup(req.Gid, db.DBHandler); err != nil {
		resp.Res = &pb.Result{Code: http.StatusInternalServerError, Msg: err.Error()}
	}

	// 1. 更新该用户在哪些组的那个set
	uIds, err = tgc.GetGroupMem(req.Gid, cache.GetRedisClient())
	for _, id := range uIds {
		if err = tuc.RemoveUserForSingleGroupCache(int32(id), req.Gid, cache.GetRedisClient()); err != nil {
			logger.Error("RemoveGrp RemoveUserForSingleGroupCache error: ", err)
		}
		_, _ = ServerWorker{NotifyType: sendType}.WebJanusImPublish(context.TODO(), int32(id), req.Gid)
	}

	// 2. 清除整个组
	if err = tgc.DelGroupAll(req.Gid); err != nil {
		logger.Debugln("del group from cache error: %+id", err)
	}

	// 3. 往消息队列里发送消息，通知janus删除房间
	go func() {
		if msg, err := json.Marshal(model.GrpcNotifyJanus{
			MsgType: mq_sender.SignalType,
			Ip:      "127.0.0.1",
			SignalMsg: &model.NotifyJanus{
				SignalType: mq_sender.DestroyGroupReq,
				DestroyGroupReq: &model.DestroyGroupReq{
					GId:          int(req.Gid),
					DispatcherId: int(req.Uid),
				},
			},
		}); err != nil {
			logger.Debugln("janus destroy Group msg json marshal is fail")
		} else {
			mq_sender.SendMsg(&model.MsgObject{Msg: msg, Option: &model.SendMsgOption{Timeout: 10 * time.Second}})
		}
	}()

	return &pb.GroupDelRsp{Res: &pb.Result{Code: http.StatusOK, Msg: "del group success"}}, err
}

// 获取群组列表
func GetGroupList(uid int32, gList chan *pb.GroupListRsp, errMap *sync.Map, wg *sync.WaitGroup) {
	logger.Debugf("Get group list start")
	defer logger.Debugf("Get group list done")

	var (
		groupListResp = &pb.GroupListRsp{}
		groupList     []*pb.GroupInfo
	)

	// 先去缓存取，取不出来再去mysql取
	gIds, err := tuc.GetUserIncludedInGroups(int32(uid), cache.GetRedisClient())
	if err != nil && err != sql.ErrNoRows {
		logger.Error("No find In CacheError with err: %+v", err)
		errMap.Store("GetGroupList", err)
		gList <- groupListResp
		return
	} else if err == sql.ErrNoRows {
		logger.Debugln("user is not in any group")
		gList <- groupListResp
		return
	}

	groupList, err = tgc.GetGroupListInfos(gIds, cache.GetRedisClient())
	if err == nil {
		logger.Debugln("Get GroupList In Cache success")
		groupListResp.GroupList = groupList
		gList <- groupListResp
		return
	} else {
		logger.Debugln("redis is not find")
		for {
			groupListResp, _, err = tg.GetGroupListFromDB(int32(uid), db.DBHandler)
			if err != nil {
				errMap.Store("GetGroupList", err)
				break
			}
			logger.Debugln("start update redis GetGroupListFromDB")
			// 新增到缓存 更新两个地方，首先，每个组的信息要更新，就是group data，记录了群组的id和名字
			if err := tgc.AddGroupInCache(groupListResp, cache.GetRedisClient()); err != nil {
				errMap.Store("GetGroupList", err)
				break
			}

			// 其次更新一个userSet  就是一个组里有哪些用户
			if err := tuc.AddUserInGroupToCache(groupListResp, cache.GetRedisClient()); err != nil {
				errMap.Store("GetGroupList", err)
				break
			}

			// 每个用户的信息
			for _, g := range groupListResp.GroupList {
				for _, u := range g.UserList {
					if err := tuc.AddUserDataInCache(u.Id, []interface{}{
						pub.USER_Id, u.Id,
						pub.IMEI, u.Imei,
						pub.USER_NAME, u.NickName,
						pub.USER_TYPE, u.UserType,
						pub.DEVICE_TYPE, u.DeviceType,
						pub.ONLINE, u.Online,
						pub.LOCK_GID, u.LockGroupId,
					}, cache.GetRedisClient()); err != nil {
						logger.Error("Add user information to cache with error: ", err)
					}
				}
			}

			// 每一个群组拥有的成员
			for _, v := range groupListResp.GroupList {
				if err := tgc.AddGroupCache(v.UserList, v, cache.GetRedisClient()); err != nil {
					errMap.Store("AddGroupCache", err)
					break
				}
			}
			break
		}
		gList <- groupListResp
	}
}
