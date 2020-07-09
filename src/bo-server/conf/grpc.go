/*
@Time : 2019/4/29 10:35
@Author : yanKoo
@File : grpc
@Software: GoLand
@Description:
*/
package conf

import (
	"flag"
	"fmt"
	"github.com/go-ini/ini"
	"os"
	"strings"
)

const (
	DEFAULT_CONFIG = "bo_conf.ini"
)

var (
	GrpcSPort        string // grpc服务监听端口
	RedisCoMax       int    // 启动grpc服务的时候，读取redis工作的最大连接数
	NeedLoadData     string // 启动grpc服务的时候，是否需要从数据库更新到缓存
	Server_Work_Mode string // 启动grpc服务的时候，以什么模式工作， 目前主要用来区分日志路径

	FILE_BASE_URL string // 保存文件到fastdfs服务器之后的访问前缀（ip、域名）

	Interval    int    // im 心跳检测时间间隔
	PttMsgKey   string // ptt音视频数据在redis中的key
	PttWaitTime int    // ptt音视频获取时阻塞等待的时间

	PprofAddr  string // pprof监听的地址
	ExpireTime int

	TimeLayout     string // 时间模板
	LogLevel       string // 日志级别
	LogPath        string // 日志地址
	PTT_BASE_URL   string // ptt文件地址
	AppVersionCode int    // app最新版本号，小于这个就需要更新
	AppUrl         string // app更新地址

	AddAccountUrl   string // 创建设备，需要通知gateway更新
	AddDeviceUrl    string // 导入设备，需要通知gateway更新
	UpdateDeviceUrl string // 转移设备，需要通知web gateway更新
	WebGatewayAddrs string // 转移设备，需要通知web gateway更新

	ConsumerAddrs     []string // kafka消费者消费的地址
	ConsumerGroup     string   // kafka消费者组
	ConsumerTopic     string   // kafka消费者消费的topic
	ConsumerPartition int32    // kafka消费的分区
	ProducerAddrs     []string // kafka生产者投递消息的地址
	ProducerTopic     string   // kafka生产者投递消息的topic
)

func init() {
	iniFilePath := flag.String("d", DEFAULT_CONFIG, "grpc server conf file path")
	flag.Parse()
	cfg, err := ini.Load(*iniFilePath) // 编译之后的执行文件所在位置的相对位置
	if err != nil {
		fmt.Printf("Fail to read file: %v", err)
		os.Exit(1)
	}

	GrpcSPort = cfg.Section("grpc_server").Key("port").String()
	RedisCoMax, _ = cfg.Section("grpc_server").Key("redisCoMax").Int()
	Server_Work_Mode = cfg.Section("grpc_server").Key("work_mode").String()
	NeedLoadData = cfg.Section("grpc_server").Key("need_init_data").String()

	WebGatewayAddrs = cfg.Section("web-gateway").Key("addrs").String()
	AddAccountUrl = cfg.Section("web-gateway").Key("add_account_url").String()
	AddDeviceUrl = cfg.Section("web-gateway").Key("add_device_url").String()
	UpdateDeviceUrl = cfg.Section("web-gateway").Key("update_device_url").String()

	Interval, _ = cfg.Section("im").Key("heartbeat_interval").Int()

	PprofAddr = cfg.Section("pprof").Key("pprof_addr").String()

	FILE_BASE_URL = cfg.Section("upload_file").Key("save_path_url").String()

	PttMsgKey = cfg.Section("ptt").Key("ptt_msg_key").String()

	PttWaitTime, _ = cfg.Section("ptt").Key("wait_time").Int()

	PTT_BASE_URL = cfg.Section("ptt").Key("file_base_url").String()

	ExpireTime, _ = cfg.Section("im").Key("expire_time").Int()

	TimeLayout = cfg.Section("time").Key("layout").String()

	LogLevel = cfg.Section("log").Key("level").String()
	LogPath = cfg.Section("log").Key("path").String()

	AppVersionCode, _ = cfg.Section("app").Key("version_code").Int()

	AppUrl = cfg.Section("app").Key("apk_path").String()

	consumerAddrS := cfg.Section("kafka").Key("consumer_addr").String()
	ConsumerAddrs = strings.Split(consumerAddrS, " ")

	ConsumerGroup = cfg.Section("kafka").Key("consumer_group").String()
	ConsumerTopic = cfg.Section("kafka").Key("consumer_topic").String()

	consumerPartition, _ := cfg.Section("kafka").Key("consumer_partition").Int()
	ConsumerPartition = int32(consumerPartition)

	producerAddrS := cfg.Section("kafka").Key("producer_addr").String()
	ProducerAddrs = strings.Split(producerAddrS, " ")

	ProducerTopic = cfg.Section("kafka").Key("producer_topic").String()

	fmt.Println("ConsumerAddrs    :", ConsumerAddrs)
	fmt.Println("ConsumerGroup    :", ConsumerGroup)
	fmt.Println("ConsumerTopic    :", ConsumerTopic)
	fmt.Println("ConsumerPartition:", ConsumerPartition)
	fmt.Println("ProducerAddrs    :", ProducerAddrs)
	fmt.Println("ProducerTopic    :", ProducerTopic)
}
