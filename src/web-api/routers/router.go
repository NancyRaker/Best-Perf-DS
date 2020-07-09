/*
@Time : 2019/10/25 14:04 
@Author : yanKoo
@File : router
@Software: GoLand
@Description: 注册路由
*/
package routers

import (
	"github.com/gin-gonic/gin"
	"github.com/swaggo/gin-swagger"
	"github.com/swaggo/gin-swagger/swaggerFiles"
	_ "net/http/pprof"
	"web-api/controllers"
	_ "web-api/docs" // docs is generated by Swag CLI,
)


// 注册路由
func RegisterRouter(e *gin.Engine) {
	// swagger 文档
	e.GET("/web-api/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// 注册路由
	// account
	e.POST("/web-api/account/login.do/:account_name", controllers.SignIn)

	routerGroup := e.RouterGroup.Group("/web-api")
	routerGroup.POST("/account/logout.do", controllers.SignOut)
	routerGroup.POST("/account", controllers.CreateAccountBySuperior)
	routerGroup.GET("/account/:account-id", controllers.GetAccountInfo)
	routerGroup.POST("/account/info/update", controllers.UpdateAccountInfo)
	routerGroup.POST("/account/pwd/update", controllers.UpdateAccountPwd)
	routerGroup.GET("/account_class/:accountId/:searchId", controllers.GetAccountClass)
	routerGroup.GET("/account_delete/:accountId/:deleteId", controllers.DeleteLeafNodeAccount)
	routerGroup.GET("/account_device/:accountId/:getAdviceId", controllers.GetAccountDevice)
	routerGroup.POST("/account_device/:accountId", controllers.TransAccountDevice)
	routerGroup.POST("/account_device_gps/:accountId", controllers.GetDeviceLocation)
	routerGroup.GET("/account_devices/:accountId", controllers.GetDeviceList)
	routerGroup.GET("/account_junior/:accountId", controllers.GetJuniorAccount)
	routerGroup.POST("/account_clock/:account-id", controllers.SetAccountReportInfo)
	routerGroup.GET("/account_clock/:account-id", controllers.QueryAccountReportInfo)

	// group
	routerGroup.POST("/group", controllers.CreateGroup)
	routerGroup.POST("/group/update", controllers.UpdateGroup)
	routerGroup.POST("/group_manager/update", controllers.UpdateGroupManager)
	routerGroup.POST("/group/delete", controllers.DeleteGroup)
	routerGroup.POST("/group/devices/update", controllers.UpdateGroupDevice)
	routerGroup.GET("/group/info/:accountId/:groupId", controllers.GetGroupInfo) // 获取群组信息
	routerGroup.GET("/group/list/:accountId", controllers.GetGroupList)          // 获取群组列表信息

	// device
	routerGroup.POST("/device/import", controllers.ImportDeviceByRoot)
	routerGroup.POST("/device/update", controllers.UpdateDeviceInfo)
	routerGroup.GET("/device/clock/:account-id/:device-id/:start-timestamp", controllers.QueryDeviceClockStatus)

	// upload file
	routerGroup.POST("/upload", controllers.DispatcherUploadFile)
	routerGroup.POST("/upload/device", controllers.DeviceUploadFile)

	routerGroup.POST("/wifi/:accountId", controllers.OperationWifiInfo) // wifi 增删改
	routerGroup.GET("/wifi/:accountId", controllers.GetWifiInfo)        // 获取wifi信息

	// NFC tag标签
	routerGroup.POST("/tags/:accountId", controllers.OperationTagsInfo) // 标签 增删改
	routerGroup.GET("/tags/:accountId", controllers.GetTagsInfo)        // 获取标签信息

	// tag_tasks
	routerGroup.POST("/tag_tasks/:accountId", controllers.OperationTagTasksList)             // 标签任务 增删改
	routerGroup.GET("/tag_tasks/:account-id/:task-id", controllers.QueryTaskDetail)          // 单个任务详情查询
	routerGroup.GET("/tag_tasks_device/:account-id/:device-id", controllers.QueryDeviceTask) // 单个设备的任务查询

	routerGroup.GET("/trace/:accountId/:deviceId/:start/:end", controllers.GetTraceInfo) // 获取轨迹回放数据

	// im server
	e.GET("/im-server/:accountId", controllers.ImPush)

	// 获取日志信息(最新的一条)  // 内部调试
	routerGroup.GET("/device/log/:accountId/:deviceId", controllers.GetDeviceLog)
	routerGroup.GET("/internal/account_devices/:account-id", controllers.GetDeviceListByInternal)
	routerGroup.POST("/internal/device/update", controllers.UpdateDeviceInfoByInternal)

	// 热切换数据库引擎
	routerGroup.GET("/db/engine/change", controllers.ChangeDBEngine)
}