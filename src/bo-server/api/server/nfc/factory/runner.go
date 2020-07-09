/*
@Time : 2019/10/17 11:08 
@Author : yanKoo
@File : runner
@Software: GoLand
@Description:
*/
package factory

import (
	"bo-server/api/server/nfc/factory/report_builder"
	"bo-server/api/server/nfc/factory/report_sender"
	"bo-server/api/server/nfc/factory/report_task_generator"
)

const (
	tagTaskChanSize    = 50 // 获取发送任务的channel队列大小
	reportBuilderCount = 10 // 报告生成工作者的数量
	reportChanSize     = 50 // 报告channel队列大小
	reportSenderCount  = 10 // 报告发送邮差的数量
)

// 启动NFC打卡状态上报任务
func NFCFactoryInit() {
	// 1. 启动任务获取器
	go report_task_generator.NewTimeTaskGenerator(tagTaskChanSize).Run()

	// 2. 启动报告生成器
	go report_builder.NewReportBuilderManager(reportBuilderCount, reportChanSize).Run()

	// 3. 启动邮差发送器
	go report_sender.NewSenderManager(reportSenderCount).Run()
}
