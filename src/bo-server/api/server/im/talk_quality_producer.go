/*
@Time : 2019/11/6 14:16 
@Author : yanKoo
@File : talk_quality_producer
@Software: GoLand
@Description: 通话质量的消息推送
*/
package im

import (
	pb "bo-server/api/proto"
	"bo-server/common/comon_manager"
	"bo-server/common/scheduler"
	"bo-server/dao/pub"
	tuc "bo-server/dao/user_cache"
	"bo-server/engine/cache"
	"bo-server/internal/mq/mq_receiver"
	"bo-server/logger"
	"bo-server/model"
	"net/http"
)

/**
 * 设备对讲语音的时候，通话质量消息的推送实现
 */
type talkQualityMsgPush struct {
	comon_manager.BaseManager
	workerCount         int
	talkQualityMsgQueue []*model.TalkQualityMsg
	talkQualityC        chan *model.TalkQualityMsg
}

// 从消息队列获取TalkQuality消息
func (tqmp talkQualityMsgPush) receiveTalkQualityMsg(TalkQualityD chan *model.TalkQualityMsg) {
	for {
		select {
		case msg := <-mq_receiver.TalkQualityMessage():
			TalkQualityD <- msg
		}
	}
}

// 实现推送的client中的dispatcher方法
func (tqmp talkQualityMsgPush) dispatcher(dc *DataContext, ds DataSource) {
	// 1. 从kafka获取消息
	go tqmp.receiveTalkQualityMsg(tqmp.talkQualityC)

	// 2. 运行调度器,分配对讲通话质量消息和分发处理器之间的关系
	tqmp.Scheduler.Run()

	// 3. 首先创建几个builder用来生成报告。
	for i := 0; i < tqmp.workerCount; i++ {
		tqmp.createQualityMsgWorker(dc, tqmp.Scheduler)
	}

	// 4. 创建好文件消息处理的submitter
	var submitterChan = tqmp.CreateSubmitter()

	// 5. 分发调度
	for {
		var activeChan chan interface{}
		var activeMsg *model.TalkQualityMsg
		if len(tqmp.talkQualityMsgQueue) > 0 {
			activeChan = submitterChan
			activeMsg = tqmp.talkQualityMsgQueue[0]
		}

		select {
		case t := <-tqmp.talkQualityC:
			tqmp.talkQualityMsgQueue = append(tqmp.talkQualityMsgQueue, t)
		case activeChan <- activeMsg:
			tqmp.talkQualityMsgQueue = tqmp.talkQualityMsgQueue[1:]
		}
	}
}

// TalkQuality消息分发
func (tqmp talkQualityMsgPush) createQualityMsgWorker(dc *DataContext, s comon_manager.Scheduler) {
	tc := make(chan interface{})
	go func() {
		for {
			s.WorkerReady(tc) // 告诉调度器任务空闲
			select {
			case msg := <-tc:
				qualityMsg := msg.(*model.TalkQualityMsg)
				logger.Debugf("Will send TalkQuality msg: %+v", msg)
				if qualityMsg != nil && qualityMsg.UserId > 0 {
					//tqmp.talkQualityMsgDispatcher(dc, qualityMsg)
					tqmp.saverTalkQualityMsg(qualityMsg)
				}
			}
		}
	}()
}

// 保存TalkQuality消息
func (tqmp talkQualityMsgPush) saverTalkQualityMsg(talkQualityMsg *model.TalkQualityMsg) {
	// 增加到缓存
	if err := tuc.AddUserDataInCache(int32(talkQualityMsg.UserId), []interface{}{
		pub.USER_STATUS, int32(talkQualityMsg.UserStatus),
		pub.IN_LINK_QUALITY, int32(talkQualityMsg.InLinkQuality),
		pub.IN_MEDIA_LINK_QUALITY, int32(talkQualityMsg.InMediaLinkQuality),
		pub.OUT_LINK_QUALITY, int32(talkQualityMsg.OutLinkQuality),
		pub.OUT_MEDIA_LINK_QUALITY, int32(talkQualityMsg.OutMediaLinkQuality),
		pub.RTT, int32(talkQualityMsg.Rtt),
	}, cache.GetRedisClient()); err != nil {
		logger.Error("saverTalkQualityMsg AddUserDataInCache user information to cache with error: ", err)
	}
}

// 处理TalkQuality消息
func (tqmp talkQualityMsgPush) talkQualityMsgDispatcher(dc *DataContext, talkQualityMsg *model.TalkQualityMsg) {
	imU, err := tuc.GetUserFromCache(int32(talkQualityMsg.UserId))
	if err != nil {
		logger.Debugf("talkQualityMsgPush dispatcher GetUserFromCache error: %+v", err)
	}

	qualityMsg := &pb.TalkQualityMsg{
		UserId:              int32(talkQualityMsg.UserId),
		UserStatus:          int32(talkQualityMsg.UserStatus),
		InLinkQuality:       int32(talkQualityMsg.InLinkQuality),
		InMediaLinkQuality:  int32(talkQualityMsg.InMediaLinkQuality),
		OutLinkQuality:      int32(talkQualityMsg.OutLinkQuality),
		OutMediaLinkQuality: int32(talkQualityMsg.OutMediaLinkQuality),
		Rtt:                 int32(talkQualityMsg.Rtt),
	}

	logger.Debugf("TalkQuality want send to :+v", imU.AccountId)
	resp := &pb.StreamResponse{
		DataType:   JANUS_TALK_QUALITY_MSG,
		QualityMsg: qualityMsg,
		Res:        &pb.Result{Code: http.StatusOK, Msg: "receiver im message successful"},
	}

	// 发送在线用户消息
	if imU.AccountId != 0 {
		dc.Task <- *NewImTask(qualityMsg.UserId, []int32{imU.AccountId}, resp)
		logger.Debugf("talkQualityMsgPush dispatcher to %d with %+v", qualityMsg.UserId, resp)
	}
}

// 实现推送的client中的dispatcherScheduler方法
func (talkQualityMsgPush) dispatcherScheduler(dc *DataContext, longLived bool) {
	singleDataScheduler(dc, longLived)
}

// 分发对讲音频消息
func janusTalkQualityMsgPublish() {
	c := &Client{
		WorkType:  WORK_BY_GORONTINE, // 持续获取数据分发，所以用一个协程挂起来分发
		Dc:        NewDataContent(),
		Ds:        nil, // 因为是去KAFKA获取
		LongLived: true,
		Cf: talkQualityMsgPush{
			BaseManager:  comon_manager.BaseManager{Scheduler: scheduler.NewScheduler()},
			talkQualityC: make(chan *model.TalkQualityMsg, 100),
			workerCount:  15, // 15个goroutine去处理消息
		},
	}
	c.Dc.UId = -3 // 方便日志区分single scheduler，比如-2是sos -1是ptt
	c.Run()
}
