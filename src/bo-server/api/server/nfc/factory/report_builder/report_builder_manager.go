/*
@Time : 2019/10/17 11:10 
@Author : yanKoo
@File : report_builder
@Software: GoLand
@Description:
*/
package report_builder

import (
	"bo-server/api/server/nfc/factory/report_task_generator"
	"bo-server/common/comon_manager"
	"bo-server/common/scheduler"
	"bo-server/logger"
)

type reportBuilderManager struct {
	comon_manager.BaseManager
	workerCount    int
	tagTaskQueue   []interface{}
	reportSendChan chan interface{} // TODO 报告任务发送内容，还没定
}

var bm *reportBuilderManager

func NewReportBuilderManager(workerCount int, bufChanSize int) *reportBuilderManager {
	bm = &reportBuilderManager{
		BaseManager:    comon_manager.BaseManager{Scheduler: scheduler.NewScheduler()},
		workerCount:    workerCount,
		reportSendChan: make(chan interface{}, bufChanSize),
	}
	return bm
}

func ReportSendMsg() <-chan interface{} {
	return bm.reportSendChan
}

func (rbm *reportBuilderManager) Run() {
	// 0. 运行调度器,分配管理生成报告任务和builder的关系
	rbm.Scheduler.Run()

	// 1. 首先创建几个builder用来生成报告。
	for i := 0; i < rbm.workerCount; i++ {
		rbm.createBuilder(rbm.Scheduler)
	}

	// 2. 创建好文件消息处理的submitter
	var submitterChan = rbm.CreateSubmitter()

	// 3. 监听tagTask channel消息元数据然后分发处理
	logger.Debugln("report builder manager start running...")
	for {
		var (
			activeSubmitterChan chan interface{}
			activeSubmitContent interface{}
		)
		if len(rbm.tagTaskQueue) > 0 {
			activeSubmitterChan = submitterChan
			activeSubmitContent = rbm.tagTaskQueue[0]
		}

		select {
		case fileSource := <-report_task_generator.TagTaskMsg():
			rbm.tagTaskQueue = append(rbm.tagTaskQueue, fileSource)
		case activeSubmitterChan <- activeSubmitContent:
			// 让队列调度器提交了生成报告消息之后，就把engine队列中刚刚提交的消息内容出队
			rbm.tagTaskQueue = rbm.tagTaskQueue[1:]
		}
	}
}

// 创建报告生成器
func (rbm *reportBuilderManager) createBuilder(s comon_manager.Scheduler) {
	// 为每一个Worker创建一个channel
	in := make(chan interface{})
	go func() {
		// 报告生成器
		b := newBuilder()

		for {
			s.WorkerReady(in) // 告诉调度器任务空闲
			select {
			case reportTask := <-in:
				rbm.reportSendChan <- b.generateReport(reportTask)
			}
		}
	}()
}
