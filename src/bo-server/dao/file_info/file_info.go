/*
@Time : 2019/4/28 19:51
@Author : yanKoo
@File : file_info
@Software: GoLand
@Description:
*/
package file_info

import (
	"bo-server/engine/db"
	"bo-server/model"
)

const (
	IM_TEXT_MSG         = 1 // 普通文本
	IM_IMAGE_MSG        = 2 // 图片
	IM_VOICE_MSG        = 3 // 音频文件
	IM_VIDEO_MSG        = 4 // 视频文件
	IM_PTT_MSG          = 5 // ptt音频文件
	LOG_FILE_MSG        = 6 // 设备日志文件
	APP_FILE            = 9999 // 设备升级文件
	IM_UNKNOWN_TYPE_MSG = 10000
)

// 增加文件信息
func AddFileInfo(fc *model.FileContext) error {
	stmtIns, err := db.DBHandler.Prepare("INSERT INTO file_info (uid, f_name, f_size, f_upload_t, f_mdf, fdfs_id, f_type) VALUES (?, ?, ?, ?, ?, ?, ?) ")
	if err != nil {
		return err
	}
	defer stmtIns.Close()

	if _, err := stmtIns.Exec(fc.UserId, fc.FileName, fc.FileSize, fc.FileUploadTime, fc.FileMD5, fc.FileFastId, fc.FileType); err != nil {
		return err
	}

	return nil
}

// 获取文件信息
func GetFileInfo(uId int32) (*model.FileContext, error) {
	stmtOut, err := db.DBHandler.Prepare("SELECT f_name, f_size, f_upload_t, f_mdf, fdfs_id FROM file_info WHERE uid = ? AND f_type = ? ORDER BY create_time DESC LIMIT 1")
	if err != nil {
		return nil, err
	}
	defer stmtOut.Close()

	fc := &model.FileContext{}
	if err = stmtOut.QueryRow(uId, LOG_FILE_MSG).Scan(&fc.FileName, &fc.FileSize, &fc.FileUploadTime, &fc.FileMD5, &fc.FileFastId); err != nil {
		return nil, err
	}

	return fc, nil
}

// 获取文件信息
func GetApkFileInfo(uId int32) (*model.FileContext, error) {
	stmtOut, err := db.DBHandler.Prepare("SELECT f_name, f_size, f_upload_t, f_mdf, fdfs_id FROM file_info WHERE uid = ? AND f_type = ? ORDER BY create_time DESC LIMIT 1")
	if err != nil {
		return nil, err
	}
	defer stmtOut.Close()

	fc := &model.FileContext{}
	if err = stmtOut.QueryRow(uId, APP_FILE).Scan(&fc.FileName, &fc.FileSize, &fc.FileUploadTime, &fc.FileMD5, &fc.FileFastId); err != nil {
		return nil, err
	}

	return fc, nil
}
