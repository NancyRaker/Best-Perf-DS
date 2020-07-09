/**
* @Author: yanKoo
* @Date: 2019/3/11 11:03
* @Description: 用来定义返回错误的结构体和消息格式 还有一个操作成功的消息
 */
package model

import "github.com/gin-gonic/gin"

// 错误结构体
var (
	ErrorRequestBodyParseFailed = gin.H{
		"error":      "Request body is not correct.",
		"error_code": "0001",
	} // 不能解析消息体

	ErrorDBError = gin.H{
		"error":      "The process failed, please try again later.",
		"error_code": "003",
	} // 数据库操作错误

	ErrorNotAuthSession = gin.H{
		"error":      "session is not right.",
		"error_code": "006",
	} // 账户不合法，不存在

	ErrorCreateAccountError = gin.H{
		"error":      "You can only create accounts for junior users,",
		"error_code": "0010",
	} // 创建用户等级不合法

	ErrorCreateAccountPriError = gin.H{
		"error":      "dispatcher can only create dispatcher accounts",
		"error_code": "0010",
	} // 创建用户等级不合法

	ErrorInternalServerError = gin.H{
		"error":      "The Internal Server failed, please try again later.",
	} // 请求返回统一错误，不能把错误传出去

)

var (
	SuccessResponse = gin.H{
		"msg":      "The process successfully!",
	} // 操作成功返回的信息

)
