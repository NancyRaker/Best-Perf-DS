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
	"sync"
	"web-ips/bean"
	"web-ips/dao/customer"
	"web-ips/logger"
	"web-ips/model"
	pb "web-ips/proto/talk_cloud"
	"web-ips/service"
)

// 返回设备的服务的ip和端口
// url：/sever/addr/?imei=123456789123456&version=v1&ip=127.0.0.1
func GetServerAddr(c *gin.Context) {
	// 根据scr鉴权
	/*if c.Param("scr") != "jimi" {
		c.JSON(http.StatusUnauthorized, gin.H{"res": "valid user"})
		return
	}*/

	logger.Debugf("access with : imei:%s, verison:%s, ip:%s",
		c.Query("imei"), c.Query("version"), c.Query("ip"))

	serverAddr := service.ChooseServerIp(c.Query("imei"))

	c.JSON(http.StatusOK, gin.H{"server": serverAddr})

	var out interface{}
	if serverAddr != nil {
		out = serverAddr.Grpc
	} else {
		out = serverAddr
	}
	logger.Debugf("imei:%s, ip:%s will conn im: %+v",
		c.Query("imei"), c.Query("ip"), out)
}

// 返回非设备账号的服务的ip和端口
// url：/sever/addr/?imei=123456789123456&version=v1&ip=127.0.0.1
func GetDispatcherServerAddr(c *gin.Context) {
	// 根据scr鉴权
	/*if c.Param("scr") != "jimi" {
		c.JSON(http.StatusUnauthorized, gin.H{"res": "valid user"})
		return
	}*/

	logger.Debugf("access with :  account-id:%s, account-name:%s, version:%s, ip:%s",
		c.Query("account-id"), c.Query("account-name"),
		c.Query("version"), c.Query("ip"))

	var serverAddr *model.ServerResp
	accountName := c.Query("account-name")
	accountIdStr := c.Query("account-id")
	if accountName == "" && accountIdStr == "" {
		err := "GetDispatcherServerAddr params is invalid"
		logger.Debugln(err)
		c.JSON(http.StatusUnprocessableEntity, gin.H{"msg": err})
		return
	} else if accountIdStr != "" { // 优先使用id去查询
		accountId, err := strconv.Atoi(accountIdStr)
		if err != nil {
			c.JSON(http.StatusUnprocessableEntity, gin.H{"msg": err})
			return
		}
		serverAddr = service.ChooseServerIpForDispatcher(accountId)
	} else { // 否则用accountName去查询
		serverAddr = service.ChooseServerIpForDispatcher(accountName)
	}

	c.JSON(http.StatusOK, gin.H{"server": serverAddr})

	var out interface{}
	if serverAddr != nil {
		out = serverAddr.Grpc
	} else {
		out = serverAddr
	}
	logger.Debugf("account-name:%s, ip:%s will conn im: %+v",
		c.Query("account-name"), c.Query("ip"), out)
}

// 切换服务 /server/change TODO 对请求ip进行限制，只允许微软云和东莞进行通信
func ChangeServer(c *gin.Context) {
	var (
		m sync.Mutex
	)
	m.Lock()
	defer m.Unlock() // TODO 两个操作不是原子的，会出现并发问题

	serverCodeStr := c.Query("server-code")
	accountNameStr := c.Query("account-name")
	if serverCodeStr == "" || accountNameStr == "" {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"err": "invalid param"})
		return
	}
	// TODO 校验参数是否合法
	serverCode, _ := strconv.Atoi(serverCodeStr)
	var tempId int
	if tempId = bean.AccountMap.Get(accountNameStr); tempId == -1 {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"err": "no dispatcher"})
		return
	}
	// 1. 首先查account-id在哪个map下
	id := uint(tempId)
	var flagId = - 1
	logger.Debugf("%d will change server code: %d", serverCode, id)
	for i := 0; i < len(service.ServerNodeList); i++ {
		if service.ServerNodeList[i].BitMapNode.IsExist(id) {
			flagId = i
			service.ServerNodeList[i].BitMapNode.Remove(id)
		}
	}

	// 2. 设置调度员对应的服务map
	if serverCode <= len(service.ServerNodeList) && serverCode > 0 {
		service.ServerNodeList[serverCode-1].BitMapNode.Add(id)
	}

	// TODO 修改数据库customer表中的server_addr字段
	err := customer.UpdateCustomer([]int{serverCode}, []int{tempId})
	if err != nil {
		logger.Errorf("UpdateCustomer fail err: %+v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"err": "process error"})
		return
	}
	// 发送https请求通知web-api更新调度员 // 不通知了，等调度员再次登录的时候自己来访问

	c.JSON(http.StatusOK, gin.H{"res": "success"})
	// 调用GRPC方法，让im通知设备和调度员下线从新获取ips
	if flagId != -1 {
		logger.Debugf("%+v", service.ServerNodeList[flagId].Ip)
		resp, err := service.ServerNodeList[flagId].ConnPoolNode.RpcCall(context.TODO(), &pb.NotifyIpsChangedReq{AccountId: int32(id)}, "NotifyIpsChanged")
		if err != nil {
			logger.Debugf("ChangeServer rpc call err %+v", err)
			//c.JSON(http.StatusInternalServerError, gin.H{"err": "process error"})
			//return
		}
		logger.Debugf("ChangeServer rpc call resp: %+v", resp)
	}
}

// TODO 设备更换调度员
func ChangeDispatcher(c *gin.Context) {
	var devicesReq = &model.DeviceChange{}
	if err := c.Bind(devicesReq); err != nil {
		logger.Debugf("ChangeDispatcher error: %+v", err)
		c.JSON(http.StatusUnprocessableEntity, gin.H{"msg": "param is invalid"})
		return
	}
	// TODO 这个接口可不能随便调
	for _, imei := range devicesReq.IMeis {
		service.DeviceTrie.ChangeAccount(imei, devicesReq.AccountId)
	}

	c.JSON(http.StatusOK, gin.H{"res": "success"})
}

// 添加设备
func AddDeviceForDispatcher(c *gin.Context) {
	var (
		devicesReq = &model.DeviceAdd{}
	)
	if err := c.Bind(devicesReq); err != nil {
		logger.Debugf("AddDeviceForDispatcher error: %+v", err)
		c.JSON(http.StatusUnprocessableEntity, gin.H{"msg": "param is invalid"})
		return
	}
	logger.Debugf("%+v", devicesReq)
	// TODO 这个接口可不能随便调
	for _, imei := range devicesReq.IMeis {
		service.DeviceTrie.Insert(imei, devicesReq.AccountId)
	}

	c.JSON(http.StatusOK, gin.H{"res": "success"})
}

// 添加调度员
func AddDispatcher(c *gin.Context) {
	var (
		dispatcher = &model.DispatcherAdd{}
	)
	if err := c.Bind(dispatcher); err != nil {
		logger.Debugf("AddDeviceForDispatcher error: %+v", err)
		c.JSON(http.StatusUnprocessableEntity, gin.H{"msg": "param is invalid"})
		return
	}
	// TODO 校验参数
	logger.Debugf("Add Dispatcher with:%+v", dispatcher)

	// 1. 找到创建这个账户的人在哪个map，然后再去更新map
	for i := 0; i < len(service.ServerNodeList); i++ {
		if service.ServerNodeList[i].BitMapNode.IsExist(dispatcher.CreatorId) {
			service.ServerNodeList[i].BitMapNode.Add(dispatcher.AccountId)
			bean.AccountMap.Set(dispatcher.AccountName, dispatcher.AccountId)
		}
	}

	c.JSON(http.StatusOK, gin.H{"res": "success"})
}
