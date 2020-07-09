/*
@Time : 2019/6/20 10:22 
@Author : yanKoo
@File : wifi
@Software: GoLand
@Description:
*/
package controllers

import (
	"context"
	"errors"
	"fmt"
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

var (
	SAVE_WIFI_INFO   int32 = 1
	UPDATE_WIFI_INFO int32 = 2
	DELETE_WIFI_INFO int32 = 3
)

//
// @Summary 增加，删除，修改wifi信息
// @Produce  json
// @Param Authorization header string true "登录时返回的sessionId"
// @Param accountId path string true "当前用户的账号Id"
// @Param Body body model.SwagWifiInfoReq true "获取wifi的model"
// @Success 200 {object} model.SwagGetDeviceLocationResp "正确处理返回的数据"
// @Router /wifi/{accountId} [post]
func OperationWifiInfo(c *gin.Context) {
	aidStr := c.Param("accountId")
	accountWifiInfo := &pb.WifiInfoReq{}
	if err := c.BindJSON(accountWifiInfo); err != nil {
		logger.Debugf("json parse fail , error : %s", err)
		c.JSON(http.StatusBadRequest, model.ErrorRequestBodyParseFailed)
		return
	}

	// 校验ops参数
	if accountWifiInfo.Ops < 1 || accountWifiInfo.Ops > 3 {
		logger.Errorf("accountWifi Info ops error: %d", accountWifiInfo.Ops)
		c.JSON(http.StatusBadRequest, model.ErrorRequestBodyParseFailed)
		return
	}

	// 使用session来校验用户
	aid, _ := strconv.Atoi(aidStr)
	if !service.ValidateAccountSession(c.Request, aid) {
		c.JSON(http.StatusUnauthorized, model.ErrorNotAuthSession)
		return
	}
	accountWifiInfo.AccountId = int32(aid)
	logger.Errorf("accountWifi Info : %+v", accountWifiInfo)
	switch accountWifiInfo.Ops {
	case SAVE_WIFI_INFO:
		saveWifiInfo(c, accountWifiInfo)
	case UPDATE_WIFI_INFO:
		updateWifiInfo(c, accountWifiInfo)
	case DELETE_WIFI_INFO:
		deleteWifiInfo(c, accountWifiInfo)
	}
}

// 保存wifi信息
func saveWifiInfo(c *gin.Context, wifiInfos *pb.WifiInfoReq) {
	var err error
	// 校验数据
	for _, wifi := range wifiInfos.Wifis {
		// 校验bssId
		fmt.Println(wifi.BssId)
		if !utils.CheckBssId(wifi.BssId) {
			logger.Debugf("saveWifiInfo validate BssId :%s fail", wifi.BssId)
			err = errors.New("MAC addr is invalid")
			break
		}

		// 校验经度
		fmt.Println(wifi.Longitude)
		if !utils.CheckLongitude(strconv.FormatFloat(wifi.Longitude, 'f', 6, 64)) {
			err = errors.New("longitude is invalid")
			break
		}

		// 校验纬度
		if !utils.CheckLatitude(strconv.FormatFloat(wifi.Latitude, 'f', 6, 64)) {
			err = errors.New("latitude is invalid")
			break
		}

		// 校验位置信息
		if wifi.Des == "" {
			err = errors.New("wifi description is invalid")
			break
		}
	}

	if err != nil {
		logger.Debugf("saveWifiInfo validate wifi info fail with error: %+v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error msg": err.Error()})
		return
	}

	wifiInfoProcess(c, wifiInfos)
}

// 修改wifi信息
func updateWifiInfo(c *gin.Context, wifiInfos *pb.WifiInfoReq) {
	var err error
	// 校验数据
	for _, wifi := range wifiInfos.Wifis {
		// 校验bssId
		if !utils.CheckBssId(wifi.BssId) {
			err = errors.New("MAC addr is invalid")
			break
		}

		// 校验经度
		if !utils.CheckLongitude(strconv.FormatFloat(wifi.Longitude, 'f', 6, 64)) {
			err = errors.New("longitude is invalid")
			break
		}

		// 校验纬度
		if !utils.CheckLatitude(strconv.FormatFloat(wifi.Latitude, 'f', 6, 64)) {
			err = errors.New("update WifiInfo latitude is invalid")
			break
		}

		// 校验位置信息
		if wifi.Des == "" {
			err = errors.New("wifi description is invalid")
			break
		}
	}
	if err != nil {
		logger.Debugf("validate wifi info fail with error: %+v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error msg": err.Error()})
		return
	}

	wifiInfoProcess(c, wifiInfos)
}

// 删除wifi信息
func deleteWifiInfo(c *gin.Context, wifiInfos *pb.WifiInfoReq) {
	var err error
	// 校验数据
	for _, wifi := range wifiInfos.Wifis {
		// 校验bssId
		fmt.Println(wifi.BssId)
		if !utils.CheckBssId(wifi.BssId) {
			logger.Debugf("saveWifiInfo validate BssId :%s fail", wifi.BssId)
			err = errors.New("MAC addr is invalid")
			break
		}
		// 校验id
		if wifi.Id <= 0 {
			err = errors.New("wifi id must ge 0")
			break
		}
	}
	if err != nil {
		logger.Debugf("validate wifi info fail with error: %+v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error msg": err.Error()})
		return
	}

	wifiInfoProcess(c, wifiInfos)
}

// grpc调用，提交信息
func wifiInfoProcess(c *gin.Context, wifiInfos *pb.WifiInfoReq) {
	var (
		res interface{} //*pb.WifiInfoResp
		err error
	)
	aidStr := c.Param("accountId")
	aid, _ := strconv.Atoi(aidStr)
	if res, err = grpc_pool.GrpcWebRpcCall(aid, context.TODO(), wifiInfos, grpc_pool.PostWifiInfo); err != nil {
	//if res, err = webCli.PostWifiInfo(context.TODO(), wifiInfos); err != nil {
		logger.Errorf("operate wifi info by grpc with error: %+v", err)
		c.JSON(http.StatusInternalServerError, model.ErrorInternalServerError)
		return
	}

	if res.(*pb.WifiInfoResp).Res.Code != http.StatusOK {
		c.JSON(http.StatusBadRequest, gin.H{"msg":res.(*pb.WifiInfoResp).Res.Msg})
		return
	}

	c.JSON(http.StatusOK, model.SuccessResponse)
}

// @Summary 获取wifi信息
// @Produce  json
// @Param accountId path string true "当前用户的账号Id"
// @Param Authorization header string true "登录时返回的sessionId"
// @Success 200 {string} json "	{"msg":"delete account successfully.","result": "success"}"
// @Router /wifi/{accountId} [get]
func GetWifiInfo(c *gin.Context) {
	var  (
		err error
     	aidStr = c.Param("accountId")
     	wifis interface{} // *pb.GetWifiInfoResp
	)
	// 使用session来校验用户
	aid, _ := strconv.Atoi(aidStr)
	if !service.ValidateAccountSession(c.Request, aid) {
		c.JSON(http.StatusUnauthorized, model.ErrorNotAuthSession)
		return
	}

	//if wifis, err = webCli.GetWifiInfo(context.TODO(), &pb.GetWifiInfoReq{AccountId:int32(aid)}); err != nil {
	if wifis, err = grpc_pool.GrpcWebRpcCall(aid, context.TODO(), &pb.GetWifiInfoReq{AccountId:int32(aid)}, grpc_pool.GetWifiInfo); err != nil {
		logger.Errorf("Get WifiInfo by grpc with error: %+v", err)
		c.JSON(http.StatusInternalServerError, model.ErrorInternalServerError)
		return
	}

	c.JSON(http.StatusOK, gin.H{"msg":wifis.(*pb.GetWifiInfoResp)})
}
