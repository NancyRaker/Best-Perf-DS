/**
* @Author: yanKoo
* @Date: 2019/3/11 10:48
* @Description: 处理请求的业务逻辑
 */
package controllers

import (
	"context"
	"database/sql"
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
	"time"
	tc "web-api/dao/customer"
	ss "web-api/dao/session"
	"web-api/engine/grpc_pool"
	"web-api/logger"
	"web-api/model"
	pb "web-api/proto/talk_cloud"
	"web-api/service"
	"web-api/utils"
)

func nowInMilli() int64 {
	return time.Now().UnixNano() / 1000000
}

// @Summary 登录
// @Description login by account name and pwd
// @Accept  json
// @Produce  json
// @Param account_name path string true "登录的用户名，eg：elephant"
// @Param body body model.AccountForSwag true "登录信息，username和pwd必填"
// @Success 200 {string} json "{"session_id": "1c2c46b8-f44a-4073-b219-d93d22dd2a43", "success": "true"}"
// @Router /account/login.do/{account_name} [post]
func SignIn(c *gin.Context) {
	// 1. 取出Post请求中的body内容
	time0 := time.Now()
	logger.Debugf("start singIn with %d", time0.UnixNano())
	signINBody := &model.Account{}
	if err := c.BindJSON(signINBody); err != nil {
		logger.Debugf("%s", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Request body is not correct.",
		})
		return
	}
	logger.Debugf("login params: %+v", signINBody)

	// 2. 验证body里面的名字和url的名字
	if c.Param("account_name") != signINBody.Username {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":      "User 和 url 不匹配",
			"error_code": "0021",
		})
		return
	}

	// 3. 对登录表单进行验证
	if !utils.CheckUserName(signINBody.Username) {
		logger.Debugln("Username format error")
		c.JSON(http.StatusUnprocessableEntity, gin.H{
			"error":      "用户名只能输入5-20个包含字母、数字或下划线的字串",
			"error_code": "0022",
		})
		return
	}

	// 校验密码
	if !utils.CheckPwd(signINBody.Pwd) {
		logger.Debugln("Pwd format error")
		c.JSON(http.StatusUnprocessableEntity, gin.H{
			"error":      "密码6位-16位，至少包含一个数字字母",
			"error_code": "0023",
		})
		return
	}

	time1 := time.Now()
	logger.Debugf("start get user password with %d, used: %d", time1, time0.Sub(time1))
	// 4. 数据库查询密码，看是否和发过来的相同
	uInfo, err := tc.GetAccount(signINBody.Username)
	if err != nil && err != sql.ErrNoRows {
		logger.Debugln("login db err:", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":      "process error, please try again.",
			"error_code": "003",
		})
		return
	}
	if err == sql.ErrNoRows {
		logger.Debugln("no found this user")
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":      "User does not exist.",
			"error_code": "0024",
		})
		return
	}

	logger.Debugf("Login pwd: %s, loginBody pwd is %s", uInfo.Pwd, signINBody.Pwd)
	if err != nil || uInfo.Pwd != signINBody.Pwd {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":      "User password is wrong.",
			"error_code": "0025",
		})
		return
	}

	time2 := time.Now()
	logger.Debugf("start check already login with %d, used: %d", time2, time1.Sub(time2))
	// 新加登录逻辑，判断是否有人在登录 判断有没有的没过期的sessionId
	if signINBody.ForceLogin != "f" && ss.ExistSession("", uInfo.Id) {
		// 如果有存在session，就返回，让用户判断是否继续登录
		c.JSON(http.StatusConflict, gin.H{
			"error":      "User is already login, continue?",
			"error_code": "0026",
		})
		return
	}

	// 5. 插入session 因为，你已经login请求了，说明新建立一个session，直接更新session, 但是需要覆盖原来的sessionId
	time3 := time.Now()
	logger.Debugf("start insertSessionInfo with %d, used: %d", time3, time2.Sub(time3))
	sId, err := insertSessionInfo(uInfo)
	if err != nil {
		logger.Debugln("login update session err: ", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Process error, please try again later.",
		})
		return
	}

	time4 := time.Now()
	logger.Debugf("signIn end with %d, used: %d", time4, time1.Sub(time4))
	// 6. 返回登录成功的消息
	c.JSON(http.StatusOK, gin.H{
		//"success":    "true",
		"account_id": uInfo.Id,
		"session_id": sId,
	})
}

// 更新session 主要是登录的时候用  TODO 得要判断是不是重复登录，去踢人下线
func insertSessionInfo(aInfo *model.Account) (string, error) {
	newSid, _ := utils.NewUUID()
	ct := nowInMilli()
	ttl := ct + 30*60*1000 // Severside session valid time: 30 min

	ttlStr := strconv.FormatInt(ttl, 10)
	sInfo := &model.SessionInfo{
		SessionID: newSid,
		UserName:  aInfo.Username,
		UserPwd:   aInfo.Pwd,
		AccountId: aInfo.Id,
		TTL:       ttlStr,
	}

	// 先判断有没有的没过期的sessionId
	if ss.ExistSession(sInfo.SessionID, sInfo.AccountId) {
		// 发送消息给IM, 通知web下线
		_, err := grpc_pool.GrpcAppRpcCall(sInfo.AccountId, context.Background(), &pb.ImMsgReqData{
			Id:           int32(sInfo.AccountId),
			ReceiverId:   0,
			ReceiverType: 1,
			MsgType:      WEB_REPEATED_LOGIN,
			MsgCode:      strconv.FormatInt(time.Now().Unix(), 10),
		}, grpc_pool.ImMessagePublish, true)
		if err != nil {
			logger.Error("repeated login err with Grpc : %+v", err)
			return "", err
		}
	} else {
		// 检查一遍grpc地址本地内存对不对
		logger.Debugln("check")
		cli := grpc_pool.GRPCManager.GetGRPCConnClientById(sInfo.AccountId, true/*需要check grpc地址*/)
		if cli != nil {
			cli.Close()
		}
	}

	if err := ss.InsertSession(sInfo); err != nil {
		return "", err
	}

	return newSid, nil
}

// @Summary 退出
// @Description logout by account name and pwd, 请求头中Authorization参数设置为登录时返回的sessionId
// @Accept  json
// @Produce  json
// @Param Authorization header string true "登录时返回的sessionId eg:1c2c46b8-f44a-4073-b219-d93d22dd2a43"
// @Param body body model.AccountForSwag true "登录信息，username和id必填"
// @Success 200 {string} json "{"success":"true","msg": "SignOut is successful"}"
// @Router /account/logout.do [post]
func SignOut(c *gin.Context) {
	// 1. 取出body中的内容
	signOutBody := &model.Account{}
	if err := c.BindJSON(signOutBody); err != nil {
		logger.Debugf("%s", err)
		c.JSON(http.StatusBadRequest, model.ErrorRequestBodyParseFailed)
		return
	}

	// 2. 验证session
	if !service.ValidateAccountSession(c.Request, signOutBody.Id) {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":      "Session is not exist.",
			"error_code": "401",
		})
		return
	}

	// 3. 删除session
	if err := service.DeleteSessionInfo(service.GetSessionId(c.Request), signOutBody); err != nil {
		logger.Debugf("SignOut delete session db error : %s", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":      "Process error, please try again later.",
			"error_code": "003",
		})
		return
	}

	// 通知im断开grpcstream连接
	// 发送消息给IM, 通知web下线
	//webCli := pb.NewTalkCloudClient(conn)
	//res, err := webCli.ImMessagePublish(context.Background(), &pb.ImMsgReqData{
	res, err := grpc_pool.GrpcAppRpcCall(signOutBody.Id, context.Background(), &pb.ImMsgReqData{
		Id:           int32(signOutBody.Id),
		ReceiverId:   -1,
		ReceiverType: 1,
		MsgType:      IM_REFRESH_WEB, // 本来是刷新界面来关闭stream，这里效果一样
		MsgCode:      strconv.FormatInt(time.Now().Unix(), 10),
	}, grpc_pool.ImMessagePublish)
	logger.Debugf("logout rpc res:%+v, err: %+v", res.(*pb.ImMsgRespData), err)
	if err != nil {
		logger.Debugf("logout err with Grpc : %+v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":      "Process error, please try again later.",
			"error_code": "003",
		})
		return
	}

	// 4. 返回消息
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"msg":     "SignOut is successful",
	})
}

