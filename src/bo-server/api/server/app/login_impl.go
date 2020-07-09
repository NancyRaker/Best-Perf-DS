/*
@Time : 2019/4/12 19:29
@Author : yanKoo
@File : talk_cloud_app_login_impl
@Software: GoLand
@Description:实现登录和注销两个rpc调用
*/
package app

import (
	pb "bo-server/api/proto"
	"bo-server/api/server/im"
	cfgGs "bo-server/conf"
	"bo-server/dao/login_record"
	"bo-server/dao/pub"
	s "bo-server/dao/session"
	ss "bo-server/dao/session"
	tu "bo-server/dao/user"
	tuc "bo-server/dao/user_cache"
	tuf "bo-server/dao/user_friend"
	"bo-server/engine/cache"
	"bo-server/engine/db"
	"bo-server/logger"
	"bo-server/model"
	"bo-server/utils"
	"context"
	"database/sql"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"strconv"
	"sync"
	"time"
)

const DISPATCHER = 2 // 数据库中调度员的类型

// TalkCloud 服务的所有方法实现
type TalkCloudServiceImpl struct{}

// 用户登录
func (tcs *TalkCloudServiceImpl) Login(ctx context.Context, req *pb.LoginReq) (*pb.LoginRsp, error) {
	start := time.Now().UnixNano()
	logger.Debugf("%s enter login with time: %d with params: %+v", req.Name, start, req)
	//　验证用户名是否存在以及密码是否正确，然后就生成一个uuid session, 把sessionid放进metadata返回给客户端，
	//  然后之后的每一次连接都需要客户端加入这个metadata，使用拦截器，对用户进行鉴权
	if req.Name == "" || req.Passwd == "" || req.GrpcServer == "" {
		return &pb.LoginRsp{Res: &pb.Result{Code: 422, Msg: "login fail with invalid params: " + req.String()}}, nil
	}

	// 校验版本号
	if req.AppVersionCode == 0 && req.AppVersion == "" {
		return &pb.LoginRsp{Res: &pb.Result{Code: 422, Msg: "app login fail with invalid version params: " + req.String()}}, nil
	}

	if req.AppVersion == "" && req.AppVersionCode < int32(cfgGs.AppVersionCode) { // 暂时版本号为2是最新版本，小于2就返回url
		return &pb.LoginRsp{
			UpdateUri: cfgGs.AppUrl,
			Res:       &pb.Result{Code: 301, Msg: "app login fail with invalid version: " + req.String()}}, nil
	}
	if req.AppVersion == "" {
		req.AppVersion = strconv.Itoa(int(req.AppVersionCode))
	}

	// 校验用户是否存在
	res, err := tu.SelectUserByKey(req.Name)
	if err != nil && err != sql.ErrNoRows {
		logger.Debugf("App login >>> err != nil && err != sql.ErrNoRows <<< error : %s", err)
		loginRsp := &pb.LoginRsp{
			Res: &pb.Result{
				Code: 500,
				Msg:  "User Login Process failed. Please try again later"},
		}
		return loginRsp, nil
	}

	if err == sql.ErrNoRows {
		logger.Debugf("App login error : %s", err)
		loginRsp := &pb.LoginRsp{
			Res: &pb.Result{
				Code: 501,
				Msg:  "User is not exist error. Please try again later"},
		}
		return loginRsp, nil
	}

	if res.PassWord != req.Passwd {
		logger.Debugf("App login pwd res.PassWord :%s req.Pwd: %s error : %+v", res.PassWord, req.Passwd, err)
		loginRsp := &pb.LoginRsp{Res: &pb.Result{Code: 500, Msg: "User Login pwd error. Please try again later"}}
		return loginRsp, nil
	}
	userInfo := &pb.Member{
		Id:          int32(res.Id),
		IMei:        res.IMei,
		UserName:    res.UserName,
		NickName:    res.NickName,
		UserType:    int32(res.UserType),
		LockGroupId: int32(res.LockGroupId),
		StartLog:    int32(res.StartLog),
		Online:      pub.USER_ONLINE, // 登录就在线
	}

	// create and send header
	newSid, err := insertSessionInfo(userInfo)
	header := metadata.Pairs("session-id", newSid)
	grpc.SendHeader(ctx, header)

	// 如果是调度员等用户，直接返回
	if res.UserType >= DISPATCHER {
		logger.Debugf("login account # %d type is dispatcher", res.Id)
		return &pb.LoginRsp{Res: &pb.Result{Code: 200, Msg: req.Name + " login successful"}}, nil
	}

	var (
		errMap   = &sync.Map{}
		wg       sync.WaitGroup
		fList    = make(chan *pb.FriendsRsp, 1)
		gList    = make(chan *pb.GroupListRsp, 1)
		existErr bool
	)
	defer func() {
		close(fList)
		close(gList)
	}()

	// 0. 处理登录session
	//processSession(req, errMap, &wg)
	// 往dc里面写上线通知 TODO 主协程退出会有问题
	//var notify = im.ServerWorker{}
	//go func() {
	//_, _ = notify.ImLoginNotifyPublish(context.TODO(), &model.NotifyInfo{Id: userInfo.Id, NotifyType: im.LOGIN_NOTIFY_MSG})
	go im.NotifyToOther(im.GlobalTaskQueue.Tasks, userInfo.Id, im.LOGIN_NOTIFY_MSG)

	// 1. 将登陆ip存入mysql
	ip, err := utils.GetClietIP(ctx)
	if err != nil {
		errMap.Store("GetClietIP", err)
	}
	err = login_record.SaveDeviceLoginInfo(&login_record.LoginSaveData{
		UId:        userInfo.Id,
		Ip:         ip,
		LoginTime:  time.Now().Unix(),
		LoginAddr:  req.GrpcServer,
		AppVersion: strconv.Itoa(int(req.AppVersionCode)),
	})

	if err != nil {
		errMap.Store("SaveDeviceLoginInfo", err)
	}
	logger.Debugf("%s login start prepare data with time : %d", req.Name, time.Now().UnixNano())
	// 2. 将用户信息添加进redis
	addUserInfoToCache(userInfo, &wg)

	logger.Debugf("%s login start prepare friend data with time : %d", req.Name, time.Now().UnixNano())
	// 3. 获取好友列表
	getFriendList(int32(res.Id), fList, errMap, &wg)

	logger.Debugf("%s login start prepare group data with time : %d", req.Name, time.Now().UnixNano())
	// 4. 群组列表
	im.GetGroupList(int32(res.Id), gList, errMap, &wg)

	//wg.Wait()

	//log.log.Println("----------------->test here--------------------------")
	//遍历该map，参数是个函数，该函数参的两个参数是遍历获得的key和value，返回一个bool值，当返回false时，遍历立刻结束。
	errMap.Range(func(k, v interface{}) bool {
		err = v.(error)
		if err != nil {
			logger.Errorf("%s gen error: ", k, err)
			existErr = true
			return false
		}
		return true
	})

	//log.log.Println("----------------->test now--------------------------")
	if existErr {
		return &pb.LoginRsp{Res: &pb.Result{Code: 500, Msg: "process error, please try again"}}, nil
	}
	//log.log.Println("----------------->test--------------------------")

	loginRep := &pb.LoginRsp{
		UserInfo:   userInfo,
		FriendList: (<-fList).FriendList,
		GroupList:  (<-gList).GroupList,
		Res:        &pb.Result{Code: 200, Msg: req.Name + " login successful"},
		SessionId:  newSid,
	}

	end := time.Now().UnixNano()
	logger.Debugf("%s login done with time : %d, used: %d ms", req.Name, end, (end-start)/1000000)
	return loginRep, nil
}

// 更新session 主要是登录的时候用  TODO 得要判断是不是重复登录，去踢人下线
func insertSessionInfo(aInfo *pb.Member) (string, error) {
	newSid, _ := utils.NewUUID()
	sInfo := &model.SessionInfo{
		SessionId: newSid,
		Id:        aInfo.Id,
		Online:    60 * 3, // 300 s 过期时间
	}

	// 直接覆盖，登录嘛，必须直接覆盖
	if err := ss.InsertSession(sInfo); err != nil {
		return "", err
	}

	return newSid, nil
}

// 处理登录session validate TODO 给stream模式加metadata
func processSession(req *pb.StreamRequest, errMap *sync.Map, wg *sync.WaitGroup) {
	defer wg.Done()
	sessionId, err := utils.NewUUID()
	if err != nil {
		logger.Debugf("session id is error%s", err)
		errMap.Store("processSession", err)
		return
	}

	sInfo := &model.SessionInfo{SessionId: sessionId, Online: pub.USER_ONLINE, Id: req.Uid}
	if err := s.InsertSession(sInfo); err != nil {
		logger.Debugf("session id insert is error%s", err)
		errMap.Store("processSession", err)
		return
	}
}

// 获取好友列表
func getFriendList(uid int32, fList chan *pb.FriendsRsp, errMap *sync.Map, wg *sync.WaitGroup) {
	logger.Debugln("get FriendList start")
	var err error
	fl, _, err := tuf.GetFriendReqList(int32(uid), db.DBHandler)
	if err != nil {
		errMap.Store("getFriendList", err)
		fList <- nil
	} else {
		fList <- fl
	}
	logger.Debugln("get FriendList done")
}

// 增加缓存
func addUserInfoToCache(userInfo *pb.Member, wg *sync.WaitGroup) {
	redisCli := cache.GetRedisClient()
	if err := tuc.AddUserDataInCache(userInfo.Id, []interface{}{
		pub.USER_Id, userInfo.Id,
		pub.IMEI, userInfo.IMei,
		pub.USER_NAME, userInfo.UserName,
		pub.NICK_NAME, userInfo.NickName,
		pub.USER_TYPE, userInfo.UserType,
		pub.LOCK_GID, userInfo.LockGroupId,
		pub.ONLINE, userInfo.Online,
	}, redisCli); err != nil {
		logger.Error("Add user information to cache with error: ", err)
	}
	logger.Debugln("addUserInfoToCache done")
}

// 用户注销
func (tcs *TalkCloudServiceImpl) Logout(ctx context.Context, req *pb.LogoutReq) (*pb.LogoutRsp, error) {
	//md, ok := metadata.FromIncomingContext(ctx)
	//if !ok {
	//	log.log.Error("session id metadata set  error")
	//	return &pb.LogoutRsp{Res: &pb.Result{Code: 403, Msg: "server internal error"}}, nil
	//}
	// TODO 考虑要不要验证sessionInfo中的name和password
	if err := s.DeleteSession(0, nil); err != nil {
		logger.Debugf("sessionid metadata delete  error%s", err)
		return &pb.LogoutRsp{Res: &pb.Result{Code: 500, Msg: "server internal error"}}, err
	}
	return &pb.LogoutRsp{Res: &pb.Result{Code: 200, Msg: req.Name + "logout successful"}}, nil
}
