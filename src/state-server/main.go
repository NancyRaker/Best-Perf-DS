/*
@Time : 2019/9/18 13:41
@Author : yanKoo
@File : main
@Software: GoLand
@Description:
*/
package main

import (
	"net/http"
	_ "net/http/pprof"
	"state-server/conf"
	"state-server/core/engine"
	"state-server/core/scheduler"
	"state-server/logger"
)

func main() {
	prepare()

	engine.NewEngine(
		scheduler.NewScheduler(),
		conf.Config.Common.WorkerCount,
	).Run()
}

func prepare() {
	//  创建pprof辅助监听进程运行状态
	go func() {
		logger.Debugf("PProf %s listening...", conf.Config.Common.PPRofAddr)
		_ = http.ListenAndServe(conf.Config.Common.PPRofAddr, nil)
	}()
}
