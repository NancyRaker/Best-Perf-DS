/*
@Time : 2019/4/11 10:30
@Author : yanKoo
@File : file_controller
@Software: GoLand
@Description:
*/
package controllers

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"errors"
	"github.com/gin-gonic/gin"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"
	cfgWs "web-api/config"
	tfi "web-api/dao/file_info"
	"web-api/dao/user_cache"
	"web-api/engine/grpc_pool"
	"web-api/logger"
	"web-api/model"
	pb "web-api/proto/talk_cloud"
	"web-api/utils"
)

type uploadFileFunc func(fContext *model.FileContext) (interface{}, error)

// @Summary 上传文件
// @Produce  json
// @Param fileInfo formData model.ImMsgData true "上传文件携带的参数"
// @Success 200 {string} json "{"message":"User information obtained successfully",	"account_info": ai,	"device_list" deviceAll,"group_list": gList,"tree_data":resp}"
// @Router /upload [post]
func DispatcherUploadFile(c *gin.Context) {
	uploadFile(c, func(fContext *model.FileContext) (interface{}, error) {
		return grpc_pool.GrpcAppRpcCall(fContext.FileParams.Id, context.Background(), &pb.ImMsgReqData{
			Id:           int32(fContext.FileParams.Id),
			SenderName:   fContext.FileParams.SenderName,
			ReceiverType: int32(fContext.FileParams.ReceiverType),
			ReceiverId:   int32(fContext.FileParams.ReceiverId),
			ResourcePath: fContext.FilePath,
			MsgType:      fContext.FileType,
			ReceiverName: fContext.FileParams.ReceiverName,
			SendTime:     fContext.FileParams.SendTime,
			MsgCode:      strconv.FormatInt(time.Now().Unix(), 10),
		}, grpc_pool.ImMessagePublish)
	})
}

// 设备上传文件
func DeviceUploadFile(c *gin.Context) {
	uploadFile(c, func(fContext *model.FileContext) (interface{}, error) {
		// 设备采用imei号查找登录服务地址
		device, err := user_cache.GetUserFromCache(int32(fContext.FileParams.Id))
		if err != nil {
			return nil, err
		}
		return grpc_pool.GrpcAppRpcCall(int(device.AccountId), context.Background(), &pb.ImMsgReqData{
			Id:           int32(fContext.FileParams.Id),
			SenderName:   fContext.FileParams.SenderName,
			ReceiverType: int32(fContext.FileParams.ReceiverType),
			ReceiverId:   int32(fContext.FileParams.ReceiverId),
			ResourcePath: fContext.FilePath,
			MsgType:      fContext.FileType,
			ReceiverName: fContext.FileParams.ReceiverName,
			SendTime:     fContext.FileParams.SendTime,
			MsgCode:      strconv.FormatInt(time.Now().Unix(), 10),
		}, grpc_pool.ImMessagePublish)
	})
}

func uploadFile(c *gin.Context, f uploadFileFunc) {
	logger.Debugln("start upload file.")
	err := uploadFilePre(c)
	if err != nil {
		logger.Debugln("uploadFilePre error: ", err)
		c.JSON(http.StatusUnprocessableEntity, gin.H{"msg": "Uploaded File params error, please try again later.", "code": 001})
		return
	}

	// 保存文件
	fContext, err := fileStore(c)
	if err != nil {
		logger.Debugln("fileStore", err)
		c.JSON(http.StatusInternalServerError, gin.H{"msg": "Upload File fail, please try again later.", "code": 002})
		return
	}

	logger.Debugln("url: ", fContext.FilePath, "fParams: ", fContext.FileParams)

	// app上传日志文件
	if fContext.FileType == LOG_FILE_MSG {
		c.JSON(http.StatusCreated, gin.H{"msg": "Uploaded successfully"})
		return
	}

	resp, err := f(fContext)
	res := resp.(*pb.ImMsgRespData)
	if err != nil {
		logger.Error("Upload file err with Grpc : %+v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"msg": "Uploaded File, please try again later.", "code": 001})
		return
	}
	logger.Debugf("upload file success by grpc: %+v", res)

	c.JSON(http.StatusCreated, gin.H{
		"msg":          "Uploaded successfully",
		"code":         res.Result.Code,
		"MsgCode":      res.MsgCode,
		"resourcePath": fContext.FilePath,
		"resourceName": fContext.FileName,
	})
}

// 进行文件大小,存在等判断，body里面等参数的判断
func uploadFilePre(c *gin.Context) error {
	r := c.Request
	// 判断文件大小
	r.Body = http.MaxBytesReader(c.Writer, r.Body, MAX_UPLOAD_SIZE)
	if err := c.Request.ParseMultipartForm(MAX_UPLOAD_SIZE); err != nil {
		logger.Debugln("File is too big.")
		return err
	}

	// TODO 允许重复上传文件?
	return nil
}

// 文件存储
func fileStore(c *gin.Context) (*model.FileContext, error) {
	file, header, err := c.Request.FormFile("file") // TODO 会报空针
	if err != nil {
		logger.Debugln("fileStore err: ", err)
		return nil, err
	}
	uploadT := time.Now().Format(cfgWs.TimeLayout)
	// 获取上传文件所带参数
	id, _ := strconv.ParseInt(c.Request.FormValue("id"), 10, 32)
	senderName := c.Request.FormValue("SenderName")
	receiverType, _ := strconv.ParseInt(c.Request.FormValue("ReceiverType"), 10, 64)
	receiverId, _ := strconv.ParseInt(c.Request.FormValue("ReceiverId"), 10, 64)
	receiverName := c.Request.FormValue("ReceiverName")
	sTime := c.Request.FormValue("SendTime")

	//简单做一下数据判断
	if id <= 0 || receiverId <= 0 || receiverType <= 0 {
		return nil, errors.New("file param is cant be nil")
	}

	fParams := &model.ImMsgData{
		Id:           int(id),
		SenderName:   senderName,
		ReceiverType: int(receiverType),
		ReceiverId:   int(receiverId),
		ReceiverName: receiverName,
		SendTime:     sTime,
	}
	logger.Debugf("file params: %+v", fParams)

	//写入文件
	fName := strconv.FormatInt(int64(fParams.Id), 10) + "_" +
		strconv.FormatInt(time.Now().Unix(), 10) + "_" +
		header.Filename

	fSrc, err := ioutil.ReadAll(file)
	if err != nil {
		logger.Debugln("read file error: ", err)
		return nil, err
	}
	fileType, fExtName := utils.GetImFileType(header.Filename)

	// 先检验文件的hash值，避免重复上传
	md5h := md5.New()
	md5h.Write(fSrc)
	fMd5 := hex.EncodeToString(md5h.Sum([]byte("")))
	logger.Debugf("this file md5: %s\n", hex.EncodeToString(md5h.Sum([]byte("")))) //md5

	// 存储文件到fastdfs
	fileId, err := client.UploadByBuffer(fSrc, fExtName)
	if err != nil {
		logger.Debugln("UploadByBuffer to fastdfs error: ", err)
		return nil, err
	}
	logger.Debugf("file size: %d ", len(fSrc))

	fContext := &model.FileContext{
		UserId:         fParams.Id,
		FilePath:       cfgWs.FILE_BASE_URL + fileId,
		FileParams:     fParams,
		FileType:       fileType,
		FileName:       fName,
		FileSize:       len(fSrc),
		FileMD5:        fMd5,
		FileFastId:     fileId,
		FileUploadTime: uploadT,
	}

	// 记录存储到mysql
	if err := tfi.AddFileInfo(fContext); err != nil {
		logger.Debugf("Add file info to mysql error: %s", err.Error())
		return nil, err
	}

	return fContext, nil
}
