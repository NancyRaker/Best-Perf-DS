/*
@Time : 2019/9/18 14:23 
@Author : yanKoo
@File : concurrentEngine
@Software: GoLand
@Description: 启动文件存储服务，从全局队列里面获取数据然后由任务调度器分配空闲的worker来处理
*/
package engine

import (
	"file-server/core/listener"
	"file-server/core/worker"
	"file-server/logger"
	"file-server/model"
)

type concurrentEngine struct {
	scheduler           Scheduler
	workerCount         int
	fileSourceInfos     chan *model.InterphoneMsg
	fileSourceInfoQueue []*model.InterphoneMsg
}

type Scheduler interface {
	Submit(request *model.InterphoneMsg) // 提交任务
	//ConfigMasterWorkerChan(chan model.FileSourceInfo)
	WorkerReady(w chan *model.InterphoneMsg)
	Run()
}

func NewEngine(scheduler Scheduler, workerCount int, fileSourceInfoSize int) *concurrentEngine {
	return &concurrentEngine{
		scheduler:       scheduler,
		workerCount:     workerCount,
		fileSourceInfos: make(chan *model.InterphoneMsg, fileSourceInfoSize),
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

	// 3. 创建好文件消息处理的submitter
	var submitterChan = e.createSubmitter()

	// 3. 监听消息队列中的消息
	go listener.NewFileListener(e.fileSourceInfos).Listen()

	// 4. 从engine队列中获取ptt消息元数据然后分发处理
	logger.Debugln("file storage server start running...")
	for {
		var (
			activeSubmitterChan chan *model.InterphoneMsg
			activeSubmitContent *model.InterphoneMsg
		)
		if len(e.fileSourceInfoQueue) > 0 {
			activeSubmitterChan = submitterChan
			activeSubmitContent = e.fileSourceInfoQueue[0]
		}

		select {
		case fileSource := <-e.fileSourceInfos:
			e.fileSourceInfoQueue = append(e.fileSourceInfoQueue, fileSource)
		case activeSubmitterChan <- activeSubmitContent:
			// 让队列调度器提交了文件消息之后，就把engine队列中刚刚提交的消息内容出队
			e.fileSourceInfoQueue = e.fileSourceInfoQueue[1:]
		}
	}
}

func (e *concurrentEngine) createSubmitter() chan *model.InterphoneMsg {
	messageC := make(chan *model.InterphoneMsg)
	go func() {
		for {
			fileSourceInfo := <-messageC
			e.scheduler.Submit(fileSourceInfo)
		}
	}()
	return messageC
}

// 创建工作协程
func createWorker(s Scheduler) {
	// 为每一个Worker创建一个channel
	in := make(chan *model.InterphoneMsg)
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
