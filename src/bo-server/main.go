package main

import (
	"bo-server/api/server"
	cfgGs "bo-server/conf"
	"bo-server/init_load"
	"runtime"

	_ "net/http/pprof"
)

const needLoadData = "y"

func init() {
	// 1. 加载数据库中所有的数据到缓存
	if cfgGs.NeedLoadData == needLoadData {
		init_load.ConcurrentEngine{
			Scheduler:   &init_load.SimpleScheduler{},
			WorkerCount: cfgGs.RedisCoMax, // 加载redis数据的协程数
		}.Run()
	}

	// 2. 设置内核cpu数目
	runtime.GOMAXPROCS(runtime.NumCPU())
}

func main() {
	server.NewBoServer().InitAssistServer().RegisterServer().Run()
}