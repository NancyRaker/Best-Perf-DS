/*
@Time : 2019/6/20 10:22 
@Author : yanKoo
@File : TagTasks
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
	SAVE_TAG_TASK_INFO   int32 = 1
	UPDATE_TAG_TASK_INFO int32 = 2
	DELETE_TAG_TASK_INFO int32 = 3
)

// @Summary 增加，删除，修改TagTasks信息
// @Produce  json
// @Param Authorization header string true "登录时返回的sessionId"
// @Param accountId path string true "当前用户的账号Id"
// @Param Body body model.SwagTagTasksInfoReq true "获取TagTasks的model"
// @Success 200 {object} model.SwagGetDeviceLocationResp "正确处理返回的数据"
// @Router /tag_tasks/{accountId} [post]
func OperationTagTasksList(c *gin.Context) {
	aidStr := c.Param("accountId")
	accountTagTasksInfo := &pb.TagTasksListReq{}
	if err := c.BindJSON(accountTagTasksInfo); err != nil {
		logger.Debugf("json parse fail , error : %s", err)
		c.JSON(http.StatusBadRequest, model.ErrorRequestBodyParseFailed)
		return
	}

	// 校验ops参数
	if accountTagTasksInfo.Ops < 1 || accountTagTasksInfo.Ops > 3 {
		logger.Errorf("accountTagTasks Info ops error: %d", accountTagTasksInfo.Ops)
		c.JSON(http.StatusBadRequest, model.ErrorRequestBodyParseFailed)
		return
	}

	// 使用session来校验用户
	aid, _ := strconv.Atoi(aidStr)
	if !service.ValidateAccountSession(c.Request, aid) {
		c.JSON(http.StatusUnauthorized, model.ErrorNotAuthSession)
		return
	}
	accountTagTasksInfo.AccountId = int32(aid)
	logger.Errorf("accountTagTasks Info : %+v", accountTagTasksInfo)
	switch accountTagTasksInfo.Ops {
	case SAVE_TAG_TASK_INFO:
		saveTagTasksInfo(c, accountTagTasksInfo)
	case UPDATE_TAG_TASK_INFO:
		updateTagTasksInfo(c, accountTagTasksInfo)
	case DELETE_TAG_TASK_INFO:
		deleteTagTasksInfo(c, accountTagTasksInfo)
	}
}

// 保存TagTasks信息
func saveTagTasksInfo(c *gin.Context, TagTasksInfos *pb.TagTasksListReq) {
	var err error

	// 校验数据
	for _, tagTaskList := range TagTasksInfos.TagTaskLists {
		for _, tagTaskNode := range tagTaskList.TagTaskNodes {
			// 校验位置信息
			if tagTaskNode.TagId <= 0 || tagTaskNode.DeviceId <= 0 ||
				tagTaskNode.OrderEndTime <= 0 || tagTaskNode.OrderStartTime <= 0 ||
				(tagTaskNode.OrderEndTime < tagTaskNode.OrderStartTime) {
				logger.Debugf("tagTaskList params:%+v", tagTaskList)
				err = errors.New("tagTaskList name or addr or timestamp or uuid is invalid")
				break
			}
		}
	}

	if err != nil {
		logger.Debugf("saveTagTasksInfo validate tagTaskList info fail with error: %+v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error msg": err.Error()})
		return
	}

	TagTasksInfoProcess(c, TagTasksInfos)
}

// 修改TagTasks信息
func updateTagTasksInfo(c *gin.Context, TagTasksInfos *pb.TagTasksListReq) {
	var err error
	// 校验数据
	for _, tagTaskList := range TagTasksInfos.TagTaskLists {
		for _, tagTaskNode := range tagTaskList.TagTaskNodes {
			// 校验位置信息
			if tagTaskNode.AccountId <= 0 || tagTaskNode.TagId <= 0 || tagTaskNode.DeviceId <= 0 ||
				tagTaskNode.OrderEndTime <= 0 || tagTaskNode.OrderStartTime <= 0 ||
				(tagTaskNode.OrderEndTime < tagTaskNode.OrderStartTime) {
				logger.Debugf("tagTaskList params:%+v", tagTaskList)
				err = errors.New("tagTaskList name or addr or timestamp or uuid is invalid")
				break
			}
		}
	}

	if err != nil {
		logger.Debugf("validate TagTasks info fail with error: %+v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error msg": err.Error()})
		return
	}

	TagTasksInfoProcess(c, TagTasksInfos)
}

// 删除TagTasks信息
func deleteTagTasksInfo(c *gin.Context, TagTasksInfos *pb.TagTasksListReq) {
	var err error
	// 校验数据
	for _, tagTaskList := range TagTasksInfos.TagTaskLists {
		for _, tagTaskNode := range tagTaskList.TagTaskNodes {
			// 校验位置信息
			if tagTaskNode.TagId <= 0 || tagTaskNode.DeviceId <= 0 ||
				tagTaskNode.OrderEndTime <= 0 || tagTaskNode.OrderStartTime <= 0 ||
				(tagTaskNode.OrderEndTime < tagTaskNode.OrderStartTime) {
				logger.Debugf("tagTaskList params:%+v", tagTaskList)
				err = errors.New("tagTaskList name or addr or timestamp or uuid is invalid")
				break
			}
		}
	}

	if err != nil {
		logger.Debugf("validate TagTasks info fail with error: %+v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error msg": err.Error()})
		return
	}

	TagTasksInfoProcess(c, TagTasksInfos)
}

// grpc调用，提交信息
func TagTasksInfoProcess(c *gin.Context, TagTasksInfos *pb.TagTasksListReq) {
	var (
		res interface{} //*pb.TagTasksInfoResp
		err error
	)
	aidStr := c.Param("accountId")
	aid, _ := strconv.Atoi(aidStr)
	if res, err = grpc_pool.GrpcWebRpcCall(aid, context.TODO(), TagTasksInfos, grpc_pool.PostTagTasksList); err != nil {
		//if res, err = webCli.PostTagTasksInfo(context.TODO(), TagTasksInfos); err != nil {
		logger.Errorf("operate TagTasks info by grpc with error: %+v", err)
		c.JSON(http.StatusInternalServerError, model.ErrorInternalServerError)
		return
	}

	if res.(*pb.TagTasksListResp).Res.Code != http.StatusOK {
		c.JSON(http.StatusBadRequest, gin.H{"msg": res.(*pb.TagTasksListResp).Res.Msg})
		return
	}

	c.JSON(http.StatusOK, model.SuccessResponse)
}

//@Summary 获取TagTasks信息
//@Produce  json
//@Param accountId path string true "当前用户的账号Id"
//@Param Authorization header string true "登录时返回的sessionId"
//@Success 200 {string} json "	{"msg":"delete account successfully.","result": "success"}"
//@Router /TagTasks/{accountId} [get]
func GetTagTasksInfo(c *gin.Context) {
	//var (
	//	err      error
	//	aidStr   = c.Param("accountId")
	//	TagTasks interface{} // *pb.GetTagTasksInfoResp
	//)
	//// 使用session来校验用户
	//aid, _ := strconv.Atoi(aidStr)
	//if !service.ValidateAccountSession(c.Request, aid) {
	//	c.JSON(http.StatusUnauthorized, model.ErrorNotAuthSession)
	//	return
	//}
	//
	////if TagTaskss, err = webCli.GetTagTasksInfo(context.TODO(), &pb.GetTagTasksInfoReq{AccountId:int32(aid)}); err != nil {
	//if TagTasks, err = grpc_pool.GrpcWebRpcCall(aid, context.TODO(), &pb.GetTagTasksInfoReq{AccountId: int32(aid)}, grpc_pool.GetTagTasksInfo); err != nil {
	//	logger.Errorf("Get TagTasksInfo by grpc with error: %+v", err)
	//	c.JSON(http.StatusInternalServerError, model.ErrorInternalServerError)
	//	return
	//}
	//
	//c.JSON(http.StatusOK, TagTasks.(*pb.GetTagTasksInfoResp))
}

// @Summary 获取Task detail信息
// @Produce  json
// @Param accountId path string true "当前用户的账号Id"
// @Param Authorization header string true "登录时返回的sessionId"
// @Success 200 {string} json "	{"msg":"delete account successfully.","result": "success"}"
// @Router /tag_tasks_device/{account-id}/{device-id} [get]
func QueryDeviceTask(c *gin.Context) {
	var (
		err         error
		aidStr      = c.Param("account-id")
		deviceIdStr = c.Param("device-id")
		TagTasks    interface{} // *pb.GetTagTasksInfoResp
	)
	// 使用session来校验用户
	aid, _ := strconv.Atoi(aidStr)
	deviceId, _ := strconv.Atoi(deviceIdStr)
	if !service.ValidateAccountSession(c.Request, aid) {
		c.JSON(http.StatusUnauthorized, model.ErrorNotAuthSession)
		return
	}

	if TagTasks, err = grpc_pool.GrpcWebRpcCall(aid, context.TODO(), &pb.DeviceTasksReq{AccountId: int32(aid), DeviceId: int32(deviceId)}, grpc_pool.QueryTasksByDevice); err != nil {
		logger.Errorf("Get TagTasksInfo by grpc with error: %+v", err)
		c.JSON(http.StatusInternalServerError, model.ErrorInternalServerError)
		return
	}

	c.JSON(http.StatusOK, TagTasks.(*pb.DeviceTasksResp))
}

// @Summary 获取TagTasks信息
// @Produce  json
// @Param accountId path string true "当前用户的账号Id"
// @Param Authorization header string true "登录时返回的sessionId"
// @Success 200 {string} json "	{"msg":"delete account successfully.","result": "success"}"
// @Router /device/clock/:account-id/:device-id [get]
func QueryDeviceClockStatus(c *gin.Context) {
	var (
		err               error
		aidStr            = c.Param("account-id")
		deviceIdStr       = c.Param("device-id")
		startTimeStampStr = c.Param("start-timestamp")
		clockStatus       interface{} // *pb.GetTagTasksInfoResp
	)
	// 使用session来校验用户
	aid, _ := strconv.Atoi(aidStr)
	deviceId, _ := strconv.Atoi(deviceIdStr)
	startTimeStamp, _ := strconv.ParseUint(startTimeStampStr, 10, 64)
	if !service.ValidateAccountSession(c.Request, aid) {
		c.JSON(http.StatusUnauthorized, model.ErrorNotAuthSession)
		return
	}

	if clockStatus, err = grpc_pool.GrpcWebRpcCall(aid, context.TODO(),
		&pb.ClockStatusReq{AccountId: int32(aid), DeviceId: int32(deviceId), StartTimestamp: startTimeStamp}, grpc_pool.DeviceClockStatus); err != nil {

		logger.Errorf("Get DeviceClock Status by grpc with error: %+v", err)
		c.JSON(http.StatusInternalServerError, model.ErrorInternalServerError)
		return
	}

	c.JSON(http.StatusOK, clockStatus.(*pb.ClockStatusResp))
}

// @Summary 获取TagTasks信息
// @Produce  json
// @Param accountId path string true "当前用户的账号Id"
// @Param Authorization header string true "登录时返回的sessionId"
// @Success 200 {string} json "	{"msg":"delete account successfully.","result": "success"}"
// @Router /tag_tasks/{account-id}/{task-id} [get]
func QueryTaskDetail(c *gin.Context) {
	var (
		err       error
		aidStr    = c.Param("account-id")
		taskIdStr = c.Param("task-id")
		TagTasks  interface{} // *pb.GetTagTasksInfoResp
	)
	// 使用session来校验用户
	aid, _ := strconv.Atoi(aidStr)
	taskId, _ := strconv.Atoi(taskIdStr)
	if !service.ValidateAccountSession(c.Request, aid) {
		c.JSON(http.StatusUnauthorized, model.ErrorNotAuthSession)
		return
	}

	if TagTasks, err = grpc_pool.GrpcWebRpcCall(aid, context.TODO(), &pb.TaskDetailReq{AccountId: int32(aid), TaskId: int32(taskId)}, grpc_pool.QueryTaskDetail); err != nil {
		logger.Errorf("Get TagTasksInfo by grpc with error: %+v", err)
		c.JSON(http.StatusInternalServerError, model.ErrorInternalServerError)
		return
	}

	c.JSON(http.StatusOK, TagTasks.(*pb.TaskDetailResp))
}

// @Summary 增加，删除，修改调度员报告发送等信息
// @Produce  json
// @Param Authorization header string true "登录时返回的sessionId"
// @Param accountId path string true "当前用户的账号Id"
// @Param Body body model.SwagTagTasksInfoReq true "获取TagTasks的model"
// @Success 200 {object} model.SwagGetDeviceLocationResp "正确处理返回的数据"
// @Router /account/clock/{account-id} [post]
func SetAccountReportInfo(c *gin.Context) {
	aidStr := c.Param("account-id")
	reportInfoReq := &pb.ReportInfoReq{}
	if err := c.BindJSON(reportInfoReq); err != nil {
		logger.Debugf("json parse fail , error : %s", err)
		c.JSON(http.StatusBadRequest, model.ErrorRequestBodyParseFailed)
		return
	}

	// 校验参数
	if reportInfoReq.DetailParam.AccountId <= 0 {
		c.JSON(http.StatusBadRequest, model.ErrorRequestBodyParseFailed)
		return
	}

	// 使用session来校验用户
	aid, _ := strconv.Atoi(aidStr)
	if !service.ValidateAccountSession(c.Request, aid) {
		c.JSON(http.StatusUnauthorized, model.ErrorNotAuthSession)
		return
	}

	if _, err := grpc_pool.GrpcWebRpcCall(aid, context.TODO(), reportInfoReq, grpc_pool.SetReportInfo); err != nil {
		logger.Errorf("Get TagTasksInfo by grpc with error: %+v", err)
		c.JSON(http.StatusInternalServerError, model.ErrorInternalServerError)
		return
	}

	c.JSON(http.StatusOK, model.SuccessResponse)
}

// @Summary 获取调度员报告发送等信息
// @Produce  json
// @Param account-id path string true "当前用户的账号Id"
// @Param Authorization header string true "登录时返回的sessionId"
// @Success 200 {string} json "	{"msg":" successfully.","result": "success"}"
// @Router /account/clock/{account-id} [get]
func QueryAccountReportInfo(c *gin.Context) {
	var (
		err         error
		aidStr      = c.Param("account-id")
		clockStatus interface{}
	)
	// 使用session来校验用户
	aid, _ := strconv.Atoi(aidStr)
	if !service.ValidateAccountSession(c.Request, aid) {
		c.JSON(http.StatusUnauthorized, model.ErrorNotAuthSession)
		return
	}

	if clockStatus, err = grpc_pool.GrpcWebRpcCall(aid, context.TODO(),
		&pb.ReportSetInfoReq{AccountId: int32(aid)}, grpc_pool.QueryReportInfo); err != nil {

		logger.Errorf("Get DeviceClock Status by grpc with error: %+v", err)
		c.JSON(http.StatusInternalServerError, model.ErrorInternalServerError)
		return
	}

	c.JSON(http.StatusOK, clockStatus.(*pb.ReportSetInfoResp))
}
