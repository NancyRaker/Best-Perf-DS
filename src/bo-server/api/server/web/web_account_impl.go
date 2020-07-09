/*
@Time : 2019/4/4 14:16
@Author : yanKoo
@File : web_account_impl
@Software: GoLand
@Description: 实现web端需要用到的关于调度员、管理员、经销商需要用到的GRpc接口
*/
package web

import (
	"context"
	pb "bo-server/api/proto"
	"bo-server/conf"
	tc "bo-server/dao/customer"
	"bo-server/dao/pub"
	tu "bo-server/dao/user"
	tuc "bo-server/dao/user_cache"
	"bo-server/engine/cache"
	"bo-server/engine/db"
	"bo-server/logger"
	"bo-server/utils"
	"net/http"
	"strings"
	"sync"
)

var flag bool  // 用来确定是否转换过数据库引擎

func init() {
	flag = false
}

// 删除账号
func (wssu *WebServiceServerImpl) DeleteAccount(ctx context.Context, req *pb.DeleteAccountReq) (*pb.DeleteAccountResp, error) {
	// 判断该用户下是否有下级用户或者设备
	resp := &pb.DeleteAccountResp{}
	follower, err := tu.QueryDeviceOrAccountByAccountId(req.Id)
	if err != nil {
		resp.Res = &pb.Result{Msg:"delete account fail.",Code:http.StatusInternalServerError}
		return resp, err
	}
	if !follower {
		// 删除该账户
		if err := tc.DeleteAccount(req.Id);err != nil {
			resp.Res = &pb.Result{Msg:"delete account fail.",Code:http.StatusInternalServerError}
			return resp, err
		}
		resp.Res = &pb.Result{Msg:"delete account successfully.",Code:http.StatusOK}
		return resp, nil
	}

	resp.Res = &pb.Result{Msg:"cant delete account because this account has device or lower account.",Code:http.StatusAccepted}
	return resp, nil
}

// 创建用户
func (wssu *WebServiceServerImpl) CreateAccount(ctx context.Context, req *pb.CreateAccountReq) (*pb.CreateAccountResp, error) {
	// TODO 1. 参数校验

	// 2. 添加账户
	user, err := tc.AddAccount(req)
	if err != nil {
		logger.Debugf("CreateAccount error: %+v", err)
		return nil, err
	}

	// 设备更新到缓存
	_ = tuc.AddUserDataInCache(int32(user.Id), []interface{}{
		pub.USER_Id, int32(user.Id),
		pub.IMEI, user.IMei,
		pub.USER_NAME, user.UserName,
		pub.NICK_NAME, user.NickName,
		pub.USER_TYPE, user.UserType,  // 导入设备默认都是1
		pub.ONLINE, pub.USER_OFFLINE, // 加载数据默认全部离线
		pub.DEVICE_TYPE, user.DeviceType,
	}, cache.GetRedisClient())

	// 发送post请求给web-gateway,注册账号 TODO 错误处理
	addrs := strings.Split(conf.WebGatewayAddrs, " ")
	for _, addr := range addrs {
		utils.WebGateWay{Url: addr + conf.AddAccountUrl}.CreateDispatcherPost(&utils.DispatcherAdd{
			AccountName:     user.UserName,
			AccountId: uint(user.Id),
			CreatorId:uint(req.Pid),
		})
	}

	return &pb.CreateAccountResp{Id:int32(user.Id)}, nil
}

// 切换数据库引擎
func (wssu *WebServiceServerImpl) ChangeDBEngine(ctx context.Context, req *pb.Empty) (*pb.Result, error) {
	var (
		err error
		m   sync.Mutex
	)
	m.Lock()
	defer m.Unlock()

	// 默认收到这个请求就更换数据库引擎
	if !flag {
		flag = true // 表示即将切换到另一台
		db.DBHandler, err = db.CreateDBHandler("db2", conf.DEFAULT_CONFIG)
		if err != nil {
			return nil, err
		}

		// 通知im也更换数据库连接

	} else {
		flag = false // 切换回本机
		db.DBHandler, err = db.CreateDBHandler("db", conf.DEFAULT_CONFIG)
		if err != nil {
			return nil, err
		}
		// 通知im也更换数据库连接

	}
	return &pb.Result{Code:http.StatusOK}, nil
}
