/*
@Time : 2019/4/12 19:29
@Author : yanKoo
@File : instance
@Software: GoLand
@Description: im消息的一些实例
*/
package im

import (
	pb "bo-server/api/proto"
	tgc "bo-server/dao/group_cache"
	tuc "bo-server/dao/user_cache"
	"bo-server/engine/cache"
	"bo-server/internal/mq/mq_receiver"
	"bo-server/logger"
	"bo-server/model"
	"bo-server/utils"
	"context"
	"errors"
	"net/http"
	"strconv"
	"sync"
	"time"
)

// im相关的三个rpc方法的具体实现
type ServerWorker struct {
	NotifyType int32
}

const (
	SOS_DATA_REPORT   = 2 // 发布sos消息
	SOS_CANCLE_REPORT = 3 // 发布取消sos消息
)

/***
  * 文件上传等im消息
  */
type simpleImClientFuncImpl struct{}

func (simpleImClientFuncImpl) dispatcher(dc *DataContext, ds DataSource)           { imMessagePublishDispatcher(dc, ds) }
func (simpleImClientFuncImpl) dispatcherScheduler(dc *DataContext, longLived bool) { singleDataScheduler(dc, longLived) }

// 上传文件方式产生的IM数据推送
func (ServerWorker) ImMessagePublish(ctx context.Context, req *pb.ImMsgReqData) (*pb.ImMsgRespData, error) {
	c := &Client{
		WorkType:  WORK_BY_NORMAL,
		Dc:        NewDataContent(),
		Ds:        req,
		LongLived: false,
		Cf:        simpleImClientFuncImpl{},
	}
	c.Run()
	logger.Debugf("# %d im once done", req.Id)
	return &pb.ImMsgRespData{Result: &pb.Result{Msg: "push data done", Code: 200}, MsgCode: req.MsgCode}, nil
}

/***
  * sos消息的推送实现
  */
type sosImImpl struct{}

func (sosImImpl) dispatcher(dc *DataContext, ds DataSource) {
	var (
		req        = ds.(*pb.ReportDataReq)
		notifyType int32
	)

	// 消息进行分类  现在应该不需要了
	if req.DataType == SOS_DATA_REPORT {
		notifyType = SOS_MSG
	} else if req.DataType == SOS_CANCLE_REPORT {
		notifyType = SOS_CANCEL_MSG
	}
	notifySosToOther(dc.Task, req.DeviceInfo, notifyType)
}

func (sosImImpl) dispatcherScheduler(dc *DataContext, longLived bool) { singleDataScheduler(dc, longLived) }

// 分发sos消息
func (ServerWorker) ImSosPublish(ctx context.Context, req *pb.ReportDataReq) (*pb.ImMsgRespData, error) {
	c := &Client{
		WorkType:  WORK_BY_GORONTINE,
		Dc:        NewDataContent(),
		Ds:        req,
		LongLived: true,
		Cf:        sosImImpl{},
	}
	c.Dc.UId = -2
	c.Run()
	logger.Debugf("# %d im sos once done", req.DeviceInfo.Id)
	return &pb.ImMsgRespData{Result: &pb.Result{Msg: "push data done", Code: 200}}, nil
}

/**
 * Im消息主要的推送
 */
type imDataPublish struct{}

// Im文件等消息主要的推送的接口实现
func (imDataPublish) dispatcher(dc *DataContext, ds DataSource) { singleDataDispatcher(dc, ds) }

func (imDataPublish) dispatcherScheduler(dc *DataContext, longLived bool) { singleDataScheduler(dc, longLived) }

// 分发登录返回数据、IM离线数据、IM离线数据、Heartbeat
func (ServerWorker) DataPublish(srv pb.TalkCloud_DataPublishServer) error {
	logger.Debugf("Wait process start with time:%d", time.Now().UnixNano())
	// TODO 根据ip限制调用次数
	//ip, err := utils.GetClietIP(srv.Context())
	//if err != nil {
	//}
	//if ip == "116.30.199.130" {
	//	return nil
	//}
	c := &Client{
		WorkType:  WORK_BY_GORONTINE,
		Dc:        NewDataContent(srv.Context()),
		Ds:        srv,
		LongLived: true,
		Cf:        imDataPublish{},
	}
	c.Run()

	// 重复登录就直接返回, 正常退出也返回
	logger.Debugf("Wait process end")
	uid := <-c.Dc.ExceptionalLogin
	err := errors.New("The user with id " + strconv.FormatInt(int64(uid), 10) + " is invalid call")
	logger.Debugln(err)
	err = srv.Send(&pb.StreamResponse{
		Res: &pb.Result{
			Msg:  err.Error(),
			Code: http.StatusUnauthorized,
		},
	})
	c.Dc.Ctf()
	return err
}

/**
 * 设备对讲语音消息的推送实现
 */
type pttImMsgImpl struct{}

// 从消息队列获取ptt消息
func (p pttImMsgImpl) receiverPttMsg(pttD chan *model.InterphoneMsg) {
	for {
		select {
		case msg := <-mq_receiver.PttMessage():
			pttD <- msg
		}
	}
}

// 实现推送的client中的dispatcher方法
func (p pttImMsgImpl) dispatcher(dc *DataContext, ds DataSource) {
	// redis获取对讲音频信息，分发
	pttC := make(chan *model.InterphoneMsg, 100)
	go p.receiverPttMsg(pttC)

	var msgQueues []*model.InterphoneMsg
	var dispatcherChan = p.createPttDispatcher(dc)
	//tick := time.NewTicker(time.Second * time.Duration(5))
	for {
		var activeChan chan *model.InterphoneMsg
		var activeMsg *model.InterphoneMsg
		if len(msgQueues) > 0 {
			activeChan = dispatcherChan
			activeMsg = msgQueues[0]
		}

		select {
		case t := <-pttC:
			msgQueues = append(msgQueues, t)
		case activeChan <- activeMsg:
			msgQueues = msgQueues[1:]
		}
	}
}

// ptt消息分发
func (p pttImMsgImpl) createPttDispatcher(dc *DataContext) chan *model.InterphoneMsg {
	tc := make(chan *model.InterphoneMsg)
	go p.pttMidHandler(tc, dc)
	return tc
}

// 中间处理一下。json反序列化
func (p pttImMsgImpl) pttMidHandler(c chan *model.InterphoneMsg, dc *DataContext) {
	go func() {
		for {
			pttMsg := <-c
			logger.Debugf("Will send Ptt msg: %s", pttMsg)
			if pttMsg != nil && pttMsg.Uid != "" && pttMsg.GId != "" {
				p.pttMsgDispatcher(dc, pttMsg)
			}
		}
	}()
}

// 处理ptt消息
func (p pttImMsgImpl) pttMsgDispatcher(dc *DataContext, pttMsg *model.InterphoneMsg) {
	uId, _ := strconv.ParseInt(pttMsg.Uid, 10, 64)
	imU, err := tuc.GetUserFromCache(int32(uId))
	if err != nil {
		logger.Debugf("pttImMsgImpl dispatcher GetUserFromCache error: %+v", err)
	}

	gId, _ := strconv.ParseInt(pttMsg.GId, 10, 64)
	imG, err := tgc.GetGroupData(int32(gId), cache.GetRedisClient())
	if err != nil {
		logger.Debugf("pttImMsgImpl dispatcher GetGroupInfoFromCache error: %+v", err)
		return
	}

	onlineMem := append([]int32{}, imG.GroupOwner)

	req := &pb.ImMsgReqData{
		Id:           imU.Id,
		SenderName:   imU.NickName,
		ReceiverType: IM_MSG_FROM_UPLOAD_RECEIVER_IS_GROUP,
		ReceiverId:   imG.Gid,
		ResourcePath: pttMsg.FilePath,
		MsgType:      utils.IM_PTT_MSG,
		ReceiverName: imG.GroupName,
		SendTime:     pttMsg.Timestamp, // TODO 时间戳转换
	}

	logger.Debugf("ptt want send to :+v", onlineMem)
	resp := &pb.StreamResponse{
		DataType:  IM_MSG_FROM_UPLOAD_OR_WS_OR_APP,
		ImMsgData: req,
		Res:       &pb.Result{Code: http.StatusOK, Msg: "receiver im message successful"},
	}
	// 发送在线用户消息
	if onlineMem != nil {
		dc.Task <- *NewImTask(req.Id, onlineMem, resp)
		logger.Debugf("dispatcher to %d with %+v", req.Id, resp)
	}
}

// 实现推送的client中的dispatcherScheduler方法
func (pttImMsgImpl) dispatcherScheduler(dc *DataContext, longLived bool) {
	singleDataScheduler(dc, longLived)
}

// 分发对讲音频消息
func janusPttMsgPublish() {
	c := &Client{
		WorkType:  WORK_BY_GORONTINE, // 持续获取数据分发，所以用一个协程挂起来分发
		Dc:        NewDataContent(),
		Ds:        nil, // 因为是去redis获取
		LongLived: true,
		Cf:        pttImMsgImpl{},
	}
	c.Dc.UId = -1 // 方便日志区分single scheduler，比如-2是sos
	c.Run()
}

/**
  * web操作用来通知janus消息
  */
type webJanusImImpl struct {
	MessageType   int32   // 发送的消息类型序号
	GroupId       int32   // 群组id
	TempGroupName string  // 临时组名字
}

func (wji webJanusImImpl) dispatcher(dc *DataContext, ds DataSource) {
	var (
		id          = ds.(int32)
		errMap      = &sync.Map{}
		gpsDataResp *pb.GPSHttpResp
	)
	doNotify(&pb.DeviceInfo{Id: id}, errMap, wji.MessageType, 1,
		&pb.DeviceInfo{Id: id}, gpsDataResp, dc.Task, id, wji.GroupId, wji.TempGroupName)
}
func (webJanusImImpl) dispatcherScheduler(dc *DataContext, longLived bool) { singleDataScheduler(dc, longLived) }

// 分发web janus消息
func (sw ServerWorker) WebJanusImPublish(ctx context.Context, req int32, gId int32) (*pb.ImMsgRespData, error) {
	c := &Client{
		WorkType:  WORK_BY_NORMAL,
		Dc:        NewDataContent(),
		Ds:        req,
		LongLived: false,
		Cf: webJanusImImpl{
			GroupId:     gId,
			MessageType: sw.NotifyType,
		},
	}
	c.Run()
	logger.Debugf("# %d group update notify im once done", gId)
	return &pb.ImMsgRespData{Result: &pb.Result{Msg: "push data done", Code: 200}, MsgCode: "200"}, nil
}

// 分发web 临时组 janus消息
func (sw ServerWorker) WebJanusTempGroupImPublish(ctx context.Context, req int32, gId int32) (*pb.ImMsgRespData, error) {
	c := &Client{
		WorkType:  WORK_BY_NORMAL,
		Dc:        NewDataContent(),
		Ds:        req,
		LongLived: false,
		Cf: webJanusImImpl{
			MessageType:   sw.NotifyType,
			GroupId:       gId,
			TempGroupName: "temp group",
		},
	}
	c.Run()
	logger.Debugf("# %d temp group update notify im once done", gId)
	return &pb.ImMsgRespData{Result: &pb.Result{Msg: "push data done", Code: 200}, MsgCode: "200"}, nil
}

/**
 * 上线提醒消息 @Deprecated
 */
type loginNotifyImImpl struct{}

func (loginNotifyImImpl) dispatcher(dc *DataContext, ds DataSource) {
	NotifyToOther(dc.Task, ds.(*model.NotifyInfo).Id, ds.(*model.NotifyInfo).NotifyType)
}
func (loginNotifyImImpl) dispatcherScheduler(dc *DataContext, longLived bool) { singleDataScheduler(dc, longLived) }

// 对外暴露的上线提醒消息
func (*ServerWorker) ImLoginNotifyPublish(ctx context.Context, req *model.NotifyInfo) (*pb.ImMsgRespData, error) {
	c := &Client{
		WorkType:  WORK_BY_GORONTINE,
		Dc:        NewDataContent(),
		Ds:        req,
		LongLived: true,
		Cf:        loginNotifyImImpl{},
	}
	c.Run()
	logger.Debugf("# %d im login or logout once done", req)
	return &pb.ImMsgRespData{Result: &pb.Result{Msg: "push data done", Code: 200}}, nil
}
