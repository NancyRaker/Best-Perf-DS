/*
@Time : 2019/4/6 17:25
@Author : yanKoo
@File : user_impl
@Software: GoLand
@Description:实现了app注册和app设备注册两个rpc调用
*/
package app

import (
	pb "bo-server/api/proto"
	tu "bo-server/dao/user"
	"bo-server/logger"
	"bo-server/model"
	"context"
	"math/rand"
	"strconv"
	"time"
)

// 注册App
func (tcs *TalkCloudServiceImpl) AppRegister(ctx context.Context, req *pb.AppRegReq) (*pb.AppRegRsp, error) {
	iMei := strconv.FormatInt(int64(rand.New(rand.NewSource(time.Now().UnixNano())).Int63n(1000000000000000)), 10)
	appRegResp := &pb.AppRegRsp{}

	// 查重
	ifExist, err := tu.GetUserByName(req.Name)
	if err != nil {
		logger.Debugf("app register error : %s", err)
		appRegResp.Res = &pb.Result{
			Code: 500,
			Msg:  "User registration failed. Please try again later",
		}
		return appRegResp, nil
	}
	if ifExist > 0 {
		appRegResp.Res = &pb.Result{
			Code: 500,
			Msg:  "User name has been registered",
		}
		return appRegResp, nil
	}

	user := &model.User{
		UserName:  req.Name,
		PassWord:  req.Password,
		NickName:  req.Name[len(req.Name)-3 : len(req.Name)],
		AccountId: -1, //  app用户是-1 调度员是0，设备用户是受管理的账号id
		IMei:      iMei,
		UserType:  1,
	}

	logger.Error("app register start")
	if err := tu.AddUser(user); err != nil {
		logger.Debugf("app register error : %s", err)
		appRegResp.Res = &pb.Result{
			Code: 500,
			Msg:  "User registration failed. Please try again later",
		}
		return appRegResp, nil
	}

	res, err := tu.SelectUserByKey(req.Name)
	if err != nil {
		logger.Debugf("app register error : %s", err)
		appRegResp.Res = &pb.Result{
			Code: 500,
			Msg:  "User registration Process failed. Please try again later",
		}
		return appRegResp, nil
	}

	return &pb.AppRegRsp{Id: int32(res.Id), UserName: req.Name, Res: &pb.Result{Code: 200, Msg: "User registration successful"}}, nil
}

// 设备注册
func (tcs *TalkCloudServiceImpl) DeviceRegister(ctx context.Context, req *pb.DeviceRegReq) (*pb.DeviceRegRsp, error) {
	// TODO 设备串号和账户id进行校验
	name := string([]byte(req.DeviceList) /*[9:len(req.DeviceList)]*/)
	user := &model.User{
		UserName: name,
		//NickName: req.DeviceList,
		PassWord:  "123456",
		AccountId: int(req.AccountId),
		IMei:      req.DeviceList,
	}

	if err := tu.AddUser(user); err != nil {
		logger.Debugf("app register error : %s", err)
		return &pb.DeviceRegRsp{Res: &pb.Result{Code: 500, Msg: "Device registration failed. Please try again later"}}, err
	}

	return &pb.DeviceRegRsp{Res: &pb.Result{Code: 200, Msg: "Device registration successful"}}, nil
}

