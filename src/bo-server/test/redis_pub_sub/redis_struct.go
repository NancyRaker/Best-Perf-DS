/*
@Time : 2019/10/17 11:41 
@Author : yanKoo
@File : redis_struct
@Software: GoLand
@Description:
*/
package main

import (
	"fmt"
	"time"
)
const (
	// 定义每分钟的秒数
	SecondsPerMinute = 60
	// 定义每小时的秒数
	SecondsPerHour = SecondsPerMinute * 60
	// 定义每天的秒数
	SecondsPerDay = SecondsPerHour * 24
)
// 将传入的“秒”解析为3种时间单位
func resolveTime(seconds int) (day int, hour int, minute int, second int) {
	day = seconds / SecondsPerDay
	hour = seconds / SecondsPerHour
	minute = seconds / SecondsPerMinute
	second = seconds % SecondsPerDay
	return
}


func main() {
	//获取本地location
	toBeCharge := "2015-01-01 00:00:00"                             //待转化为时间戳的字符串 注意 这里的小时和分钟还要秒必须写 因为是跟着模板走的 修改模板的话也可以不写
	timeLayout := "2006-01-02 15:04:05"                             //转化所需模板
	loc, _ := time.LoadLocation("UTC8")                            //重要：获取时区
	theTime, _ := time.ParseInLocation(timeLayout, toBeCharge, loc) //使用模板在对应时区转化为time.time类型
	sr := theTime.Unix()                                            //转化为时间戳 类型是int64
	fmt.Println(theTime)                                            //打印输出theTime 2015-01-01 15:15:00 +0800 CST
	fmt.Println(sr)                                                 //打印输出时间戳 1420041600

	//时间戳转日期
	dataTimeStr := time.Unix(1420106400, 0).In(loc).Format(timeLayout) //设置时间戳 使用模板格式化为日期字符串
	fmt.Println(dataTimeStr)
	/*
	2015-01-01 00:00:00 +0800 CST
	1420041600
	2015-01-01 00:00:00

	America/Adak
	2015-01-01 00:00:00 -1000 HST
	1420106400
	2015-01-01 18:00:00


	*/
}


