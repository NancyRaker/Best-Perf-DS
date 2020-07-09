/*
@Time : 2019/10/17 11:03 
@Author : yanKoo
@File : report_task_runner
@Software: GoLand
@Description: 负责从mysql数据库中获取需要发送报告的任务，然后通知report_builder生成报告，然后由report_sender投递出去
*/
package report_task_generator

import (
	pb "bo-server/api/proto"
)

type timeTaskGenerator struct {
	tagTaskChan chan *pb.TagTaskListNode
}

var timeTask *timeTaskGenerator

// 生成定时任务获取器
func NewTimeTaskGenerator(tagTaskChanSize int) *timeTaskGenerator {
	timeTask = &timeTaskGenerator{tagTaskChan: make(chan *pb.TagTaskListNode, tagTaskChanSize)}
	return timeTask
}

// 启动定时任务
func (ttg *timeTaskGenerator) Run() {
	go ttg.readTask()
}

// 暴露给report_builder获取任务
func TagTaskMsg() <-chan *pb.TagTaskListNode {
	return timeTask.tagTaskChan
}

// 从mysql中读取任务 TODO
func (ttg *timeTaskGenerator) readTask() {
	//
}
