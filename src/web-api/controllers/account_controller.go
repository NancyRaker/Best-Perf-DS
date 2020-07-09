/**
* @Author: yanKoo
* @Date: 2019/3/11 10:48
* @Description: 处理请求的业务逻辑
 */
package controllers

import (
	"context"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
	"time"
	tc "web-api/dao/customer" // table customer
	tu "web-api/dao/user"
	"web-api/engine/grpc_pool"
	"web-api/logger"
	"web-api/model"
	pb "web-api/proto/talk_cloud"
	"web-api/utils"

	"strconv"
	"web-api/service"
)

// @Summary 创建下级账户
// @Produce  json
// @Param Authorization header string true "登录时返回的sessionId"
// @Param Body body model.SwagCreateAccount true "创建下级账户所用的信息，role_id只能是2、3、4， 低等级不能创建等级高的用户"
// @Success 200 {object} model.SwagCreateAccountResp true "操作成功会返回的json内容"
// @Router /account [post]
func CreateAccountBySuperior(c *gin.Context) {
	// 0. 取出Post中的表单内容
	uBody := &model.CreateAccount{}
	if err := c.BindJSON(uBody); err != nil {
		logger.Debugln("bind json error : ", err)
		c.JSON(http.StatusUnprocessableEntity, model.ErrorRequestBodyParseFailed)
		return
	}

	// 1. 使用session来校验用户
	if !service.ValidateAccountSession(c.Request, uBody.Pid) {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":      "session is not right.",
			"error_code": "006",
		})
		return
	}

	// 2. 数据格式合法性校验，首先不能为空，其次每个格式都必须校验
	/*if uBody.NickName == "" || uBody.Username == "" || uBody.Pwd == "" || uBody.ConfirmPwd == "" {
		c.JSON(http.StatusUnprocessableEntity, gin.H{
			"error":      " Please fill in the required fields ",
			"error_code": "0001",
		})
		return
	}*///改用binding

	// 校验昵称
	if !utils.CheckNickName(uBody.NickName) {
		logger.Debugln("NickName format error", uBody.NickName)
		c.JSON(http.StatusUnprocessableEntity, gin.H{
			"error":      "昵称只能输入1-20个以字母或者数字开头、可以含中文、下划线的字串。",
			"error_code": "0002",
		})
		return
	}

	if !utils.CheckUserName(uBody.Username) {
		logger.Debugln("Username format error")
		c.JSON(http.StatusUnprocessableEntity, gin.H{
			"error":      "用户名只能输入5-20个包含字母、数字或下划线的字串",
			"error_code": "0003",
		})
		return
	}

	// 名字查重
	aCount, err := tc.GetAccountByName(uBody.Username)
	if err != nil {
		logger.Debugln("db error : ", err)
		c.JSON(http.StatusInternalServerError, model.ErrorDBError)
		return
	}
	if aCount > 0 {
		c.JSON(http.StatusUnprocessableEntity, gin.H{
			"error":      "User is exist already",
			"error_code": "0005",
		})
		return
	}

	// 校验密码
	if !utils.CheckPwd(uBody.ConfirmPwd) || !utils.CheckPwd(uBody.Pwd) {
		logger.Debugln("Pwd format error")
		c.JSON(http.StatusUnprocessableEntity, gin.H{
			"error":      "密码6位-16位，至少包含一个数字字母",
			"error_code": "0004",
		})
		return
	}
	if uBody.ConfirmPwd != uBody.Pwd {
		logger.Debugln("Confirm Pwd is not match pwd")
		c.JSON(http.StatusUnprocessableEntity, gin.H{
			"error":      "两次输入密码必须一致",
			"error_code": "0005",
		})
		return
	}

	// 1是普通用户， 2是调度员， 3是经销商 4是公司，5是超级管理员root
	logger.Debugln("创建等级:", uBody.RoleId)
	if uBody.RoleId < 2 || uBody.RoleId > 4 {
		logger.Debugln("创建权限出错")
		c.JSON(http.StatusUnprocessableEntity, model.ErrorRequestBodyParseFailed)
		return
	}

	// 判断用户类型是否符合上级创建下级
	parentAccount, err := tc.GetAccount(uBody.Pid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.ErrorDBError)
		return
	}
	// 只能给下级创建或者平级
	if parentAccount.RoleId < uBody.RoleId {
		c.JSON(http.StatusInternalServerError, model.ErrorCreateAccountError)
		return
	}

	// 调度员不能创建调度员
	if parentAccount.RoleId == 2 && uBody.RoleId == 2 {
		c.JSON(http.StatusInternalServerError, model.ErrorCreateAccountPriError)
		return
	}

	// 3. 添加账户
	uId, err := grpc_pool.GrpcWebRpcCall(uBody.Pid, context.Background(), &pb.CreateAccountReq{
		Username: uBody.Username,
		NickName: uBody.NickName,
		Pwd:      uBody.Pwd,
		RoleId:   int32(uBody.RoleId),
		Pid:      int32(uBody.Pid),
	}, grpc_pool.CreateAccount)
	/*uId, err := tc.AddAccount(uBody)
	*/if err != nil {
		c.JSON(http.StatusInternalServerError, model.ErrorDBError)
		return
	}
	//4. 返回消息内容
	c.JSON(http.StatusCreated, gin.H{
		"result":     "success",
		"account_id": uId,
	})
}

// @Summary 获取账户信息
// @Produce  json
// @Param account_name path string true "当前用户的账号"
// @Param Authorization header string true "登录时返回的sessionId"
// @Success 200 {object} model.SwagGetAccountInfoResp "正确操作应该返回的数据"
// @Router /account/{account-id} [get]
func GetAccountInfo(c *gin.Context) {
	start_t := time.Now().UnixNano()
	logger.Debugf("start get account info: %d\n", start_t)
	//accountId := c.Param("account_name")
	accountIdStr := c.Param("account-id")
	if accountIdStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":      "request url is not correct.",
			"error_code": "001",
		})
		return
	}
	session_t := time.Now().UnixNano()
	logger.Debugf("start check session: %d, before now use: %d ms\n", session_t, (session_t-start_t)/1000000)
	accountId, _ := strconv.Atoi(accountIdStr)
	// 1. 使用session来校验用户
	if !service.ValidateAccountSession(c.Request, accountId) {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":      "session is not right.",
			"error_code": "006",
		})
		return
	}

	// 2. 获取账户信息
	ai_t := time.Now().UnixNano()
	logger.Debugf("start get a info: %d, before now use: %d ms\n", ai_t, (ai_t-session_t)/1000000)
	ai, err := tc.GetAccount(accountId)
	if err != nil {
		logger.Debugf("Error in GetAccountInfo: %s", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":      "get accountInfo DB error",
			"error_code": "007",
		})
		return
	}

	var (
		ctx, cancelFunc = context.WithCancel(context.TODO())
		errOut          = make(chan interface{}, 10)
		accountTreeChan = make(chan interface{}, 1)
	)

	// 5. 账户层级树
	service.GetAccountTree(ctx, ai, accountTreeChan, errOut)
	//tree_t := time.Now().UnixNano()
	//fmt.Printf("start get tree : %d, before now use: %d ms\n", tree_t, (tree_t-start_t)/1000000)
	//go service.ProcessAccountInfo(ctx, ai, accountTreeChan, errOut, getAccountTree)

	deviceAll, statMap, err := service.GetDevicesForDispatchers(ctx, cancelFunc, ai.Id, errOut)
	if err != nil {
		logger.Debugf("GetAccountInfo fail with : %+v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"err": "Internal server error, please try again."})
		return
	}
	/*	// 3. 获取所有用户设备
		ginfo_t := time.Now().UnixNano()
		fmt.Printf("start get all device info: %d, before now use: %d ms\n", ginfo_t, (ginfo_t-ai_t)/1000000)
		go service.ProcessAccountInfo(ctx, ai.Id, devicesChan, errOut, service.getAllDeviceBelongAccount)

		// 4. 获取在线状态
		stat_t := time.Now().UnixNano()
		fmt.Printf("start get all stat : %d, before now use: %d ms\n", stat_t, (stat_t-ginfo_t)/1000000)
		go service.ProcessAccountInfo(ctx, ai.Id, devicesStatusChan, errOut, service.getDevicesStatus)


		// 6.根据设备在线状态排序
		deviceAll, statMap := func(ctx context.Context, cancel context.CancelFunc,
			devicesC, statusMapC, err chan interface{}) (*[]*model.Device, map[int32]bool) {
			for {
				select {
				case <-ctx.Done():
					return nil, nil
				case errorMsg := <-err:
					cancel()
					log.Log.Debugf("GetAccountInfo fail with : %+v", errorMsg)
					c.JSON(http.StatusInternalServerError, gin.H{"err": "Internal server error, please try again."})
				default:
					devices := (<-devicesChan).(*[]*model.Device)
					statusMap := (<-statusMapC).(map[int32]bool)
					service.sortDeviceByOnlineStatus(devices, statusMap)
					return devices, statusMap
				}
			}
		}(ctx, cancelFunc, devicesChan, devicesStatusChan, errOut)*/

	ai.Pwd = "" // 不把密码暴露出去
	c.JSON(http.StatusOK, gin.H{
		"account_info": ai,
		"tree_data":    (<-accountTreeChan).(*model.AccountClass),
		"device_list":  *deviceAll,
		"device_list_info": struct {
			DeviceTotal int `json:"device_total"`
			OnlineTotal int `json:"online_total"`
		}{DeviceTotal: len(*deviceAll), OnlineTotal: len(statMap)},
	})
	end_t := time.Now().UnixNano()
	logger.Debugf("end request : %d, before now use: %d ms\n", end_t, (end_t-start_t)/1000000)
}

// @Summary 更新下级账户信息
// @Produce  json
// @Param Authorization header string true "登录时返回的sessionId"
// @Param Body body model.SwagAccountUpdate true "更新下级账户的信息，其中id和loginId和NickName不能为空"
// @Success 200 {object} model.SwagAccountUpdateResp "更新下级账户成功应该返回的信息"
// @Router /account/info/update [post]
func UpdateAccountInfo(c *gin.Context) {
	accInf := &model.AccountUpdate{}
	if err := c.BindJSON(accInf); err != nil {
		logger.Debugf("%s", err)
		c.JSON(http.StatusUnprocessableEntity, gin.H{
			"error":      "Request body is not correct.",
			"error_code": "001",
		})
		return
	}

	// 校验参数信息 ：校首先必须要有id，其次是每个参数的合法性，首先都不允许为空 TODO 校验电话号码邮箱
	if accInf.LoginId == "" {
		logger.Debugf("account id is nil")
		c.JSON(http.StatusUnprocessableEntity, gin.H{
			"error":      "The account id cannot be empty",
			"error_code": "003",
		})
		return
	}

	loginId, _ := strconv.Atoi(accInf.LoginId)

	// 使用session来校验用户
	if !service.ValidateAccountSession(c.Request, loginId) {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":      "session is not right.",
			"error_code": "006",
		})
		return
	}

	if err := tc.UpdateAccount(accInf); err != nil {
		logger.Debugln("Update account error :", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":      "Update AccountInfoDB error",
			"error_code": "009",
		})
		return
	} else {
		c.JSON(http.StatusOK, gin.H{
			"result": "success",
			"msg":    "update account success",
		})
	}
}

// @Summary 更新用户密码
// @Produce  json
// @Param Authorization header string true "登录时返回的sessionId"
// @Param Body body model.SwagAccountPwd true "更新用户密码"
// @Success 200 {object} model.SwagAccountPwdResp "正确处理返回的数据"
// @Router /account/pwd/update [post]
func UpdateAccountPwd(c *gin.Context) {
	accPwd := &model.AccountPwd{}
	if err := c.BindJSON(accPwd); err != nil {
		logger.Debugf("json parse fail , error : %s", err)
		c.JSON(http.StatusBadRequest, model.ErrorRequestBodyParseFailed)
		return
	}

	// 使用session来校验用户是否登录
	aid, _ := strconv.Atoi(accPwd.Id)
	if !service.ValidateAccountSession(c.Request, aid) {
		c.JSON(http.StatusUnauthorized, model.ErrorNotAuthSession)
		return
	}
	// 校验参数信息 ：校首先必须要有id，都不允许为空
	if accPwd.Id == "" {
		logger.Debugf("account id is nil")
		c.JSON(http.StatusUnprocessableEntity, gin.H{
			"error":      "account id is null",
			"error_code": "001",
		})
		return
	}

	// 校验密码
	if !utils.CheckPwd(accPwd.ConfirmPwd) || !utils.CheckPwd(accPwd.NewPwd) {
		logger.Debugln("Pwd format error")
		c.JSON(http.StatusUnprocessableEntity, gin.H{
			"error":      "密码6位-16位，至少包含一个数字字母",
			"error_code": "0004",
		})
		return
	}
	// 两次输入的密码必须一致
	if accPwd.ConfirmPwd != accPwd.NewPwd {
		logger.Debugln("Confirm Pwd is not match pwd")
		c.JSON(http.StatusUnprocessableEntity, gin.H{
			"error":      "两次输入密码必须一致",
			"error_code": "0005",
		})
		return
	}

	// 新密码不能和旧密码不同
	if accPwd.NewPwd == accPwd.OldPwd {
		c.JSON(http.StatusUnprocessableEntity, gin.H{
			"error":      "The new password cant't be the same as the old password.",
			"error_code": "003",
		})
		return
	}

	// 判断密码是否正确
	pwd, err := tc.GetAccountPwdByKey(aid)
	if err != nil {
		logger.Debugf("db error : %s", err)
		c.JSON(http.StatusInternalServerError, model.ErrorDBError)
		return
	}
	if pwd != accPwd.OldPwd {
		logger.Debugf("db pwd: %s", pwd)
		logger.Debugf("input old pwd %s", accPwd.OldPwd)
		c.JSON(http.StatusUnprocessableEntity, gin.H{
			"error":      "Old password is not match",
			"error_code": "002",
		})
		return
	}

	// 更新密码
	id, _ := strconv.Atoi(accPwd.Id)
	if err := tc.UpdateAccountPwd(accPwd.NewPwd, id); err != nil {
		logger.Debugln("Update account errr :", err)
		c.JSON(http.StatusInternalServerError, model.ErrorDBError)
		return
	} else {
		c.JSON(http.StatusOK, gin.H{
			"result": "success",
			"msg":    "Password changed successfully",
		})
	}
}

// @Summary 获取账户下级目录
// @Produce  json
// @Param accountId path string true "当前用户的账号Id"
// @Param searchId path string true "获取用户下级目录的账号Id"
// @Param Authorization header string true "登录时返回的sessionId"
// @Success 200 {object} model.SwagGetAccountClassResp "正确处理返回的数据"
// @Router /account_class/{accountId}/{searchId} [get]
func GetAccountClass(c *gin.Context) {
	accountId := c.Param("accountId")
	searchId := c.Param("searchId")

	logger.Debugln("searchId:", searchId, "accountId", accountId)
	//使用session来校验用户
	aid, _ := strconv.Atoi(accountId)
	sid, _ := strconv.Atoi(searchId)

	if !service.ValidateAccountSession(c.Request, aid) {
		c.JSON(http.StatusUnauthorized, model.ErrorNotAuthSession)
		return
	}

	// 查询数据返回
	root, err := tc.GetAccount(sid)
	if err != nil {
		logger.Debugf("GetAccount db error : %s", err)
		c.JSON(http.StatusInternalServerError, model.ErrorDBError)
		return
	}

	resElem, err := tc.SelectChildByPId(sid)
	if err != nil {
		logger.Debugf("SelectChildByPId db error : %s", err)
		c.JSON(http.StatusInternalServerError, model.ErrorDBError)
		return
	}

	cList := make([]*model.AccountClass, 0)
	for i := 0; i < len(resElem); i++ {
		child, err := tc.GetAccount((*resElem[i]).Id)
		if err != nil {
			logger.Debugf("db error : %s", err)
			c.JSON(http.StatusInternalServerError, model.ErrorDBError)
			return
		}

		cList = append(cList, &model.AccountClass{
			Id:              child.Id,
			AccountName:     child.Username,
			AccountNickName: child.NickName,
		})
	}

	resp := &model.AccountClass{
		Id:              sid,
		AccountName:     root.Username,
		AccountNickName: root.NickName,
		Children:        cList,
	}

	c.JSON(http.StatusOK, gin.H{
		"result":    "success",
		"tree_data": resp,
	})
}

// @Summary 获取账户的设备信息
// @Produce  json
// @Param accountId path string true "当前用户的账号Id"
// @Param getAdviceId path string true "获取用户设备信息的账号Id"
// @Param Authorization header string true "登录时返回的sessionId"
// @Success 200 {object} model.SwagGetAccountDeviceResp "正确处理返回的数据"
// @Router /account_device/{accountId}/{getAdviceId} [get]
func GetAccountDevice(c *gin.Context) {
	accountId := c.Param("accountId")
	getAdviceId := c.Param("getAdviceId")

	// 使用session来校验用户, 保证上级用户已登录
	aid, _ := strconv.Atoi(accountId)
	getAId, _ := strconv.Atoi(getAdviceId)
	if !service.ValidateAccountSession(c.Request, aid) {
		c.JSON(http.StatusUnauthorized, model.ErrorNotAuthSession)
		return
	}
	// 获取账户信息
	ai, err := tc.GetAccount(getAId)
	if err != nil {
		logger.Debugf("Error in GetAccountInfo: %s", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":      "get accountInfo DB error",
			"error_code": "007",
		})
		return
	}
	// 获取所有设备
	deviceAll, err := tu.SelectUserByAccountId(ai.Id)
	if err != nil {
		logger.Debugf("db error : %s", err)
		c.JSON(http.StatusInternalServerError, model.ErrorDBError)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"account_info": ai,
		"devices":      deviceAll,
	})
}

// @Summary 获取账户的某些设备gps信息
// @Produce  json
// @Param Authorization header string true "登录时返回的sessionId"
// @Param accountId path string true "当前用户的账号Id"
// @Param Body body model.SwagDeviceInfosReq true "获取设备的model"
// @Success 200 {object} model.SwagGetDeviceLocationResp "正确处理返回的数据"
// @Router /account_device_gps/{accountId} [post]
func GetDeviceLocation(c *gin.Context) {
	aidStr := c.Param("accountId")
	deviceInfosReq := &pb.DeviceInfosReq{}
	if err := c.BindJSON(deviceInfosReq); err != nil {
		logger.Debugf("json parse fail , error : %s", err)
		c.JSON(http.StatusBadRequest, err)
		return
	}

	// 使用session来校验用户
	aid, _ := strconv.Atoi(aidStr)
	if !service.ValidateAccountSession(c.Request, aid) {
		c.JSON(http.StatusUnauthorized, model.ErrorNotAuthSession)
		return
	}

	// 校验id值
	for _, id := range deviceInfosReq.DevicesIds {
		if !utils.CheckId(int(id)) {
			c.JSON(http.StatusBadRequest, errors.New("pramas error"))
			return
		}
	}

	res, err := grpc_pool.GrpcWebRpcCall(aid, context.Background(), deviceInfosReq, grpc_pool.GetDevicesInfo)
	//res, err := webCli.GetDevicesInfo(context.Background(), deviceInfosReq)
	if err != nil {
		logger.Error("Get Device info err with Grpc : %+v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"msg": "Get Device info fail, please try again later.", "code": 001})
		return
	}
	logger.Debugf("Get Device info success by Rpc: %+v and code:%d", res, res.(*pb.DeviceInfosResp).Res.Code)

	c.JSON(http.StatusCreated, gin.H{
		"Devices": res.(*pb.DeviceInfosResp).Devices,
	})

}

// @Summary 删除账户下级用户
// @Produce  json
// @Param accountId path string true "当前用户的账号Id"
// @Param deleteId path string true "删除用户下级的账号Id"
// @Param Authorization header string true "登录时返回的sessionId"
// @Success 200 {string} json "	{"msg":"delete account successfully.","result": "success"}"
// @Router /account_delete/{accountId}/{deleteId} [get]
func DeleteLeafNodeAccount(c *gin.Context) {
	accountId := c.Param("accountId")
	deleteId := c.Param("deleteId")

	logger.Debugln("deleteId:", deleteId, "accountId", accountId)
	//使用session来校验用户
	aid, _ := strconv.Atoi(accountId)
	dId, _ := strconv.Atoi(deleteId)

	if !service.ValidateAccountSession(c.Request, aid) {
		c.JSON(http.StatusUnauthorized, model.ErrorNotAuthSession)
		return
	}

	if !utils.CheckId(aid) || !utils.CheckId(dId) {
		c.JSON(http.StatusUnauthorized, gin.H{
			"msg": "params is invalid",
		})
		return
	}

	// 删除用户
	res, err := grpc_pool.GrpcWebRpcCall(aid, context.Background(), &pb.DeleteAccountReq{Id: int32(dId)}, grpc_pool.DeleteAccount)
	//res, err := webCli.DeleteAccount(context.Background(), &pb.DeleteAccountReq{Id: int32(dId)})
	if err != nil {
		logger.Error("Delete Account err with Grpc : %+v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"msg": "Delete Account fail, please try again later.", "code": 001})
		return
	}
	logger.Debugf("Delete Account success by grpc: %+v", res)

	if res.(*pb.DeleteAccountResp).Res.Code == http.StatusAccepted {
		c.JSON(http.StatusUnprocessableEntity, gin.H{
			"result": "error",
			"msg":    res.(*pb.DeleteAccountResp).Res.Msg,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"result": "success",
		"msg":    "delete account successfully.",
	})
}

// @Summary 获取账号名下所有的非设备成员
// @Produce  json
// @Param accountId path string true "当前用户的账号Id"
// @Param deleteId path string true "删除用户下级的账号Id"
// @Param Authorization header string true "登录时返回的sessionId"
// @Success 200 {string} json "	{"msg":"delete account successfully.","result": "success"}"
// @Router /account_junior/{accountId} [get]
func GetJuniorAccount(c *gin.Context) {
	accountId := c.Param("accountId")
	logger.Debugln("accountId", accountId)
	//使用session来校验用户
	aid, _ := strconv.Atoi(accountId)
	var (
		res []*model.JuniorAccount
		err error
	)
	if !service.ValidateAccountSession(c.Request, aid) {
		c.JSON(http.StatusUnauthorized, model.ErrorNotAuthSession)
		return
	}

	if !utils.CheckId(aid) {
		c.JSON(http.StatusUnauthorized, gin.H{
			"msg": "params is invalid",
		})
		return
	}

	if res, err = service.GetJuniorAccount(int32(aid)); err != nil {
		logger.Errorf("Get Junior Account fail: %+v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"msg": "Get Junior Account fail, please try again later."})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"junior_accounts": res,
	})
}


// TODO 切换调度员，webapi应该更新bitmap
func ChangeDispatcher(c *gin.Context){
	serverAddr := c.Query("server-addr")
	accountIdStr := c.Query("account-id")
	fmt.Println(serverAddr, accountIdStr)
	// TODO 校验参数是否合法
	accountId, _ := strconv.Atoi(accountIdStr)
	grpc_pool.GRPCManager.ChangeDispatcherServerId(accountId, serverAddr)
}
