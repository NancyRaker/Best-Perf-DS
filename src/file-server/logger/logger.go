/*
@Time : 2019/5/13 17:27
@Author : yanKoo
@File : logger
@Software: GoLand
@Description:
*/
package logger

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/lestrrat/go-file-rotatelogs"
	"github.com/rifflock/lfshook"
	"github.com/sirupsen/logrus"
	"os"
	"time"
	"file-server/conf"
)

var log *logrus.Logger

const Time_Layout = "2006-01-02 15:04:05.000000000"

func Debugf(format string, args ...interface{}) {
	log.Debugf(format, args...)
}

func Errorf(format string, args ...interface{}) {
	log.Errorf(format, args...)
}

func Fatalf(format string, args ...interface{}) {
	log.Fatalf(format, args...)
}

func Error(args ...interface{}) {
	log.Error(args...)
}

func Debugln(args ...interface{}) {
	log.Debugln(args...)
}
func init() {
	log = logrus.New()
	//禁止logrus的输出
	src, err := os.OpenFile(os.DevNull, os.O_APPEND|os.O_WRONLY, os.ModeAppend)
	if err != nil {
		fmt.Println("err", err)
	}
	log.Out = src
	// 设置日志级别
	switch conf.LogLevel{
	case "info":
		log.SetLevel(logrus.InfoLevel)
	case "debug":
		log.SetLevel(logrus.DebugLevel)
	default:
		log.SetLevel(logrus.DebugLevel)
	}

	log.SetFormatter(&logrus.TextFormatter{
		DisableColors: true,
		FullTimestamp: true,
	})

	apiLogPath := "./logger/web"
	logWriter, err := rotatelogs.New(
		apiLogPath+".%Y-%m-%d-%H-%M.log",
		rotatelogs.WithLinkName(apiLogPath),       // 生成软链，指向最新日志文件
		rotatelogs.WithMaxAge(2*24*time.Hour),     // 文件最大保存时间
		rotatelogs.WithRotationTime(24*time.Hour), // 日志切割时间间隔
	)
	writeMap := lfshook.WriterMap{
		logrus.InfoLevel:  logWriter,
		logrus.ErrorLevel: logWriter,
		logrus.DebugLevel: logWriter,
		logrus.FatalLevel: logWriter,
	}
	lfHook := lfshook.NewHook(writeMap, &logrus.TextFormatter{
		FullTimestamp:   true,
		TimestampFormat: Time_Layout,
	})
	log.AddHook(lfHook)
}

func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 开始时间
		start := time.Now()
		// 处理请求
		c.Next()
		// 结束时间
		end := time.Now()
		//执行时间
		latency := end.Sub(start)

		path := c.Request.URL.Path

		clientIP := c.ClientIP()
		method := c.Request.Method
		statusCode := c.Writer.Status()
		log.Infof("| %3d | %13v | %15s | %s  %s |",
			statusCode,
			latency,
			clientIP,
			method, path,
		)
	}
}
