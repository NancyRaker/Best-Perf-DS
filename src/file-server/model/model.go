/*
@Time : 2019/9/18 13:43 
@Author : yanKoo
@File : model
@Software: GoLand
@Description:
*/
package model

// MQ中的消息内容
type FileContext struct {
	UserId         int    // 用户id
	FilePath       string // 文件路径
	FileType       int32  // 文件类型
	FileName       string // 文件名字
	FileSize       int    // 文件大小
	FileMD5        string // 文件md5值
	FileFastId     string // 文件fastdfs位置
	FileUploadTime string // 文件上传时间
	FileDuration   int    // 文件时长
}
