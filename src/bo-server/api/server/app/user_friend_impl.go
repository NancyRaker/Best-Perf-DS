/*
@Time : 2019/4/6 17:25
@Author : yanKoo
@File : user_friend_impl
@Software: GoLand
@Description:实现了好友相关的四个rpc方法
*/
package app

import (
	pb "bo-server/api/proto"
	"bo-server/dao/user_friend"
	"bo-server/engine/db"
	"bo-server/logger"
	"context"
)

// 添加好友 TODO 暂时不等对方确认加不加好友，直接给你加上
func (serv *TalkCloudServiceImpl) AddFriend(ctx context.Context, req *pb.FriendNewReq) (*pb.FriendNewRsp, error) {
	logger.Debugf("Add friend: uid: %d, friend_id:%d", req.Uid, req.Fuid)
	resp := &pb.FriendNewRsp{Res: &pb.Result{Msg: "Add friend error, please try again later", Code: 500}}
	_, err := user_friend.AddFriend(req.Fuid, req.Uid, db.DBHandler)
	if err != nil {
		logger.Debugf("AddFriend friend add self error: %v", err)
		return resp, err
	}
	_, err = user_friend.AddFriend(req.Uid, req.Fuid, db.DBHandler)
	if err != nil {
		logger.Debugf("AddFriend self add friend error: %v", err)
		return resp, err
	}
	return &pb.FriendNewRsp{Res: &pb.Result{Msg: "Add friend success", Code: 200}}, nil
}

// 获取朋友列表
func (serv *TalkCloudServiceImpl) GetFriendList(ctx context.Context, req *pb.FriendsReq) (*pb.FriendsRsp, error) {
	fList, _, err := user_friend.GetFriendReqList(req.Uid, db.DBHandler)
	if err != nil {
		return &pb.FriendsRsp{Res: &pb.Result{Code: 500, Msg: "process error, please try again"}}, err
	}
	fList.Res = &pb.Result{Msg: "Get friend list success", Code: 200}
	return fList, nil
}

// 根据关键字查询用户,携带是否好友字段
func (serv *TalkCloudServiceImpl) SearchUserByKey(ctx context.Context, req *pb.UserSearchReq) (*pb.UserSearchRsp, error) {
	if req.Target == "" {
		return &pb.UserSearchRsp{Res: &pb.Result{Code: 422, Msg: "process error, please input target"}}, nil
	}

	// SearchUserByName
	uSResp, err := user_friend.SearchUserByName(req.Uid, req.Target, db.DBHandler)
	if err != nil {
		return &pb.UserSearchRsp{Res: &pb.Result{Code: 500, Msg: "process error, please try again"}}, err
	}
	// GetFriendReqList
	_, fMap, err := user_friend.GetFriendReqList(req.Uid, db.DBHandler)
	if err != nil {
		return &pb.UserSearchRsp{Res: &pb.Result{Code: 500, Msg: "process error, please try again"}}, err
	}

	for _, v := range uSResp.UserList {
		if _, ok := (*fMap)[v.Id]; ok {
			v.IsFriend = true
		}
	}
	uSResp.Res = &pb.Result{Msg: "Search User success", Code: 200}
	return uSResp, nil
}

// 删除好友
func (serv *TalkCloudServiceImpl) DelFriend(ctx context.Context, req *pb.FriendDelReq) (*pb.FriendDelRsp, error) {
	_, err := user_friend.RemoveFriend(req.Uid, req.Fuid, db.DBHandler)
	rsp := new(pb.FriendDelRsp)
	rsp.Err = new(pb.Result)
	rsp.Err.Code = 0
	rsp.Err.Msg = ""

	if err != nil {
		rsp.Err.Code = -1
		rsp.Err.Msg = err.Error()
	}

	return rsp, err
}
