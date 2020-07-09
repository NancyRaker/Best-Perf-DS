/*
@Time : 2019/9/18 13:41
@Author : yanKoo
@File : main
@Software: GoLand
@Description:
*/
package main

import (
	"file-server/conf"
	"file-server/core/engine"
	"file-server/core/scheduler"
	"file-server/logger"
	"net/http"
	_ "net/http/pprof"
)

func main() {
	prepare()

	engine.NewEngine(
		scheduler.NewScheduler(),
		conf.WorkerCount,
		conf.WorkerCount,
	).Run()
}

func prepare() {
	//  创建pprof辅助监听进程运行状态
	go func() {
		logger.Debugf("PProf %s listening...", conf.PprofAddr)
		_ = http.ListenAndServe(conf.PprofAddr, nil)
	}()
}
