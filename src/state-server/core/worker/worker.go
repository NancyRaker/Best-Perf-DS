/*
@Time : 2019/9/18 14:24 
@Author : yanKoo
@File : worker
@Software: GoLand
@Description: 负责下载文件以及存储到fastdfs以及数据库存储
*/
package worker

import (
	"state-server/dao/user_cache"
	"state-server/logger"
	"state-server/model"
	"time"
)

const (
	KAFKA_User_STATE_MSG = 1 // 接收到的kafka消息是状态转换消息

	USER_SLEEP = 1 // 设备更新休眠状态
)

type stateModifyWorker struct {
	sleepingInfo *model.SleepingInfo
}

func NewWorker() *stateModifyWorker {
	return &stateModifyWorker{}
}

func (smw *stateModifyWorker) Store(stateMessage *model.StateMessage) {
	start := time.Now().UnixNano()
	logger.Debugf("smw start work: %d", start)
	switch stateMessage.MsgType {
	case KAFKA_User_STATE_MSG:
		smw.modifyUserStateInfo(stateMessage.UserState)
	}
	end := time.Now().UnixNano()
	logger.Debugf("smw end work: %d, used %d", end, end-start)
}

// 更新设备状态信息
func (smw *stateModifyWorker) modifyUserStateInfo(message *model.UserState) {
	if message == nil {
		return
	}

	var valuePairs []interface{}
	switch message.StateCode {
	case USER_SLEEP:
		valuePairs = append(valuePairs, user_cache.SLEEP_STATE, message.SleepingInfo.State)
	}

	_ = user_cache.ModifyUserState(int32(message.SleepingInfo.Id), valuePairs)
}
