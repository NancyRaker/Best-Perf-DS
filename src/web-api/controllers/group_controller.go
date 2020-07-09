/**
* @Author: yanKoo
* @Date: 2019/3/11 10:48
* @Description: 处理请求的业务逻辑
 */
package controllers

import (
	"context"
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
	tg "web-api/dao/group"
	"web-api/engine/cache"
	"web-api/engine/grpc_pool"
	"web-api/logger"
	"web-api/model"
	pb "web-api/proto/talk_cloud"
	"web-api/service"
)

// @Summary web更新群组中的设备
// @Description logout by account name and pwd, 请求头中Authorization参数设置为登录时返回的sessionId
// @Accept  json
// @Produce  json
// @Param Authorization header string true "登录时返回的sessionId"
// @Param body body model.GroupListNode true "更新群组中的设备"
// @Success 200 {string} json "{"success":"true","msg": resUpd.ResultMsg.Msg}"
// @Router /group/devices/update [post]
func UpdateGroupDevice(c *gin.Context) {
	updateGroupList := &model.UpdateGroupList{}
	if err := c.BindJSON(updateGroupList); err != nil {
		logger.Debugf("json parse fail , error : %s", err)
		c.JSON(http.StatusBadRequest, model.ErrorRequestBodyParseFailed)
		return
	}
	logger.Debugf("update Group device info: %+v", updateGroupList)

	// 使用session来校验用户
	if !service.ValidateAccountSession(c.Request, updateGroupList.GroupInfo.AccountId) {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":      "session is not right.",
			"error_code": "006",
		})
		return
	}

	// TODO 校验更新群组信息的参数合法性
	if updateGroupList.GroupInfo.GroupName == "" || updateGroupList.GroupInfo.Status == "" || updateGroupList.GroupInfo.AccountId == 0 ||
		len(updateGroupList.AddDeviceInfo) < 0 {
		c.JSON(http.StatusUnprocessableEntity, gin.H{
			"error":      "You need at least the group name, the account id, and at least one device id",
			"error_code": "001",
		})
		return
	}

	status, _ := strconv.Atoi(updateGroupList.GroupInfo.Status)

	// 更新群组
	logger.Debugln("group member update :updateGroupList.GroupInfo.Id :", updateGroupList.GroupInfo.Id)
	resUpd, err := grpc_pool.GrpcWebRpcCall(updateGroupList.GroupInfo.AccountId, context.Background(), &pb.UpdateGroupReq{
		RemoveDeviceInfos: updateGroupList.RemoveDeviceInfo,
		AddDeviceInfos:    updateGroupList.AddDeviceInfo,
		GroupInfo: &pb.Group{
			Id:        int32(updateGroupList.GroupInfo.Id),
			Status:    int32(status),
			AccountId: int32(updateGroupList.GroupInfo.AccountId),
		},
	}, grpc_pool.UpdateGroup)
	if err != nil {
		logger.Debugf("Update group fail , error: %s", err)
		c.JSON(http.StatusInternalServerError, model.ErrorDBError)
		return
	}

	logger.Debugln(resUpd)
	c.JSON(http.StatusOK, gin.H{
		"result": "success",
		"msg":    resUpd.(*pb.UpdateGroupResp).ResultMsg.Msg,
	})
}

// @Summary web创建群组
// @Description web创建群组, 请求头中Authorization参数设置为登录时返回的sessionId
// @Accept  json
// @Produce  json
// @Param Authorization header string true "登录时返回的sessionId"
// @Param body body model.GroupListNode true "创建群组"
// @Success 200 {string} json "{"success":"true","msg": resUpd.ResultMsg.Msg}"
// @Router /group [post]
func CreateGroup(c *gin.Context) {
	gList := &pb.WebCreateGroupReq{} // deviceIds groupInfo
	if err := c.BindJSON(gList); err != nil {
		logger.Debugf("json parse fail , error : %s", err)
		c.JSON(http.StatusBadRequest, model.ErrorRequestBodyParseFailed)
		return
	}
	// 使用session来校验用户
	if !service.ValidateAccountSession(c.Request, int(gList.GroupInfo.AccountId)) {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":      "session is not right.",
			"error_code": "006",
		})
		return
	}
	// TODO 校验创建群组信息的参数合法性
	if gList.GroupInfo.GroupName == "" || gList.GroupInfo.AccountId == 0 { //|| len(gList.DeviceIds) == 0 {
		c.JSON(http.StatusUnprocessableEntity, gin.H{
			"error":      "You need at least the group name, the account id, and at least one device id",
			"error_code": "001",
		})
		return
	}

	// 组名查重
	res, err := tg.CheckDuplicateGName(&model.GroupInfo{AccountId: int(gList.GroupInfo.AccountId), GroupName: gList.GroupInfo.GroupName})
	if err != nil {
		logger.Debugf("CheckDuplicateGName fail , error: %s", err)
		c.JSON(http.StatusInternalServerError, model.ErrorDBError)
		return
	}
	if res > 0 {
		logger.Debugf("CheckDuplicateGName error: %+v", err)
		c.JSON(http.StatusUnprocessableEntity, gin.H{
			"msg":  "group name duplicate",
			"code": "422",
		})
		return
	}

	// 创建群组
	logger.Debugln("start rpc")
	//resCreate, err := webCli.WebCreateGroup(context.Background(), gList)
	resCreate, err := grpc_pool.GrpcWebRpcCall(int(gList.GroupInfo.AccountId), context.Background(), gList, grpc_pool.WebCreateGroup)

	logger.Debugf("create group : %+v", resCreate)
	if err != nil {
		logger.Debugf("create group fail , error: %s", err)
		c.JSON(http.StatusInternalServerError, model.ErrorDBError)
		return
	}

	logger.Debugln(resCreate)
	c.JSON(http.StatusOK, gin.H{
		"group_id":   resCreate.(*pb.GroupInfo).Gid,
		"status":     resCreate.(*pb.GroupInfo).Status,
		"group_name": gList.GroupInfo.GroupName,
	})
}

// @Summary web更新群组信息，目前只更新群组名字
// @Description web创建群组, 请求头中Authorization参数设置为登录时返回的sessionId
// @Accept  json
// @Produce  json
// @Param Authorization header string true "登录时返回的sessionId"
// @Param body body model.GroupInfo true "web更新群组"
// @Success 200 {string} json "{"success":"true","msg": resUpd.ResultMsg.Msg}"
// @Router /group/update [post]
func UpdateGroup(c *gin.Context) {
	gI := &model.GroupInfo{}
	if err := c.BindJSON(gI); err != nil {
		logger.Debugf("json parse fail , error : %s", err)
		c.JSON(http.StatusBadRequest, model.ErrorRequestBodyParseFailed)
		return
	}

	// 使用session来校验用户
	if !service.ValidateAccountSession(c.Request, gI.AccountId) {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":      "session is not right.",
			"error_code": "006",
		})
		return
	}
	// 组名查重
	res, err := tg.CheckDuplicateGName(gI)
	if err != nil {
		logger.Debugf("CheckDuplicateGName fail , error: %s", err)
		c.JSON(http.StatusInternalServerError, model.ErrorDBError)
		return
	}
	if res > 0 {
		logger.Debugf("CheckDuplicateGName error: %+v", err)
		c.JSON(http.StatusUnprocessableEntity, gin.H{
			"msg":  "group name duplicate",
			"code": "422",
		})
		return
	}

	if _, err := grpc_pool.GrpcWebRpcCall(gI.AccountId, context.TODO(),
		&pb.GroupInfo{GroupName: gI.GroupName, Gid: int32(gI.Id)}, grpc_pool.UpdateGroupInfo); err != nil {
		logger.Debugf("update group fail , error: %s", err)
		c.JSON(http.StatusInternalServerError, model.ErrorDBError)
		return
	}
	/*if err := tg.UpdateGroup(gI, db.DBHandler); err != nil {
		log.Log.Debugf("update group fail , error: %s", err)
		c.JSON(http.StatusInternalServerError, model.ErrorDBError)
		return
	}*/

	c.JSON(http.StatusOK, gin.H{
		"result": "success",
		"msg":    "Update group successfully",
	})
}

// @Summary web群组删除，目前只更新群组名字
// @Description web创建群组, 请求头中Authorization参数设置为登录时返回的sessionId
// @Accept  json
// @Produce  json
// @Param Authorization header string true "登录时返回的sessionId"
// @Param body body model.GroupInfo true "web更新群组"
// @Success 200 {string} json "{"success":"true","msg": resUpd.ResultMsg.Msg}"
// @Router /group/delete [post]
func DeleteGroup(c *gin.Context) {
	groupInfo := &model.GroupInfo{}
	if err := c.BindJSON(groupInfo); err != nil {
		logger.Debugf("json parse fail , error : %s", err)
		c.JSON(http.StatusBadRequest, model.ErrorRequestBodyParseFailed)
		return
	}
	// 使用session来校验用户
	if !service.ValidateAccountSession(c.Request, groupInfo.AccountId) {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":      "session is not right.",
			"error_code": "006"})
		return
	}

	logger.Debugln("start rpc")
	if _, err := grpc_pool.GrpcWebRpcCall(
		groupInfo.AccountId, context.Background(),
		&pb.GroupDelReq{Gid: int32(groupInfo.Id), Uid: int32(groupInfo.AccountId)},
		grpc_pool.DeleteGroup); err != nil {
		logger.Debugf("delete group fail , error: %s", err)
		c.JSON(http.StatusInternalServerError, model.ErrorInternalServerError)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"result": "success",
		"msg":    "Delete group successfully",
	})
}

// @Summary web更新群组成员是否为管理员
// @Description web更新群组成员是否为管理员, 请求头中Authorization参数设置为登录时返回的sessionId
// @Accept  json
// @Produce  json
// @Param Authorization header string true "登录时返回的sessionId"
// @Param body body model.SwagUpdGManagerReq true "web更新群组管理员body内容，注意 RoleType 只能为1或者2， 2是指修改为管理员，1是指修改为普通成员"
// @Success 200 {string} json "{"success":"true","msg": resUpd.ResultMsg.Msg}"
// @Router /group_manager/update [post]
func UpdateGroupManager(c *gin.Context) {
	updGManagerReq := &pb.UpdGManagerReq{}
	if err := c.BindJSON(updGManagerReq); err != nil {
		logger.Debugf("json parse fail , error : %s", err)
		c.JSON(http.StatusBadRequest, model.ErrorRequestBodyParseFailed)
		return
	}
	logger.Debugf("receiver request body: %+v", updGManagerReq.AccountId)
	// 使用session来校验用户
	if !service.ValidateAccountSession(c.Request, int(updGManagerReq.AccountId)) {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":      "session is not right.",
			"error_code": "006",
		})
		return
	}
	// TODO 限制管理员人数
	//webCli := pb.NewWebServiceClient(Conn)
	//res, err := webCli.UpdateGroupManager(context.TODO(), updGManagerReq)
	res, err := grpc_pool.GrpcWebRpcCall(int(updGManagerReq.AccountId), context.TODO(), updGManagerReq, grpc_pool.UpdateGroupManager)
	if err != nil {
		logger.Errorf("update management fail with error: %+v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"result": "error",
			"msg":    "Update group management fail, please try again later",
		})
		return
	}
	logger.Debugf("grpc call after with res: %+v", res.(*pb.UpdGManagerResp))

	c.JSON(http.StatusOK, gin.H{
		"result": "success",
		"msg":    "Update group Manager successfully",
	})
}

// @Summary 获取单个群组信息
// @Produce  json
// @Param accountId path string true "当前用户的账号Id"
// @Param Authorization header string true "登录时返回的sessionId"
// @Success  200 {object} model.SwagGetDeviceLogResp "处理正确返回的数据"
// @Router /group/info/{accountId}/{groupId} [get]
func GetGroupInfo(c *gin.Context) {
	var (
		err            error
		aIdStr         = c.Param("accountId")
		gIdStr         = c.Param("groupId")
		groupListNodes []*model.GroupListNode
	)

	// 使用session来校验用户
	aId, _ := strconv.Atoi(aIdStr)
	gId, _ := strconv.Atoi(gIdStr)
	if !service.ValidateAccountSession(c.Request, aId) {
		c.JSON(http.StatusUnauthorized, model.ErrorNotAuthSession)
		return
	}

	// TODO 需要去im获取这些数据
	if groupListNodes, _, err = service.GetGroupListInfos(aId, []int32{int32(gId)}, cache.GetRedisClient()); err != nil {
		logger.Errorf("Get group list by grpc with error: %+v", err)
		c.JSON(http.StatusInternalServerError, model.ErrorInternalServerError)
		return
	}

	if groupListNodes == nil || len(groupListNodes) == 0 {
		logger.Error("Get group list error : no group")
		c.JSON(http.StatusOK, gin.H{"msg": "no group"})
		return
	}

	c.JSON(http.StatusOK, groupListNodes[0])
}

// @Summary 获取群组列表信息
// @Produce  json
// @Param accountId path string true "当前用户的账号Id"
// @Param Authorization header string true "登录时返回的sessionId"
// @Success  200 {object} model.SwagGetDeviceLogResp "处理正确返回的数据"
// @Router /group/info/{accountId} [get]
func GetGroupList(c *gin.Context) {
	var (
		err            error
		aIdStr         = c.Param("accountId")
		groupListNodes []*model.GroupListNode

		gIds = make([]int32, 0)
	)

	// 使用session来校验用户
	aId, _ := strconv.Atoi(aIdStr)
	if !service.ValidateAccountSession(c.Request, aId) {
		c.JSON(http.StatusUnauthorized, model.ErrorNotAuthSession)
		return
	}

	// 4. 获取群组信息
	groups, err := tg.SelectGroupsByAccountId(aId)

	for _, groupInfo := range groups {
		gIds = append(gIds, int32(groupInfo.Id))
	}
	groupListNodes, _, err = service.GetGroupListInfos(aId, gIds, cache.GetRedisClient())
	if err != nil {
		logger.Debugf("GetAccountInfo GetGroupListInfos error : %s", err)
		c.JSON(http.StatusInternalServerError, model.ErrorInternalServerError)
		return
	}
	c.JSON(http.StatusOK, groupListNodes)
}
