/*
@Time : 2019/10/17 11:10 
@Author : yanKoo
@File : report_sender
@Software: GoLand
@Description: 负责处理发送邮件任务，TODO 发送成功就把数据库标志位打上标志，发送失败就不修改标志位
*/

package report_sender

import (
	"bo-server/api/server/nfc/factory/report_builder"
	"bo-server/common/comon_manager"
	"bo-server/logger"
)

type senderManager struct {
	comon_manager.BaseManager
	reportQueue []interface{}
	workerCount int
}

var sm *senderManager

type Scheduler interface {
	Submit(request interface{}) // 提交任务
	WorkerReady(w chan interface{})
	Run()
}

func NewSenderManager(workerCount int) *senderManager {
	sm = &senderManager{

	}
	return sm
}

func (sm *senderManager) Run() {
	// 0. 运行调度器,分配管理生成报告任务和builder的关系
	sm.Scheduler.Run()

	// 1. 首先创建几个builder用来生成报告。
	for i := 0; i < sm.workerCount; i++ {
		sm.createSender(sm.Scheduler)
	}

	// 2. 创建好文件消息处理的submitter
	var submitterChan = sm.CreateSubmitter()

	// 3. 监听tagTask channel消息元数据然后分发处理
	logger.Debugln("report builder manager start running...")
	for {
		var (
			activeSubmitterChan chan interface{}
			activeSubmitContent interface{}
		)
		if len(sm.reportQueue) > 0 {
			activeSubmitterChan = submitterChan
			activeSubmitContent = sm.reportQueue[0]
		}

		select {
		case fileSource := <-report_builder.ReportSendMsg():
			sm.reportQueue = append(sm.reportQueue, fileSource)
		case activeSubmitterChan <- activeSubmitContent:
			// 让队列调度器提交了生成报告消息之后，就把engine队列中刚刚提交的消息内容出队
			sm.reportQueue = sm.reportQueue[1:]
		}
	}
}

func (sm *senderManager) createSender(s Scheduler) {
	// 为每一个sender创建一个channel
	in := make(chan interface{})
	go func() {
		sender := newSender()

		for {
			s.WorkerReady(in) // 告诉调度器任务空闲
			select {
			case request := <-in:
				sender.sendEmail(request)
			}
		}
	}()
}
