/*
@Time : 2019/5/9 14:33
@Author : yanKoo
@File : im_file
@Software: GoLand
@Description:
*/
package utils

import (
	"strings"
	"sync"
	"web-api/logger"
)

var imFileMap sync.Map

const (
	IM_TEXT_MSG         = 1 // 普通文本
	IM_IMAGE_MSG        = 2 // 图片
	IM_VOICE_MSG        = 3 // 音频文件
	IM_VIDEO_MSG        = 4 // 视频文件
	IM_PTT_MSG          = 5 // ptt音频文件
	LOG_FILE_MSG        = 6 // 设备日志文件
	IM_UNKNOWN_TYPE_MSG = 10000
)

func init() {
	imFileMap.Store("jpg", IM_IMAGE_MSG)  // (jpg)
	imFileMap.Store("JPEG", IM_IMAGE_MSG) //JPEG (jpeg)
	imFileMap.Store("jpeg", IM_IMAGE_MSG) //JPEG (jpeg)
	imFileMap.Store("png", IM_IMAGE_MSG)  //PNG (png)
	imFileMap.Store("gif", IM_IMAGE_MSG)  //GIF (gif)
	imFileMap.Store("tif", IM_IMAGE_MSG)  //TIFF (tif)
	imFileMap.Store("bmp", IM_IMAGE_MSG)  // (bmp)

	imFileMap.Store("rmvb", IM_VIDEO_MSG) //rmvb/rm相同
	imFileMap.Store("flv", IM_VIDEO_MSG)  //flv与f4v相同
	imFileMap.Store("mp4", IM_VIDEO_MSG)
	imFileMap.Store("3gp", IM_VIDEO_MSG)
	imFileMap.Store("mpg", IM_VIDEO_MSG) //
	imFileMap.Store("wmv", IM_VIDEO_MSG) //wmv与asf相同

	imFileMap.Store("mp3", IM_VOICE_MSG)
	imFileMap.Store("wav", IM_VOICE_MSG) //Wave (wav)
	imFileMap.Store("avi", IM_VOICE_MSG)
	imFileMap.Store("mid", IM_VOICE_MSG) //MIDI (mid)

	imFileMap.Store("log", LOG_FILE_MSG) //app日志文件
}

// 获取Im上传文件类型
func GetImFileType(headerFileHeader string) (int32, string) {
	// 判断文件类型
	//fType := utils.GetFileType((*fSrc)[:10])

	fNameStr := strings.Split(headerFileHeader, ".")
	fType := fNameStr[len(fNameStr)-1]

	logger.Debugln("get file fType: ", fType)
	var fileType int32 = IM_UNKNOWN_TYPE_MSG
	imFileMap.Range(func(key, value interface{}) bool {
		if key.(string) == fType {
			logger.Debugln("find file type:", value)
			fileType = int32(value.(int))
			return false
		}
		return true
	})
	return fileType, fType
}
