/*
@Time : 2019/7/30 14:44 
@Author : yanKoo
@File : im_consumer
@Software: GoLand
@Description: 到comet获取连接，然后发送消息
*/

package im

import (
	pb "bo-server/api/proto"
	tm "bo-server/dao/msg"
	"bo-server/engine/db"
	"bo-server/logger"
	"fmt"
	"github.com/panjf2000/ants"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"sync"
	"time"
)

// 暂时使用一个channel来当全局消息队列
var GlobalTaskQueue = struct {
	Tasks chan Task
	m     sync.Mutex
}{Tasks: make(chan Task, 1000)}

// 调度从全局chan里取出的task
func executorScheduler() {
	var tasks []Task
	var executor = createExecutor()
	tick := time.NewTicker(time.Second * time.Duration(60))
	for {
		var activeExecutor chan Task
		var activeTask Task
		if len(tasks) > 0 {
			activeExecutor = executor
			activeTask = tasks[0]
		}
		select {
		case t := <-GlobalTaskQueue.Tasks:
			tasks = append(tasks, t)
		case activeExecutor <- activeTask:
			tasks = tasks[1:]
		case <-tick.C:
			logger.Debugf("now task queue len:%d", len(tasks))
		}
	}
}

// 创建executor负责分发消息
func createExecutor() chan Task {
	fmt.Printf("will create executor")
	tc := make(chan Task)
	go pushDataExecutor(tc)
	return tc
}

// Executor 推送、IM离线数据、IM离线数据
func pushDataExecutor(ct chan Task) {
	p, _ := ants.NewPool(200) // 协程池配置两百个协程
	for {
		select {
		case task := <-ct:
			_ = p.Submit(func() {
				var wg sync.WaitGroup
				logger.Debugf("executor pool work receiver: %+v", task.Receiver)
				for _, receiverId := range task.Receiver {
					go pushData(task, receiverId, task.Data, &wg)
					logger.Debugf("will send to %d", receiverId)
				}
			})
		}
	}
}

// 推送数据
func pushData(task Task, receiverId int32, resp *pb.StreamResponse, wg *sync.WaitGroup) {
	//defer wg.Done()
	//logger.Debugf("the stream map have: %p, %+v", &GoroutineMap, GoroutineMap)
	if value := GoroutineMap.GetStream(receiverId); value != nil {
		srv := value.(pb.TalkCloud_DataPublishServer)
		logger.Debugf("# %d receiver response: %+v", receiverId, resp)
		if err := srv.Send(resp); err != nil {
			// 发送失败处理 mysql的io速度太慢了
			processErrorSendMsg(err, task, receiverId, resp)
		} else {
			// 发送成功如果是离线数据（接收者等于stream id自己）就去更新状态
			logger.Debugf("send success. dc.senderId: %d, receiverId: %d", task.SenderId, receiverId)
			if task.SenderId == receiverId && resp.DataType == OFFLINE_IM_MSG && resp.OfflineImMsgResp != nil &&
				!(len(resp.OfflineImMsgResp.OfflineSingleImMsgs) == 0 &&
					len(resp.OfflineImMsgResp.OfflineGroupPttImMsgs) == 0 &&
					len(resp.OfflineImMsgResp.OfflineGroupImMsgs) == 0) {
				//  更新数据库里面的消息的状态
				if err := tm.SetMsgStat(receiverId, READ_OFFLINE_IM_MSG, db.DBHandler); err != nil {
					logger.Error("Add offline msg with error: ", err)
				}
			}
		}
	} else {
		logger.Debugf("Send to %d im that can't find stream", receiverId) //TODO 就依靠那边心跳了，这里就不管了
		// 存储即时发送失败的消息
		if resp.DataType == IM_MSG_FROM_UPLOAD_OR_WS_OR_APP && task.SenderId != receiverId && resp.ImMsgData.MsgType <= tm.IM_PTT_MSG {
			// 把发送数据保存进数据库, 如果是离线数据就忽略
			logger.Debugf("send fail. dc.senderId: %d, receiverId: %d", task.SenderId, receiverId)
			if err := tm.AddMsg(resp.ImMsgData, db.DBHandler); err != nil {
				logger.Errorf("Send fail and add offline msg with error: ", err)
			}
		}
	}
}

// 处理推送数据失败的情况
func processErrorSendMsg(err error, task Task, receiverId int32, resp *pb.StreamResponse) {
	logger.Debugf("send msg fail with error: %+v", err)

	// 判断错误类型
	if errSC, _ := status.FromError(err); errSC.Code() == codes.Unavailable || errSC.Code() == codes.Canceled {
		logger.Debugf("%+v ---- %+v start add offline msg: ", resp, task)
		if resp.DataType == IM_MSG_FROM_UPLOAD_OR_WS_OR_APP && task.SenderId != receiverId {
			// 把发送数据保存进数据库, 如果是离线数据就忽略
			logger.Debugf("send fail. dc.senderId: %d, receiverId: %d", task.SenderId, receiverId)
			if err := tm.AddMsg(resp.ImMsgData, db.DBHandler); err != nil {
				logger.Errorf("Send fail and add offline msg with error: ", err)
			}
		}
	}
}
