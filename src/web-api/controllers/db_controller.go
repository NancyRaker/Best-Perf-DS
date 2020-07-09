/*
@Time : 2019/8/6 16:36 
@Author : yanKoo
@File : db
@Software: GoLand
@Description:
*/
package controllers

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
	"sync"
	"web-api/config"
	"web-api/engine/db"
)

var flag bool

func init() {
	flag = false
}

// 切换数据库连接 /db/engine/change TODO 对请求ip进行限制，只允许微软云和东莞进行通信
func ChangeDBEngine(c *gin.Context) {
	var (
		err error
		m   sync.Mutex
	)
	m.Lock()
	defer m.Unlock()

	// 默认收到这个请求就更换数据库引擎
	if !flag {
		fmt.Println("change")
		flag = true // 表示即将切换到另一台
		db.DBHandler, err = db.CreateDBHandler("db2",  config.DEFAULT_CONFIG)
		if err != nil {
			c.JSON(http.StatusOK, gin.H{"res": "fail" + err.Error()})
		}

		// 通知im也更换数据库连接

	} else {
		flag = false // 切换回本机
		db.DBHandler, err = db.CreateDBHandler("db", config.DEFAULT_CONFIG)
		if err != nil {
			c.JSON(http.StatusOK, gin.H{"res": "fail2" + err.Error()})
		}
		// 通知im也更换数据库连接

	}
	c.JSON(http.StatusOK, gin.H{"res": "success"})
}
