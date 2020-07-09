/*
@Time : 2019/5/13 17:27
@Author : yanKoo
@File : logger
@Software: GoLand
@Description:
*/
package logger

import (
	"bo-server/conf"
	"github.com/lestrrat/go-file-rotatelogs"
	"github.com/rifflock/lfshook"
	"github.com/sirupsen/logrus"
	"os"
	"time"
)

var log *logrus.Logger

const Time_Layout = "2006-01-02 15:04:05.000000000"

func Debugf(format string, args ...interface{}) {
	log.Debugf(format, args...)
}

func Infof(format string, args ...interface{}) {
	log.Infof(format, args...)
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

	// 禁止logrus的输出
	src, err := os.OpenFile(os.DevNull, os.O_APPEND|os.O_WRONLY, os.ModeAppend)
	if err != nil {

	}
	log.Out = src

	// 设置日志级别
	switch conf.LogLevel {
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

	var apiLogPath = conf.LogPath + "/glog"
	logWriter, err := rotatelogs.New(
		apiLogPath+".%Y-%m-%d-%H-%M.log",
		rotatelogs.WithLinkName(apiLogPath),       // 生成软链，指向最新日志文件
		rotatelogs.WithMaxAge(7*24*time.Hour),     // 文件最大保存时间
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

	// Create elastic client
	/*client, err := elastic.NewClient(elastic.SetSniff(false), elastic.SetURL("http://" + config.ES_IP + ":" + config.ES_Port + "/"))
	if err != nil {
		log.WithError(err).Debug("Failed to construct elasticsearch client")
	}

	// Create logger with 5 seconds flush interval
	esHook, err := NewElasticHook(client, config.ES_IP, logrus.DebugLevel, func() string {
		return fmt.Sprintf("%s-%s", "grpc-server", time.Now().Format("2006-01-02"))
	}, time.Duration(config.ES_Interval))

	if err != nil {
		log.WithError(err).Debug("Failed to create elasticsearch hook for logger")
	}

	log.Hooks.Add(esHook)*/

}
