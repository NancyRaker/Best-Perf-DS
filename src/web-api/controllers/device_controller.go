/*
@Time : 2019/3/29 15:33
@Author : yanKoo
@File : DeviceController
@Software: GoLand
@Description: 超级管理员导入设备，调用mysql的GRPC的server端的方法
*/
package controllers

import (
	"context"
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
	"web-api/engine/grpc_pool"
	"web-api/logger"
	"web-api/model"
	pb "web-api/proto/talk_cloud"
	"web-api/service"
	"web-api/utils"
)

// @Summary 导入设备
// @Produce  json
// @Param Authorization header string true "登录时返回的sessionId"
// @Param account_name path string true "当前用户的账号"
// @Param Body body model.SwagAccountImportDeviceReq true "导入设备的model"
// @Success 200 {object} model.SwagImportDeviceByRootResp "处理正确返回的数据"
// @Router /device/import [post]
func ImportDeviceByRoot(c *gin.Context) {
	aiDReq := &model.AccountImportDeviceReq{}
	if err := c.BindJSON(aiDReq); err != nil {
		logger.Debugf("%s", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"error":      "Request body is not correct.",
			"error_code": "001",
		})
		return
	}

	// 使用session来校验用户
	if !service.ValidateAccountSession(c.Request, int(aiDReq.AccountId)) {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":      "session is not right.",
			"error_code": "006",
		})
		return
	}
	logger.Debugf("%+v", aiDReq)

	var errDevices []*model.Device
	var duliDevices []*model.Device
	var errIdx int
	var duliIdx int
	dinfo := make([]*pb.DeviceInfo, 0)
	for _, v := range aiDReq.Devices {
		// 校验imei
		if utils.CheckIMei(v.IMei) {
			// imei查重
			if r, err := grpc_pool.GrpcWebRpcCall(int(aiDReq.AccountId), context.Background(), &pb.ImeiReq{Imei: v.IMei}, grpc_pool.SelectDeviceByImei); err != nil {
				//if r, err := webCli.SelectDeviceByImei(context.Background(), &pb.ImeiReq{Imei: v.IMei}); err != nil {
				logger.Debugln("Select id by imei with error in web: ", err)
			} else {
				if r.(*pb.ImeiResp).Id > 0 {
					v.Id = duliIdx
					duliDevices = append(duliDevices, v)
					duliIdx++
					continue
				}
				dinfo = append(dinfo, &pb.DeviceInfo{
					Imei:       v.IMei,
					DeviceType: v.DeviceType,
					ActiveTime: v.ActiveTime,
					SaleTime:   v.SaleTime,
				})
			}
		} else {
			v.Id = errIdx
			errDevices = append(errDevices, v)
			errIdx++
		}

	}
	if len(dinfo) == 0 {
		// 返回格式不正确的数据
		c.JSON(http.StatusOK, gin.H{
			"error":        "Import some device error, Please try again later.",
			"err_devices":  errDevices,
			"deli_devices": duliDevices,
			"error_code":   "422",
		})
		return
	}

	logger.Debugln("ImportDeviceByRoot start rpc")
	res, err := grpc_pool.GrpcWebRpcCall(int(aiDReq.AccountId), context.Background(), &pb.ImportDeviceReq{
		AccountId: aiDReq.Receiver,
		Devices:   dinfo,
	}, grpc_pool.ImportDeviceByRoot)
	//res, err := webCli.ImportDeviceByRoot(context.Background(), &pb.ImportDeviceReq{
	//	AccountId: aiDReq.Receiver,
	//	Devices:   dinfo,
	//})
	if err != nil {
		logger.Debugln("Import device error : ", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":      "Import device error, Please try again later.",
			"error_code": "500",
			"msg":        err,
		})
		return
	}

	if len(errDevices) != 0 {
		// 返回格式不正确的数据
		c.JSON(http.StatusOK, gin.H{
			"error":      "Import some device error, Please try again later.",
			"devices":    errDevices,
			"error_code": "422",
		})

		if len(dinfo) == 0 {
			return
		}
	} else {
		c.JSON(int(res.(*pb.ImportDeviceResp).Result.Code), gin.H{
			"err_devices":  errDevices,
			"deli_devices": duliDevices,
			"msg":          res.(*pb.ImportDeviceResp).Result.Msg,
		})
	}
}

// @Summary 转移设备
// @Produce  json
// @Param Authorization header string true "登录时返回的sessionId"
// @Param accountId path string true "当前用户的账号Id"
// @Param Body body model.SwagAccountDeviceTransReq true "转移设备的model"
// @Success 200 {object} model.SwagTransAccountDeviceResp "处理正确返回的数据"
// @Router /account_device/{accountId} [post]
func TransAccountDevice(c *gin.Context) {
	aidStr := c.Param("accountId")
	accountDevices := &model.AccountDeviceTransReq{}
	if err := c.BindJSON(accountDevices); err != nil {
		logger.Debugf("json parse fail , error : %s", err)
		c.JSON(http.StatusBadRequest, model.ErrorRequestBodyParseFailed)
		return
	}

	// 使用session来校验用户
	aid, _ := strconv.Atoi(aidStr)
	if !service.ValidateAccountSession(c.Request, aid) {
		c.JSON(http.StatusUnauthorized, model.ErrorNotAuthSession)
		return
	}

	// IMEI号只能是15位数字
	// 结构体为空
	if accountDevices.Devices == nil {
		c.JSON(http.StatusBadRequest, model.ErrorRequestBodyParseFailed)
		return
	}
	for _, v := range accountDevices.Devices {
		if v.IMei == "" {
			c.JSON(http.StatusUnprocessableEntity, gin.H{
				"error":      "Imei is not correct.",
				"error_code": "001",
			})
			return
		}
	}
	if accountDevices.Receiver.AccountId == 0 {
		c.JSON(http.StatusUnprocessableEntity, gin.H{
			"error":      "receiver id can't be empty.",
			"error_code": "001",
		})
		return
	}
	var imeis = make([]string, 0)
	for _, d := range accountDevices.Devices {
		imeis = append(imeis, d.IMei)
	}

	// 转移设备
	//if _, err := webCli.MultiTransDevice(context.TODO(), &pb.TransDevices{Imeis: imeis,
	if _, err := grpc_pool.GrpcWebRpcCall(aid, context.TODO(), &pb.TransDevices{Imeis: imeis,
		ReceiverId: int32(accountDevices.Receiver.AccountId), SenderId: int32(accountDevices.Devices[0].AccountId)}, grpc_pool.MultiTransDevice); err != nil {
		logger.Debugf("db error : %s", err)
		c.JSON(http.StatusInternalServerError, model.ErrorDBError)
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"result": "success",
		"msg":    "trans successful",
	})
}

// @Summary 更新设备信息
// @Produce  json
// @Param Authorization header string true "登录时返回的sessionId"
// @Param account_name path string true "当前用户的账号"
// Param Body body model.SwagDeviceUpdate true "更新设备信息的model"
// @Success 200 {object} model.SwagUpdateDeviceInfoResp "处理正确返回的数据"
// @Router /device/update [post]
func UpdateDeviceInfo(c *gin.Context) {
	d := &pb.DeviceUpdate{}
	if err := c.BindJSON(d); err != nil {
		logger.Debugf("%s", err)
		c.JSON(http.StatusUnprocessableEntity, gin.H{
			"error":      "Request body is not correct.",
			"error_code": "001",
		})
		return
	}

	// 校验参数信息 ：校首先必须要有id，其次是每个参数的合法性，首先都不允许为空
	if d.LoginId == 0 {
		logger.Debugf("account id is nil")
		c.JSON(http.StatusUnprocessableEntity, gin.H{
			"error":      "The account id cannot be empty",
			"error_code": "003",
		})
		return
	}

	// 使用session来校验用户
	if !service.ValidateAccountSession(c.Request, int(d.LoginId)) {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":      "session is not right.",
			"error_code": "006",
		})
		return
	}

	logger.Debugf("%+v", d)
	logger.Debugln("UpdateDeviceInfo start rpc")
	//res, err := webCli.UpdateDeviceInfo(context.Background(), &pb.UpdDInfoReq{DeviceInfo: d})
	res, err := grpc_pool.GrpcWebRpcCall(int(d.LoginId), context.Background(), &pb.UpdDInfoReq{DeviceInfo: d}, grpc_pool.UpdateDeviceInfo)
	if err != nil {
		logger.Debugln("Update device error : ", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":      "Update device error, Please try again later.",
			"error_code": "500",
		})
		return
	}

	c.JSON(int(res.(*pb.UpdDInfoResp).Res.Code), gin.H{
		"msg": res.(*pb.UpdDInfoResp).Res.Msg,
	})
}

// @Summary 获取wifi信息
// @Produce  json
// @Param accountId path string true "当前用户的账号Id"
// @Param Authorization header string true "登录时返回的sessionId"
// @Success  200 {object} model.SwagGetDeviceLogResp "处理正确返回的数据"
// @Router /device/log/{accountId}/{deviceId} [get]
func GetDeviceLog(c *gin.Context) {
	var (
		err     error
		aIdStr  = c.Param("accountId")
		uIdStr  = c.Param("deviceId")
		logInfo interface{}
	)

	// 使用session来校验用户
	aid, _ := strconv.Atoi(aIdStr)
	uId, _ := strconv.Atoi(uIdStr)
	if !service.ValidateAccountSession(c.Request, aid) {
		c.JSON(http.StatusUnauthorized, model.ErrorNotAuthSession)
		return
	}

	//if logInfo, err = webCli.GetDeviceLogInfo(context.TODO(), &pb.GetDeviceLogInfoReq{Uid: int32(uId)}); err != nil {
	if logInfo, err = grpc_pool.GrpcWebRpcCall(aid, context.TODO(), &pb.GetDeviceLogInfoReq{Uid: int32(uId)}, grpc_pool.GetDeviceLogInfo); err != nil {
		logger.Errorf("Get Log info by grpc with error: %+v", err)
		c.JSON(http.StatusInternalServerError, model.ErrorInternalServerError)
		return
	}

	c.JSON(http.StatusOK, gin.H{"log_url": logInfo.(*pb.GetDeviceLogInfoResp).LogUrl})
}

// @Summary 获取轨迹回放数据
// @Produce  json
// @Param accountId path string true "当前用户的账号Id"
// @Param Authorization header string true "登录时返回的sessionId"
// @Router /trace/{accountId}/{deviceId}/{start-timestamp}/{end-timestamp} [get]
func GetTraceInfo(c *gin.Context) {
	var (
		err            error
		aidStr         = c.Param("accountId")
		deviceIdStr    = c.Param("deviceId")
		timestampStart = c.Param("start")
		timestampEnd   = c.Param("end")
		traceInfoData  interface{} //*pb.GpsForTraceResp
	)
	if aidStr == "" || deviceIdStr == "" || timestampStart == "" || timestampEnd == "" {
		c.JSON(http.StatusUnprocessableEntity, model.ErrorRequestBodyParseFailed)
		return
	}

	// 使用session来校验用户
	aid, _ := strconv.Atoi(aidStr)
	if !service.ValidateAccountSession(c.Request, aid) {
		c.JSON(http.StatusUnauthorized, model.ErrorNotAuthSession)
		return
	}

	//if traceInfoData, err = webCli.GetGpsForTrace(context.TODO(), &pb.GpsForTraceReq{
	if traceInfoData, err = grpc_pool.GrpcWebRpcCall(aid, context.TODO(), &pb.GpsForTraceReq{
		Id: deviceIdStr, TimesStampStart: timestampStart, TimesStampEnd: timestampEnd}, grpc_pool.GetGpsForTrace); err != nil {
		logger.Errorf("Get WifiInfo by grpc with error: %+v", err)
		c.JSON(http.StatusInternalServerError, model.ErrorInternalServerError)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"trace_data": traceInfoData.(*pb.GpsForTraceResp).TraceData,
		"whole":      traceInfoData.(*pb.GpsForTraceResp).Whole,
	})
}

// @Summary 获取设备列表，web调度员界面展示的列表 //TODO 修改为grpc调用去im获取数据
// @Produce  json
// @Param accountId path string true "当前用户的账号Id"
// @Param Authorization header string true "登录时返回的sessionId"
// @Router /account_devices/{accountId} [get]
func GetDeviceList(c *gin.Context) {
	var (
		err          error
		accountIdStr = c.Param("accountId")
	)
	if accountIdStr == "" {
		c.JSON(http.StatusUnprocessableEntity, model.ErrorRequestBodyParseFailed)
		return
	}

	// 使用session来校验用户
	accountId, _ := strconv.Atoi(accountIdStr)
	if !service.ValidateAccountSession(c.Request, accountId) {
		c.JSON(http.StatusUnauthorized, model.ErrorNotAuthSession)
		return
	}
	var (
		ctx, cancelFunc = context.WithCancel(context.TODO())
		errOut          = make(chan interface{}, 10)
	)
	// 4.
	deviceAll, statMap, err := service.GetDevicesForDispatchers(ctx, cancelFunc, accountId, errOut)
	if err != nil {
		logger.Debugf("GetAccountInfo fail with : %+v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"err": "Internal server error, please try again."})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"device_list": *deviceAll,
		"device_list_info": struct {
			DeviceTotal int `json:"device_total"`
			OnlineTotal int `json:"online_total"`
		}{DeviceTotal: len(*deviceAll), OnlineTotal: len(statMap)},
	})
}


// 项目内部控制控制设备
func GetDeviceListByInternal(c *gin.Context) {
	var (
		err          error
		accountIdStr = c.Param("account-id")
	)
	if accountIdStr == "" {
		c.JSON(http.StatusUnprocessableEntity, model.ErrorRequestBodyParseFailed)
		return
	}

	//// 使用session来校验用户
	accountId, _ := strconv.Atoi(accountIdStr)
	//if !service.ValidateAccountSession(c.Request, accountId) {
	//	c.JSON(http.StatusUnauthorized, model.ErrorNotAuthSession)
	//	return
	//}
	var (
		ctx, cancelFunc = context.WithCancel(context.TODO())
		errOut          = make(chan interface{}, 10)
	)
	// 4.
	deviceAll, statMap, err := service.GetDevicesForDispatchers(ctx, cancelFunc, accountId, errOut)
	if err != nil {
		logger.Debugf("GetAccountInfo fail with : %+v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"err": "Internal server error, please try again."})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"device_list": *deviceAll,
		"device_list_info": struct {
			DeviceTotal int `json:"device_total"`
			OnlineTotal int `json:"online_total"`
		}{DeviceTotal: len(*deviceAll), OnlineTotal: len(statMap)},
	})
}


func UpdateDeviceInfoByInternal(c *gin.Context) {
	d := &pb.DeviceUpdate{}
	if err := c.BindJSON(d); err != nil {
		logger.Debugf("%s", err)
		c.JSON(http.StatusUnprocessableEntity, gin.H{
			"error":      "Request body is not correct.",
			"error_code": "001",
		})
		return
	}

	// 校验参数信息 ：校首先必须要有id，其次是每个参数的合法性，首先都不允许为空
	if d.LoginId == 0 {
		logger.Debugf("account id is nil")
		c.JSON(http.StatusUnprocessableEntity, gin.H{
			"error":      "The account id cannot be empty",
			"error_code": "003",
		})
		return
	}

	// 使用session来校验用户
	/*if !service.ValidateAccountSession(c.Request, int(d.LoginId)) {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":      "session is not right.",
			"error_code": "006",
		})
		return
	}*/

	logger.Debugf("%+v", d)
	logger.Debugln("UpdateDeviceInfo start rpc")
	//res, err := webCli.UpdateDeviceInfo(context.Background(), &pb.UpdDInfoReq{DeviceInfo: d})
	res, err := grpc_pool.GrpcWebRpcCall(int(d.LoginId), context.Background(), &pb.UpdDInfoReq{DeviceInfo: d}, grpc_pool.UpdateDeviceInfo)
	if err != nil {
		logger.Debugln("Update device error : ", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":      "Update device error, Please try again later.",
			"error_code": "500",
		})
		return
	}

	c.JSON(int(res.(*pb.UpdDInfoResp).Res.Code), gin.H{
		"msg": res.(*pb.UpdDInfoResp).Res.Msg,
	})
}
