/*
@Time : 2019/6/20 10:22 
@Author : yanKoo
@File : Tags
@Software: GoLand
@Description:
*/
package controllers

import (
	"context"
	"errors"
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
	"web-api/engine/grpc_pool"
	"web-api/logger"
	"web-api/model"
	pb "web-api/proto/talk_cloud"
	"web-api/service"
)

var (
	SAVE_TAG_INFO   int32 = 1
	UPDATE_TAG_INFO int32 = 2
	DELETE_TAG_INFO int32 = 3
)

//
// @Summary 增加，删除，修改Tags信息
// @Produce  json
// @Param Authorization header string true "登录时返回的sessionId"
// @Param accountId path string true "当前用户的账号Id"
// @Param Body body model.SwagTagsInfoReq true "获取Tags的model"
// @Success 200 {object} model.SwagGetDeviceLocationResp "正确处理返回的数据"
// @Router /tags/{accountId} [post]
func OperationTagsInfo(c *gin.Context) {
	aidStr := c.Param("accountId")
	accountTagsInfo := &pb.TagsInfoReq{}
	if err := c.BindJSON(accountTagsInfo); err != nil {
		logger.Debugf("json parse fail , error : %s", err)
		c.JSON(http.StatusBadRequest, model.ErrorRequestBodyParseFailed)
		return
	}

	// 校验ops参数
	if accountTagsInfo.Ops < 1 || accountTagsInfo.Ops > 3 {
		logger.Errorf("accountTags Info ops error: %d", accountTagsInfo.Ops)
		c.JSON(http.StatusBadRequest, model.ErrorRequestBodyParseFailed)
		return
	}

	// 使用session来校验用户
	aid, _ := strconv.Atoi(aidStr)
	if !service.ValidateAccountSession(c.Request, aid) {
		c.JSON(http.StatusUnauthorized, model.ErrorNotAuthSession)
		return
	}
	accountTagsInfo.AccountId = int32(aid)
	logger.Errorf("accountTags Info : %+v", accountTagsInfo)
	switch accountTagsInfo.Ops {
	case SAVE_TAG_INFO:
		saveTagsInfo(c, accountTagsInfo)
	case UPDATE_TAG_INFO:
		updateTagsInfo(c, accountTagsInfo)
	case DELETE_TAG_INFO:
		deleteTagsInfo(c, accountTagsInfo)
	}
}

// 保存Tags信息
func saveTagsInfo(c *gin.Context, TagsInfos *pb.TagsInfoReq) {
	var err error
	// 校验数据
	for _, tag := range TagsInfos.Tags {
		// 校验位置信息
		if tag.TagName == "" || tag.TagAddr == "" || tag.ImportTimestamp <= 0 || tag.Uuid == ""{
			logger.Debugf("tag params:%+v", tag)
			err = errors.New("tag name or addr or timestamp or uuid is invalid")
			break
		}
	}

	if err != nil {
		logger.Debugf("saveTagsInfo validate tag info fail with error: %+v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error msg": err.Error()})
		return
	}

	TagsInfoProcess(c, TagsInfos)
}

// 修改Tags信息
func updateTagsInfo(c *gin.Context, TagsInfos *pb.TagsInfoReq) {
	var err error
	// 校验数据
	for _, tag := range TagsInfos.Tags {
		// 校验位置信息
		if tag.TagName == "" || tag.TagAddr == "" {
			err = errors.New("tag name or addr is invalid")
			break
		}
	}

	if err != nil {
		logger.Debugf("validate Tags info fail with error: %+v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error msg": err.Error()})
		return
	}

	TagsInfoProcess(c, TagsInfos)
}

// 删除Tags信息
func deleteTagsInfo(c *gin.Context, TagsInfos *pb.TagsInfoReq) {
	var err error
	// 校验数据
	for _, tag := range TagsInfos.Tags {
		// 校验位置信息
		if tag.TagName == "" || tag.TagAddr == "" {
			err = errors.New("tag name or addr is invalid")
			break
		}
	}

	if err != nil {
		logger.Debugf("validate Tags info fail with error: %+v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error msg": err.Error()})
		return
	}

	TagsInfoProcess(c, TagsInfos)
}

// grpc调用，提交信息
func TagsInfoProcess(c *gin.Context, TagsInfos *pb.TagsInfoReq) {
	var (
		res interface{} //*pb.TagsInfoResp
		err error
	)
	aidStr := c.Param("accountId")
	aid, _ := strconv.Atoi(aidStr)
	if res, err = grpc_pool.GrpcWebRpcCall(aid, context.TODO(), TagsInfos, grpc_pool.PostTagsInfo); err != nil {
		//if res, err = webCli.PostTagsInfo(context.TODO(), TagsInfos); err != nil {
		logger.Errorf("operate Tags info by grpc with error: %+v", err)
		c.JSON(http.StatusInternalServerError, model.ErrorInternalServerError)
		return
	}

	if res.(*pb.TagsInfoResp).Res.Code != http.StatusOK {
		c.JSON(http.StatusBadRequest, gin.H{"msg": res.(*pb.TagsInfoResp).Res.Msg})
		return
	}

	c.JSON(http.StatusOK, model.SuccessResponse)
}

// @Summary 获取Tags信息
// @Produce  json
// @Param accountId path string true "当前用户的账号Id"
// @Param Authorization header string true "登录时返回的sessionId"
// @Success 200 {string} json "	{"msg":"delete account successfully.","result": "success"}"
// @Router /tags/{accountId} [get]
func GetTagsInfo(c *gin.Context) {
	var (
		err    error
		aidStr = c.Param("accountId")
		tags   interface{} // *pb.GetTagsInfoResp
	)
	// 使用session来校验用户
	aid, _ := strconv.Atoi(aidStr)
	if !service.ValidateAccountSession(c.Request, aid) {
		c.JSON(http.StatusUnauthorized, model.ErrorNotAuthSession)
		return
	}

	if tags, err = grpc_pool.GrpcWebRpcCall(aid, context.TODO(), &pb.GetTagsInfoReq{AccountId: int32(aid)}, grpc_pool.GetTagsInfo); err != nil {
		logger.Errorf("Get TagsInfo by grpc with error: %+v", err)
		c.JSON(http.StatusInternalServerError, model.ErrorInternalServerError)
		return
	}

	c.JSON(http.StatusOK, tags.(*pb.GetTagsInfoResp))
}
