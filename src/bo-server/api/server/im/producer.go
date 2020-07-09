/*
@Time : 2019/7/30 14:45 
@Author : yanKoo
@File : producer
@Software: GoLand
@Description: 负责im消息的生成，往消息队列里面投递。
*/
package im

import (
	pb "bo-server/api/proto"
	tg "bo-server/dao/group"
	tgm "bo-server/dao/group_member"
	tlc "bo-server/dao/location"
	tm "bo-server/dao/msg"
	"bo-server/dao/pub"
	s "bo-server/dao/session"
	tu "bo-server/dao/user"
	tuc "bo-server/dao/user_cache"
	"bo-server/engine/cache"
	"bo-server/engine/db"
	"bo-server/logger"
	"bo-server/model"
	"bo-server/utils"
	"context"
	"google.golang.org/grpc/metadata"
	"net/http"
	"strconv"
	"sync"
	"time"
)

const (
	FIRST_LOGIN_DATA                = 1  // 初次登录返回的数据。比如用户列表，群组列表，该用户的个人信息
	OFFLINE_IM_MSG                  = 2  // 用户离线时的IM数据
	IM_MSG_FROM_UPLOAD_OR_WS_OR_APP = 3  // APP和web通过httpClient上传的文件信息、在线时通信的im数据
	KEEP_ALIVE_MSG                  = 4  // 用户登录后，每隔interval秒向stream发送一个消息，测试能不能连通
	LOGOUT_NOTIFY_MSG               = 5  // 用户掉线之后，通知和他在一个组的其他成员
	LOGIN_NOTIFY_MSG                = 6  // 用户上线之后，通知和他在一个组的其他成员
	SOS_MSG                         = 7  // 用户按SOS按键呼救
	SOS_CANCEL_MSG                  = 8  // 用户按SOS取消按键呼救
	WEB_JANUS_NOTIFY                = 9  // web更新群组动作通知给app
	APP_JANUS_NOTIFY                = 10 // APP用户切换janus房间操作群组通知web
	TEMP_GROUP_CREATE_NOTIFY        = 11 // web创建临时组
	TEMP_GROUP_REM_NOTIFY           = 9  // web移除临时组成员
	NORMAL_GROUP_REM_NOTIFY         = 9  // web移除普通组成员
	TEMP_GROUP_DEL_NOTIFY           = 14 // web删除临时组
	NORMAL_GROUP_DEL_NOTIFY         = 15 // web删除普通组
	LOG_TURN_ON_NOTIFY              = 16 // 通知设备开启日志
	LOG_TURN_OFF_NOTIFY             = 17 // 通知设备关闭日志
	APP_CREATE_GROUP_NOTIFY         = 18 // app创建对讲房间，通知其他人进入房间
	APP_REMOVE_GROUP_NOTIFY         = 19 // app创建对讲房间，通知其他人进入房间
	IPS_CHANGED_NOTIFY              = 20 // ips已经切换
	JANUS_TALK_QUALITY_MSG          = 21 // janus通话质量消息

	DEVICE_NICKNAME_CHANGE_MSG      = 22 // janus通话质量消息

	IM_REFRESH_WEB = 10001 // web刷新界面

	IM_MSG_FROM_UPLOAD_RECEIVER_IS_USER  = 1 // APP和web通过httpClient上传的IM信息是发给个人
	IM_MSG_FROM_UPLOAD_RECEIVER_IS_GROUP = 2 // APP和web通过httpClient上传的IM信息是发给群组

	UNREAD_OFFLINE_IM_MSG = 1 // 用户离线消息未读
	READ_OFFLINE_IM_MSG   = 2 // 用户离线消息已读

	CLIENT_EXCEPTION_EXIT = -1 // 客户端异常终止

	WORK_BY_GORONTINE = 2
	WORK_BY_NORMAL    = 1 // 普通调用
)

// 分发的任务
type Task struct {
	Data     *pb.StreamResponse // 具体的消息
	Receiver []int32
	SenderId int32
}

// im推送上下文
type DataContext struct {
	UId              int32
	TempId           chan int32
	Task             chan Task
	ExceptionalLogin chan int32 // 重复登录
	Ctx              context.Context
	Ctf              context.CancelFunc
}

// im推送的对象类
type Client struct {
	WorkType  int32 // 开启协程还是不用  1是不用 2是开协程
	LongLived bool  // 分发多次数据
	Dc        *DataContext
	Ds        DataSource
	Cf        ClientFunc
}

// client需要实现的方法
type ClientFunc interface {
	dispatcher(dc *DataContext, ds DataSource)
	dispatcherScheduler(dc *DataContext, longLived bool)
}

// 消息数据来源
type DataSource interface{}

// gen im task
func NewImTask(senderId int32, receiver []int32, resp *pb.StreamResponse) *Task {
	return &Task{SenderId: senderId, Receiver: receiver, Data: resp}
}

// gen 数据上下文
func NewDataContent(opt ...context.Context) *DataContext {
	var p context.Context
	for _, v := range opt {
		p = v
	}
	if len(opt) == 0 {
		p = context.TODO()
	}

	ctx, ctxFunc := context.WithCancel(p)
	return &DataContext{
		Task:             make(chan Task, 100),
		ExceptionalLogin: make(chan int32, 10),
		TempId:           make(chan int32, 5),
		Ctx:              ctx,
		Ctf:              ctxFunc,
	}
}

// client运行分发数据
func (c *Client) Run() {
	if c.WorkType == WORK_BY_GORONTINE {
		go c.Cf.dispatcher(c.Dc, c.Ds)
		go c.Cf.dispatcherScheduler(c.Dc, c.LongLived)
	}
	if c.WorkType == WORK_BY_NORMAL {
		c.Cf.dispatcher(c.Dc, c.Ds)
		c.Cf.dispatcherScheduler(c.Dc, c.LongLived)
	}
}

// 分发上传文件方式产生的IM数据
func imMessagePublishDispatcher(dc *DataContext, ds DataSource) {
	// 获取要发送的数据
	req := ds.(*pb.ImMsgReqData)

	// web 重连
	if req.MsgType == IM_REFRESH_WEB {
		exitStream(dc, req.Id)
		return
	}

	offlineMem := make([]int32, 0)
	onlineMem := make([]int32, 0)

	logger.Debugf("grpc receive im from web : %+v", req)

	// 获取在线、离线用户id
	if req.ReceiverType == IM_MSG_FROM_UPLOAD_RECEIVER_IS_USER { // 发给单人
		// 判断是否在线
		v := GoroutineMap.GetStream(req.ReceiverId)
		logger.Debugln(req.ReceiverId, v)
		//log.log.Debugf("now dc.StreamMap map have: %+v， %p", StreamMap, &StreamMap)
		if v != nil {
			logger.Debugln(req.ReceiverId, " is online")
			onlineMem = append(onlineMem, req.ReceiverId)
		} else {
			logger.Debugln(req.ReceiverId, " is offline")
			// 保存进数据库
			if err := tm.AddMsg(req, db.DBHandler); err != nil {
				logger.Errorf("Add offline msg with error: ", err)
			}
			offlineMem = append(offlineMem, req.ReceiverId)
		}
		onlineMem = append(onlineMem, req.Id)
	}

	if req.ReceiverType == IM_MSG_FROM_UPLOAD_RECEIVER_IS_GROUP { // 发送给群组成员，然后区分离线在线
		logger.Debugf("want send msg to group %d", req.ReceiverId)
		res, err := tgm.SelectDeviceIdsByGroupId(int(req.ReceiverId))
		if err != nil {
			logger.Errorf("imMessagePublishDispatcher SelectDeviceIdsByGroupId with error: %+v", err)
		}

		logger.Debugf("the group %d has %+v", req.ReceiverId, res)
		//log.log.Debugf("now dc.StreamMap map have: %+v， %p", StreamMap, &StreamMap)
		for _, v := range res {
			//log.log.Debugf("now stream map have:%+v", StreamMap)
			srv := GoroutineMap.GetStream(int32(v))
			logger.Debugf("the group # %d member %d online state: %+v", req.ReceiverId, v, srv)
			if srv != nil {
				onlineMem = append(onlineMem, int32(v))
			} else {
				offlineMem = append(offlineMem, int32(v))
			}
		}
		// 存储离线消息
		logger.Debugf("the offline: %+v， the length is %d", offlineMem, len(offlineMem))
		if len(offlineMem) != 0 {
			if err := tm.AddMultiMsg(req, offlineMem, db.DBHandler); err != nil {
				logger.Error("Add offline msg with error: ", err)
			}
		}
	}

	logger.Debugf("web api want send to : %+v", onlineMem)
	resp := &pb.StreamResponse{
		DataType:  IM_MSG_FROM_UPLOAD_OR_WS_OR_APP,
		ImMsgData: req,
		Res:       &pb.Result{Code: http.StatusOK, Msg: "receiver im message successful"},
	}
	// 发送在线用户消息
	if onlineMem != nil {
		dc.Task <- *NewImTask(req.Id, onlineMem, resp)
		logger.Debugf("dispatcher finish %d <-||||-> %+v", req.Id, resp)
	}
}

// 主动结束stream连接
func exitStream(dc *DataContext, uId int32) {
	logger.Debugf("will exit stream, start del status in cache")
	if err := s.DeleteSession(uId, cache.GetRedisClient()); err != nil {
		logger.Error("Update user online state error:", err)
	}

	// 如果存在临时组，就删除临时组
	logger.Debugln("start check temp group resource")
	if gId, err := tg.CheckDuplicateCreateGroup(uId); err != nil {
		logger.Debugf("imMessagePublishDispatcher CheckDuplicateCreateGroup is has error: %+v", err)
	} else {
		logger.Debugf("start remove temp group %d", gId)
		if gId != 0 {
			var c = Common{}
			res, err := c.RemoveGroup(context.TODO(), &pb.GroupDelReq{Gid: gId, Uid:uId}, WEB_JANUS_NOTIFY)
			logger.Debugf("imMessagePublishDispatcher Remove temp group res:%+v, err:%+v", res, err)
		}
	}
	logger.Debugf("end exit stream")
	//TODO 退出需要清理scheduler
	dc.ExceptionalLogin <- uId
	dc.Task <- *NewImTask(uId, []int32{uId}, &pb.StreamResponse{})
}

// dispatcher 根据数据类型，调用不同的函数，产生不同的数据，然后把数据放到发送chan中
func singleDataDispatcher(dc *DataContext, ds DataSource) {
	logger.Debugf("dispatcher client msg")
	// 1. 转换为stream接口类型
	srv := ds.(pb.TalkCloud_DataPublishServer)

	// 正式开始等待数据
	for {
		select {
		case <-dc.Ctx.Done():
			return
		default:
			// 2.0  调用接口发方法获取客户端发送的数据data
			data, err := srv.Recv()
			logger.Debugf("receive this stream data: %+v, err: %+v", data, err)

			if data == nil {
				logger.Debugf("this # %d stream is no data, maybe is offline", dc.UId)
				dc.ExceptionalLogin <- dc.UId // TODO 这里加上可能导致app重连失败
				return
			}

			// 2.1 如果id不合法那么就给dc的chan里面写一个终止信号
			if data.Uid <= 0 {
				logger.Error("this stream is id is bad")
				dc.ExceptionalLogin <- data.Uid
				return
			}
			logger.Debugf("receive this stream data id: %d, err: %+v", data.Uid, err)
			dc.TempId <- data.Uid

			// 2.2 重复登录调用DataPublish方法， 改用sessionId来校验
			md, ok := metadata.FromIncomingContext(srv.Context()) // get context from stream
			if !ok || md == nil || md.Len() <= 0 || len(md.Get("session-id")) <= 0 {
				logger.Error("this metadata is bad")
				dc.ExceptionalLogin <- data.Uid
				return
			}
			//logger.Debugf("metadata: %+v", md)
			sessionIds := md.Get("session-id")
			if sessionIds == nil || len(sessionIds) != 1 {
				logger.Error("sessionIds this metadata is bad")
				dc.ExceptionalLogin <- data.Uid
				return
			}
			sessionId := sessionIds[0]

			// 2.3 判断metadata里面的sessionId是否合法
			valid, err := s.CheckSession(md.Get("session-id")[0], data.Uid)
			if (err != nil) || (!valid) {
				logger.Debugf("DataPush CheckSession is invalid %+v", err)
				dc.ExceptionalLogin <- data.Uid
				return
			}

			// 3.0 是否是重复登录,如果重复登录就覆盖前一个，把前一个退出
			GoroutineMap.GetAndSet(data.Uid, newClientContext(dc, srv))

			// 3.1 更新stream和redis状态
			// 即时是其他类型的消息，也会走这里，也可以看做是心跳
			if err := s.UpdateSessionInCache(&model.SessionInfo{Id: data.Uid, SessionId: sessionId, Online: pub.USER_ONLINE}); err != nil {
				logger.Errorf("Update user online state error:", err)
			}

			switch data.DataType {
			case OFFLINE_IM_MSG:
				res, err := getOfflineImMsgFromDB(data)
				if err != nil {
					logger.Errorf("getOfflineImMsgFromDB error:", err)
				}

				dc.Task <- *NewImTask(data.Uid, []int32{data.Uid}, res)

				// 往dc里面写上线通知
				//go NotifyToOther(dc.Task, data.Uid, LOGIN_NOTIFY_MSG)

			case IM_MSG_FROM_UPLOAD_OR_WS_OR_APP:
				// im消息
				imMessagePublishDispatcher(dc, data.ImMsg)

			case KEEP_ALIVE_MSG:
				logger.Debugf("# %d receive heat", data.Uid)
				dc.Task <- *NewImTask(data.Uid, []int32{data.Uid},
					&pb.StreamResponse{DataType: KEEP_ALIVE_MSG, KeepAlive: &pb.KeepAlive{Uid: data.Uid, Ack: http.StatusOK}})
				if data.DeviceInfo != nil {
					_ = tuc.AddUserDataInCache(data.Uid, []interface{}{
						pub.BATTERY, data.DeviceInfo.Battery,
					}, cache.GetRedisClient())
				}
			}
		}
	}
}

// TODO 直接扔掉这个调度器，直接所有数据往全局队列里送
func singleDataScheduler(dContent *DataContext, multiSend bool) {
	logger.Debugf("start Scheduler im msg")
	ip, err := utils.GetClietIP(dContent.Ctx)
	if err != nil {

	}
	var notify int32
	tick := time.Tick(time.Minute * 5)
	for {
		// 接收任务
		select {
		case id := <-dContent.TempId:
			dContent.UId = id
		case <-dContent.Ctx.Done():
			logger.Debugln("the scheduler will cancel")
			return
		case t := <-dContent.Task:
			go func() { GlobalTaskQueue.Tasks <- t }()
			if t.Data.DataType == LOGOUT_NOTIFY_MSG ||
				t.Data.DataType == SOS_MSG ||
				t.Data.DataType == SOS_CANCEL_MSG ||
				t.Data.DataType == WEB_JANUS_NOTIFY {
				notify++
				logger.Debugf("notify: %d, total: %d", notify, t.Data.NotifyTotal)
				if notify == t.Data.NotifyTotal {
					logger.Debugf("# %d scheduler finish", dContent.UId)
					return
				}
			}
			if !multiSend {
				logger.Debugf("only scheduler once")
				return
			}
		case <-tick:
			if dContent.UId == 0 {
				dContent.ExceptionalLogin <- -8
			}
			logger.Debugf("uid ip : %s single # %d im task queue", ip, dContent.UId) // TODO 合理退出，关闭调度器
		}
	}
}

// 返回的IM离线数据
func getOfflineImMsgFromDB(req *pb.StreamRequest) (*pb.StreamResponse, error) {
	// 去数据库拉取离线数据
	logger.Debugf("start get offline im msg")
	offlineMsg, err := tm.GetMsg(req.Uid, UNREAD_OFFLINE_IM_MSG, db.DBHandler)
	if err != nil {
		logger.Error("Get offline msg fail with error:", err)
		return nil, err
	}
	logger.Debugf("get offline msg %+v", offlineMsg)

	var (
		idIndexSMap     = map[int32]int{}
		idIndexGMap     = map[int32]int{}
		idIndexGPttMap  = map[int32]int{}
		respPkgGroupPtt = make([]*pb.OfflineImMsg, 0)
		respPkgSingle   = make([]*pb.OfflineImMsg, 0)
		respPkgGroup    = make([]*pb.OfflineImMsg, 0)
		idxGpPtt        = 0
		idxG            = 0
		idxS            = 0
	)
	// 遍历离线数据集，记录数据用户id和位置
	for _, msg := range offlineMsg {
		//msg.MsgCode = msg.SendTime  // TODO 和app讨论这两个字段，时间戳还是sendtime的时间格式
		//msg.SendTime = utils.UnixStrToTimeFormat(msg.SendTime)
		if msg.ReceiverType == IM_MSG_FROM_UPLOAD_RECEIVER_IS_USER {
			if v, ok := idIndexSMap[msg.Id]; ok {
				// 已经发现了这个用户的一条消息，那么就把消息加到对应的切片下的
				respPkgSingle[v].ImMsgData = append(respPkgSingle[v].ImMsgData, msg)
			} else {
				// 首次找到这个用户的第一条单人消息，就respPackage添加一个slice，并记录index
				var userMsgs = &pb.OfflineImMsg{
					SenderId:        msg.Id,
					Name:            msg.SenderName,
					MsgReceiverType: IM_MSG_FROM_UPLOAD_RECEIVER_IS_USER,
				}
				userMsgs.ImMsgData = append(make([]*pb.ImMsgReqData, 0), msg)
				respPkgSingle = append(respPkgSingle, userMsgs)
				idIndexSMap[msg.Id] = idxS
				idxS++

			}
		}

		// 群组
		if msg.ReceiverType == IM_MSG_FROM_UPLOAD_RECEIVER_IS_GROUP {

			if msg.MsgType == utils.IM_PTT_MSG {
				if v, ok := idIndexGPttMap[msg.ReceiverId]; ok {
					// 已经发现了这个用户的一条消息，那么就把消息加到对应的切片下的
					logger.Debugf("v %d, s %v msg %+v", v, ok, msg)
					respPkgGroupPtt[v].ImMsgData = append(respPkgGroupPtt[v].ImMsgData, msg)
				} else {
					// 首次找到这个用户的第一条群聊ptt消息，就respPackage添加一个slice，并记录index
					var userMsgs = &pb.OfflineImMsg{
						GroupId:         msg.ReceiverId,
						Name:            msg.ReceiverName,
						MsgReceiverType: IM_MSG_FROM_UPLOAD_RECEIVER_IS_GROUP,
					}
					userMsgs.ImMsgData = append(make([]*pb.ImMsgReqData, 0), msg)
					respPkgGroupPtt = append(respPkgGroupPtt, userMsgs)
					idIndexGPttMap[msg.ReceiverId] = idxGpPtt
					idxGpPtt++
				}
			} else {
				if v, ok := idIndexGMap[msg.ReceiverId]; ok {
					// 已经发现了这个用户的一条消息，那么就把消息加到对应的切片下的
					logger.Debugf("v %d, s %v msg %+v", v, ok, msg)
					respPkgGroup[v].ImMsgData = append(respPkgGroup[v].ImMsgData, msg)
				} else {
					// 首次找到这个用户的第一条单人消息，就respPackage添加一个slice，并记录index
					var userMsgs = &pb.OfflineImMsg{
						GroupId:         msg.ReceiverId,
						Name:            msg.ReceiverName,
						MsgReceiverType: IM_MSG_FROM_UPLOAD_RECEIVER_IS_GROUP,
					}
					userMsgs.ImMsgData = append(make([]*pb.ImMsgReqData, 0), msg)
					respPkgGroup = append(respPkgGroup, userMsgs)
					idIndexGMap[msg.ReceiverId] = idxG
					idxG++
				}
			}
		}
	}

	logger.Debugf("%+v \n %+v", respPkgSingle, respPkgGroup)

	return &pb.StreamResponse{
		OfflineImMsgResp: &pb.OfflineImMsgResp{
			OfflineSingleImMsgs:   respPkgSingle,
			OfflineGroupImMsgs:    respPkgGroup,
			OfflineGroupPttImMsgs: respPkgGroupPtt},
		DataType: OFFLINE_IM_MSG}, nil
}

// 上线通知sos通知, 仅在当前janus组
func notifySosToOther(dcTask chan Task, req *pb.Device, notifyType int32) {
	var (
		errMap      = &sync.Map{}
		selfGList   = make(chan *pb.GroupListRsp, 1)
		notifyTotal = int32(0)
		notifyMap   = make(map[int32]bool)
		//uLocation   *pb.GPS
		gpsDataResp *pb.GPSHttpResp
	)
	logger.Debugf("notify root id :%d", req.Id)
	uInfo, _ := tuc.GetUserFromCache(req.Id)
	user, _ := tu.SelectUserByKey(int(req.Id))
	gpsDataResp, _, _ = tlc.GetUserLocationInCache(req.Id, cache.GetRedisClient())

	GetGroupList(req.Id, selfGList, errMap, nil)
	gl := <-selfGList
	logger.Debugf("notifySosToOther group list: %+v", gl)
	if gl != nil && uInfo != nil {
		for _, g := range gl.GroupList {
			if g.Gid != req.CurrentGroup {
				continue
			}
			for _, u := range g.UserList {
				if u.Id != req.Id && (u.Online == pub.USER_ONLINE || u.Online == pub.USER_JANUS_ONLINE) &&
					!notifyMap[u.Id] && u.LockGroupId == req.CurrentGroup { // 每个人只发送一次，并且只发送给当前janus房间里有的人
					logger.Debugf("will notify - - - - - - - - - - - >%d", u.Id)
					notifyMap[u.Id] = true
				}
			}
		}
		if user != nil { // 调度员
			notifyMap[int32(user.AccountId)] = true
		}
		notifyMap[req.Id] = true
		notifyTotal = int32(len(notifyMap))
		logger.Debugf("notify total: %d and notify type: %d with all id: %+v", notifyTotal, notifyType, notifyMap)

		for u := range notifyMap {
			doNotify(&pb.DeviceInfo{Id: u}, errMap, notifyType, notifyTotal, uInfo, gpsDataResp, dcTask, req.Id)
		}
	} else {
		logger.Errorf("cant load notify info %v----------%v", gl, uInfo)
	}
}

func doNotify(u *pb.DeviceInfo, errMap *sync.Map, notifyType int32, notifyTotal int32,
	uInfo *pb.DeviceInfo, gpsDataResp *pb.GPSHttpResp, dcTask chan Task, uId int32, opt ...interface{}) {
	logger.Debugf("do notify type info: %d", notifyType)
	// 对于群里每一位都要通知到
	var (
		recvId     = make([]int32, 0)
		notifyInfo *pb.LoginOrLogoutNotify
		gList      = make(chan *pb.GroupListRsp, 1)
	)

	recvId = append(recvId, u.Id)

	if notifyType == WEB_JANUS_NOTIFY {
		logger.Debugf("notify start get group list %d", notifyType)
		GetGroupList(u.Id, gList, errMap, nil)
		notifyInfo = &pb.LoginOrLogoutNotify{UserInfo: uInfo,  GroupList: (<-gList).GroupList}

	} else if notifyType == TEMP_GROUP_CREATE_NOTIFY || notifyType == TEMP_GROUP_REM_NOTIFY { // 临时组
		GetGroupList(u.Id, gList, errMap, nil)
		var gId int32
		var groupName string
		for i, v := range opt {
			if i == 0 && v != nil {
				gId = v.(int32)
			} else if i == 1 && v != nil {
				groupName = v.(string)
			}
		}
		notifyInfo = &pb.LoginOrLogoutNotify{
			UserInfo: uInfo,
			GroupInfo: &pb.GroupInfo{
				Gid:       gId,
				GroupName: groupName,
				Status:    tg.TEMP_GROUP},
			GroupList: (<-gList).GroupList}
	} else if notifyType == NORMAL_GROUP_REM_NOTIFY {
		var gId int32
		for i, v := range opt {
			if i == 0 && v != nil {
				gId = v.(int32)
			}
		}
		GetGroupList(u.Id, gList, errMap, nil)
		notifyInfo = &pb.LoginOrLogoutNotify{
			UserInfo:  uInfo,
			GroupInfo: &pb.GroupInfo{Gid: gId, Status: tg.NORMAL_GROUP},
			GroupList: (<-gList).GroupList}
	} else {
		notifyInfo = &pb.LoginOrLogoutNotify{UserInfo: uInfo, GpsResp: gpsDataResp}
	}

	resp := &pb.StreamResponse{
		DataType:    notifyType,
		NotifyTotal: notifyTotal,
		Notify:      notifyInfo,
		Res:         &pb.Result{Code: 200, Msg: strconv.FormatInt(int64(u.Id), 10) + " notify successful"},
	}
	logger.Debugf("will send %d notify to %+v", notifyType, recvId)
	dcTask <- *NewImTask(uId, recvId, resp)
}

// 上线通知所有人，掉线通知所有人
func NotifyToOther(dcTask chan Task, uId int32, notifyType int32) {
	var (
		errMap      = &sync.Map{}
		selfGList   = make(chan *pb.GroupListRsp, 1)
		notifyTotal = int32(0)
		notifyMap   = make(map[int32]bool)
		//uLocation   *pb.GPS
		gpsDataResp *pb.GPSHttpResp
	)
	logger.Debugf("notify root id :%d", uId)
	uInfo, _ := tuc.GetUserFromCache(uId)
	gpsDataResp, _, _ = tlc.GetUserLocationInCache(uId, cache.GetRedisClient())

	GetGroupList(uId, selfGList, errMap, nil)
	gl := <-selfGList
	if gl != nil && uInfo != nil {
		for _, g := range gl.GroupList {
			for _, u := range g.UserList {
				if u.Id != uId && (u.Online == pub.USER_ONLINE || u.Online == pub.USER_JANUS_ONLINE) && !notifyMap[u.Id] {
					logger.Debugf("will notify - - - - - - - - - - - >%d", u.Id)
					notifyMap[u.Id] = true
				}
			}
		}
		notifyMap[uId] = true
		notifyTotal = int32(len(notifyMap))
		logger.Debugf("notify total: %d and notify all id: %+v", notifyTotal, notifyMap)

		for u := range notifyMap {
			doNotify(&pb.DeviceInfo{Id: u}, errMap, notifyType, notifyTotal, uInfo, gpsDataResp, dcTask, uId)
		}
	} else {
		logger.Errorf("cant load notify info %v----------%v", gl, uInfo)
	}
}

// 写的很简单，目前只用来给web发送一个更新groupList的提示消息 会有一个问题，把这个消息扔来这里的时候，阻塞了怎么办。。。
func SendSingleNotify(receiverId []int32, senderId int32, resp *pb.StreamResponse) {
	logger.Debugf("will send %d notify to %+v, with %+v", senderId, receiverId, resp)
	GlobalTaskQueue.Tasks <- *NewImTask(senderId, receiverId, resp)
}
