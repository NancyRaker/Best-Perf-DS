/*
@Time : 2019/10/16 9:55 
@Author : yanKoo
@File : tag_task
@Software: GoLand
@Description: tag_task表格的增删改查
*/
package tag_task

import (
	pb "bo-server/api/proto"
	"bo-server/engine/db"
	"bo-server/logger"
	"database/sql"
	"errors"
)
// 保存标签打卡任务
func SaveTagTasksList(req *pb.TagTasksListReq) error {
	if db.DBHandler == nil {
		return errors.New("db conn is nil")
	}

	taskStmtInsG, err := db.DBHandler.Prepare(`INSERT INTO nfc_tasks(task_name, time_zone, save_time, a_id, send_email, send_time) VALUES(?, ?, ?, ?, ?, ?)`)
	if err != nil {
		return err
	}
	tagTaskStmtInsG, err := db.DBHandler.Prepare(`INSERT INTO nfc_tag_task(a_id, u_id, tag_id, order_start_time, order_end_time, task_id) VALUES(?, ?, ?, ?, ?, ?)`)
	if err != nil {
		return err
	}
	defer taskStmtInsG.Close()
	defer tagTaskStmtInsG.Close()

	for _, tagTaskList := range req.TagTaskLists {
		tagTaskList.AccountId = int32(req.AccountId)
		insGroupRes, err := taskStmtInsG.Exec(tagTaskList.TaskName, tagTaskList.TimeZone, tagTaskList.SaveTime, tagTaskList.AccountId, tagTaskList.SendEmail, tagTaskList.SendTime)
		if err != nil {
			logger.Error("Insert tag task error: ", err)
			return err
		}
		if insGroupRes != nil {
			tagTaskList.TagTaskId, _ = insGroupRes.LastInsertId()
		}

		for _, tag := range tagTaskList.TagTaskNodes {
			tag.AccountId = int32(req.AccountId)
			insGroupRes, err := tagTaskStmtInsG.Exec(tag.AccountId, tag.DeviceId, tag.TagId, tag.OrderStartTime, tag.OrderEndTime, tagTaskList.TagTaskId)
			if err != nil {
				logger.Error("Insert tag task error: ", err)
				return err
			}
			if insGroupRes != nil {
				tagId, _ := insGroupRes.LastInsertId()
				tag.Id = int32(tagId)
				tag.IsClock = 1  // 生成任务的时候默认是1
				tag.ClockTime = 1  // 打卡时间默认是1，等设备真的打开才来更新时间戳
			}
		}
	}

	return nil
}

// 删除打卡任务
func DelTagTasksList(req *pb.TagTasksListReq) error {
	if db.DBHandler == nil {
		return errors.New("sql db conn is nil")
	}

	var (
		err     error
		delStat *sql.Stmt
		delTaskStat *sql.Stmt
	)

	delStat, err = db.DBHandler.Prepare("DELETE FROM `nfc_tag_task` WHERE task_id = ?")
	if err != nil {
		return err
	}
	defer delStat.Close()

	delTaskStat, err = db.DBHandler.Prepare("DELETE FROM `nfc_tasks` WHERE id = ?")
	if err != nil {
		return err
	}
	defer delTaskStat.Close()

	for _, Tags := range req.TagTaskLists {
		if _, err = delStat.Exec(Tags.TagTaskId); err != nil {
			return err
		}
		if _, err = delTaskStat.Exec(Tags.TagTaskId); err != nil {
			return err
		}
	}

	return nil
}