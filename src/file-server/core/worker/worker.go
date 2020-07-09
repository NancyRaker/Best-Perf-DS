/*
@Time : 2019/9/18 14:24 
@Author : yanKoo
@File : worker
@Software: GoLand
@Description: 负责下载文件以及存储到fastdfs以及数据库存储
*/
package worker

import (
	"file-server/conf"
	dfi "file-server/dao/file_info"
	"file-server/logger"
	"file-server/model"
	"file-server/pub/fdfs_client"
	"file-server/utils"
	"io/ioutil"
	"net/http"
	"net/url"
	"runtime/debug"
	"strconv"
	"strings"
	"time"
)

var client *fdfs_client.Client

const FILE_MAX_SIZE = 1024 * 1025 * 2 // 2m

func init() {
	var err error
	client, err = fdfs_client.NewClientWithConfig()
	if err != nil {
		logger.Fatalf("Client: %+v NewClientWithConfig fastDFS error: %+v", client, err)
	}
	// defer client.Destory() TODO destory?
}

type storeWorker struct {
	pttMsg      *model.InterphoneMsg
	fileContext *model.FileContext
}

func NewWorker() *storeWorker {
	return &storeWorker{}
}

func (sw *storeWorker) Store(pttMsg *model.InterphoneMsg) {
	start := time.Now().UnixNano()
	logger.Debugf("sw start work: %d", start)
	sw.pttMsg = pttMsg
	sw.storeFileToFastDFS()
	sw.storeFileInfo()
	end := time.Now().UnixNano()
	logger.Debugf("sw end work: %d, used %d", end, end-start)
}

// 文件元信息存储
func (sw *storeWorker) storeFileInfo() {
	start := time.Now().UnixNano()
	logger.Debugf("sw add db file info: %d", start)
	// 记录存储到mysql
	if sw.fileContext == nil {
		return // 说明文件存储失败
	}
	if err := dfi.AddFileInfo(sw.fileContext); err != nil {
		logger.Errorf("pttImMsgImpl dispatcher Add file info to mysql error: %s", err.Error())
	}
	end := time.Now().UnixNano()
	logger.Debugf("sw end add db file info: %d, used %d", end, end-start)
}

// 文件存储到fastDFS //TODO 判断文件大小
// 1. 根据链接地址下载文件
// 2. 文件存储到fastDFS
func (sw *storeWorker) storeFileToFastDFS() {
	start := time.Now().UnixNano()
	logger.Debugf("sw store get file: %d", start)

	logger.Debugf("storeFileToFastDFS start store file with:%s", sw.pttMsg.FilePath)
	uri, err := url.ParseRequestURI(sw.pttMsg.FilePath)
	if err != nil {
		logger.Errorf("storeFileToFastDFS ParseRequestURI %s, err:%+v", sw.pttMsg.FilePath, err)
		return
	}

	fileExtName := uri.Path[strings.LastIndex(uri.Path, ".")+1:]
	logger.Debugf("storeFileToFastDFS Filename " + fileExtName)

	cli := http.DefaultClient
	cli.Timeout = time.Second * 60 //设置超时时间
	resp, err := cli.Get(sw.pttMsg.FilePath)
	if err != nil {
		logger.Errorf("storeFileToFastDFS cat send GET http.")
		return
	}
	if resp.ContentLength <= 0 {
		logger.Errorf("storeFileToFastDFS server does not support breakpoint download.")
		return
	}
	if resp.Body == nil {
		return
	}
	defer func() {
		debug.FreeOSMemory()
		resp.Body.Close()
	}()
	buffer, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		logger.Errorf("storeFileToFastDFS ReadAll error: %+v", err)
		return
	}
	logger.Debugln(len(buffer), cap(buffer))

	endgetFile := time.Now().UnixNano()
	logger.Debugf("sw end get file: %d, used %d us to get file", endgetFile, endgetFile-start)

	// 存储文件到fastdfs
	fileId, err := client.UploadByBuffer(buffer, fileExtName)
	if err != nil {
		logger.Errorf("storeFileToFastDFS UploadByBuffer to fastdfs error: %+v", err)
		return
	}
	logger.Debugln(conf.FILE_BASE_URL + fileId)
	sw.pttMsg.FilePath = conf.FILE_BASE_URL + fileId

	uId, _ := strconv.ParseInt(sw.pttMsg.Uid, 10, 64)
	fDuration, _ := strconv.Atoi(sw.pttMsg.Duration)
	sw.fileContext = &model.FileContext{
		UserId:         int(uId),
		FilePath:       sw.pttMsg.FilePath,
		FileType:       utils.IM_PTT_MSG,
		FileName:       sw.pttMsg.FilePath,
		FileSize:       len(buffer), // 文件大小 TODO janus没有返回
		FileDuration:   fDuration,
		FileMD5:        utils.GetMd5(buffer),
		FileFastId:     fileId,
		FileUploadTime: time.Now().Format(conf.TimeLayout),
	}

	endfastdfsFile := time.Now().UnixNano()
	logger.Debugf("sw end fastdfs file: %d, used %d us to upload file", endfastdfsFile, endfastdfsFile-endgetFile)
}
