/*
@Time : 2019/10/17 15:38 
@Author : yanKoo
@File : base_manager
@Software: GoLand
@Description:
*/
package comon_manager

type Scheduler interface {
	Submit(request interface{}) // 提交任务
	WorkerReady(w chan interface{})
	Run()
}

type BaseManager struct {
	Scheduler Scheduler
}

func (baseManager *BaseManager) CreateSubmitter() chan interface{} {
	messageC := make(chan interface{})
	go func() {
		for {
			sourceInfo := <-messageC
			baseManager.Scheduler.Submit(sourceInfo)
		}
	}()
	return messageC
}
