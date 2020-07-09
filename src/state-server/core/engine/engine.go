/*
@Time : 2019/9/18 14:23 
@Author : yanKoo
@File : concurrentEngine
@Software: GoLand
@Description: 启动文件存储服务，从全局队列里面获取数据然后由任务调度器分配空闲的worker来处理
*/
package engine

import (
	"state-server/core/listener"
	"state-server/core/worker"
	"state-server/logger"
	"state-server/model"
)

type concurrentEngine struct {
	scheduler           Scheduler
	workerCount         int
	stateSourceInfos     chan *model.StateMessage
	stateSourceInfoQueue []*model.StateMessage
}

type Scheduler interface {
	Submit(request *model.StateMessage) // 提交任务
	//ConfigMasterWorkerChan(chan model.stateSourceInfo)
	WorkerReady(w chan *model.StateMessage)
	Run()
}

func NewEngine(scheduler Scheduler, workerCount int) *concurrentEngine {
	return &concurrentEngine{
		scheduler:       scheduler,
		workerCount:     workerCount,
		stateSourceInfos: make(chan *model.StateMessage, 32),
	}
}

// 启动文件存储服务
func (e *concurrentEngine) Run() {
	// 1. 挂起engine的队列调度器
	e.scheduler.Run()

	// 2. 创建好工作worker goroutine
	for i := 0; i < e.workerCount; i++ {
		createWorker(e.scheduler)
	}

	// 3. 创建好状态消息处理的submitter
	var submitterChan = e.createSubmitter()

	// 3. 监听消息队列中的消息
	go listener.NewStateMsgListener(e.stateSourceInfos).Listen()

	// 4. 从engine队列中获取消息数据然后分发处理
	logger.Debugln("file storage server start running...")
	for {
		var (
			activeSubmitterChan chan *model.StateMessage
			activeSubmitContent *model.StateMessage
		)
		if len(e.stateSourceInfoQueue) > 0 {
			activeSubmitterChan = submitterChan
			activeSubmitContent = e.stateSourceInfoQueue[0]
		}

		select {
		case stateSource := <-e.stateSourceInfos:
			e.stateSourceInfoQueue = append(e.stateSourceInfoQueue, stateSource)
		case activeSubmitterChan <- activeSubmitContent:
			// 让队列调度器提交了状态消息之后，就把engine队列中刚刚提交的消息内容出队
			e.stateSourceInfoQueue = e.stateSourceInfoQueue[1:]
		}
	}
}

func (e *concurrentEngine) createSubmitter() chan *model.StateMessage {
	messageC := make(chan *model.StateMessage)
	go func() {
		for {
			stateSourceInfo := <-messageC
			e.scheduler.Submit(stateSourceInfo)
		}
	}()
	return messageC
}

// 创建工作协程
func createWorker(s Scheduler) {
	// 为每一个Worker创建一个channel
	in := make(chan *model.StateMessage)
	go func() {
		for {
			s.WorkerReady(in) // 告诉调度器任务空闲
			select {
			case request := <-in:
				w := worker.NewWorker()
				w.Store(request)
			}
		}
	}()
}
