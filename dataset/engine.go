/*
@Time : 2019/12/4 15:02 
@Author : yanKoo
@File : engine
@Software: GoLand
@Description:
*/
package dataset

import (
	"model"
	"fmt"
	"log"
	"strconv"
	"sync/atomic"
)

type concurrentEngine struct {
	scheduler        Scheduler
	counter          chan int32
	workerCount      int
	companyListNodes []*model.CompanyListNode
	finishNodes      []*model.CompanyListNode
}

type Scheduler interface {
	Submit(request *model.WorkTaskNode) // 提交任务
	WorkerReady(w chan *model.WorkTaskNode)
	Run()
}

func NewEngine(scheduler Scheduler, workerCount int) *concurrentEngine {
	e := &concurrentEngine{
		scheduler:        scheduler,
		workerCount:      workerCount,
		counter:          make(chan int32, 65535),
		companyListNodes: ReadQuerySource(),
	}

	e.finishNodes = make([]*model.CompanyListNode, len(e.companyListNodes))

	return e
}

const BasePicSavePath = "codepic"
const BasePicCodeResultFile = "codeRusult"

var counter int32
// 启动爬虫
func (e *concurrentEngine) Run() {

	total := int32(len(e.companyListNodes))

	log.Printf("finsh with %d", total)

	// 1. 挂起engine的队列调度器
	e.scheduler.Run()

	// 2. 创建好工作worker goroutine
	for i := 0; i < e.workerCount; i++ {
		createWorker(BasePicSavePath+strconv.Itoa(i)+".jpg",
			BasePicCodeResultFile+strconv.Itoa(i)+".txt", e)
	}

	// 3. 创建好文件消息处理的submitter
	var submitterChan = e.createSubmitter()

	// 4. 分发处理
	log.Println("parser server start running...")
	for {
		var (
			activeSubmitterChan chan *model.WorkTaskNode
			activeSubmitContent *model.WorkTaskNode
		)
		if len(e.companyListNodes) > 0 {
			activeSubmitterChan = submitterChan
			activeSubmitContent = &model.WorkTaskNode{
				Index:           len(e.companyListNodes) - 1,
				CompanyListNode: e.companyListNodes[len(e.companyListNodes)-1],
			}
		}

		select {
		case c := <-e.counter:
			if total == c {
				log.Printf("finsh with %d", total)
				SaveCompanyInfo(e.finishNodes)
				return
			}
		case activeSubmitterChan <- activeSubmitContent:
			// 从后面取
			e.companyListNodes = e.companyListNodes[:len(e.companyListNodes)-1]
		}

	}

}

func (e *concurrentEngine) createSubmitter() chan *model.WorkTaskNode {
	messageC := make(chan *model.WorkTaskNode)
	go func() {
		for {
			fileSourceInfo := <-messageC
			e.scheduler.Submit(fileSourceInfo)
		}
	}()
	return messageC
}

// 创建工作协程
func createWorker(picSavePath, picCodeResultFile string, engine *concurrentEngine) {
	// 为每一个Worker创建一个channel
	in := make(chan *model.WorkTaskNode)
	worker := NewQueryParser(picSavePath, picCodeResultFile)
	go func(e *concurrentEngine) {
		for {
			e.scheduler.WorkerReady(in) // 告诉调度器任务空闲
			select {
			case task := <-in:
				offset := worker.StartQuery(task.Index, task.CompanyListNode, 0)
				if offset >= 0 {
					e.finishNodes[offset] = task.CompanyListNode
				}
			}
			atomic.AddInt32(&counter, 1)
			e.counter <- atomic.LoadInt32(&counter)
			fmt.Println(atomic.LoadInt32(&counter))
		}
	}(engine)
}
