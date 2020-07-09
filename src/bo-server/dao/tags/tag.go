/*
@Time : 2019/10/14 16:58 
@Author : yanKoo
@File : tag
@Software: GoLand
@Description:
*/
package tags

import (
	pb "bo-server/api/proto"
	"bo-server/engine/db"
	"bo-server/logger"
	"database/sql"
	"errors"
)

// 增加tag标签
func SaveTagsInfo(req *pb.TagsInfoReq) error {
	if db.DBHandler == nil {
		return errors.New("db conn is nil")
	}
	stmtInsG, err := db.DBHandler.Prepare(`INSERT INTO nfc_tags (tag_name, addr, uuid, a_id, import_time) VALUES (?, ?, ?, ?, ?)`)
	if err != nil {
		return err
	}
	defer stmtInsG.Close()

	for _, tag := range req.Tags {
		insRes, err := stmtInsG.Exec(tag.TagName, tag.TagAddr, tag.Uuid, req.AccountId, tag.ImportTimestamp)
		if err != nil {
			logger.Error("Insert tag error : ", err)
			return err
		}
		if insRes != nil {
			tagId, _ := insRes.LastInsertId()
			tag.Id = int32(tagId)
			tag.AccountId = int32(req.AccountId)
		}
	}

	return nil
}

// 修改tag标签
func UpdateTagsInfo(TagsObjs *pb.TagsInfoReq) error {
	if db.DBHandler == nil {
		return errors.New("sql db conn is nil")
	}

	var (
		updStmt *sql.Stmt
		err     error
	)
	if updStmt, err = db.DBHandler.Prepare("UPDATE `nfc_tags` SET `tag_name` = ?, `addr` = ? WHERE id = ?"); err != nil {
		return err
	}
	defer updStmt.Close()

	for _, tag := range TagsObjs.Tags {
		if _, err := updStmt.Exec(tag.TagName, tag.TagAddr, tag.Id); err != nil {
			return err
		}
	}
	return nil
}

// 删除tag标签信息
func DelTagsInfo(TagsObjs *pb.TagsInfoReq) error {
	if db.DBHandler == nil {
		return errors.New("sql db conn is nil")
	}

	var (
		err     error
		delStat *sql.Stmt
	)

	delStat, err = db.DBHandler.Prepare("DELETE FROM `nfc_tags` WHERE id = ?")
	if err != nil {
		return err
	}
	defer delStat.Close()

	for _, Tags := range TagsObjs.Tags {
		if _, err = delStat.Exec(Tags.Id); err != nil {
			return err
		}
	}

	return nil
}

// 根据调度员查找tag标签
