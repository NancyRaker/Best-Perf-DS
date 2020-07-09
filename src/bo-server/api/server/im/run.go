/*
@Time : 2019/7/30 15:05 
@Author : yanKoo
@File : engine
@Software: GoLand
@Description:im模块启动。启动的时候首先要把消息消费者（协程池）和ptt消息监听协程以及检查stream和redis之间的数据同步的协程运行起来
*/
package im

import (
	"bo-server/internal/event_hub"
)

type imEngine struct{}

func init() {
	imEngine{}.Run()
}

// 启动im的一系列需要挂起的任务：executor pool、ptt dispatcher
func (ie imEngine) Run() {
	// 1. 消息推送模块 global executor
	go executorScheduler()

	// 2. 根据时间间隔，检查stream map和redis
	//go syncStreamWithRedis(cfgGs.Interval)  // TODO 建议修改为redis的pub和sub的模式，自动监听redis过期的key
	go syncStreamWithRedis()

	// 3. 启动信号管理器统一管理janus发送过来的非对讲的操作回复消息
	event_hub.SignalManagerInit()

	// 4. redis持续获取im数据，dispatcher
	janusPttMsgPublish() // 修改架构结构之后的ptt消息处理
	/*janusPttMsgPublish()*/

	// 5. 通话质量类消息
	janusTalkQualityMsgPublish()
}
