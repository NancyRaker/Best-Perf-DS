/*
@Time : 2019/4/20 17:11
@Author : yanKoo
@File : conf
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

var (
	WorkerCount int //  负责下载文件和上传文件的worker数量
	FILE_BASE_URL     string // 保存文件到fastdfs服务器之后的访问前缀（ip、域名）
	TrackerServerAddr string // fastdfs的tracker服务器的地址（包含ip）
	MaxConn           int    // fastdfs上传文件的最大连接数

	PprofAddr string // pprof监听的地址

	TimeLayout string // 时间模板
	LogLevel string // 日志级别

	PttWaitTime int

	PttMsgKey string

	ConsumerAddrs     []string // kafka消费者消费的地址
	ConsumerGroup     string   // kafka消费者组
	ConsumerTopic     string   // kafka消费者消费的topic
	ConsumerPartition int32    // kafka消费的分区
	ProducerAddrs     []string // kafka生产者投递消息的地址
	ProducerTopic     string   // kafka生产者投递消息的topic
)

const (
	DEFAULT_CONFIG = "file_storage_conf.ini"
)

func init() {
	cfgFilePath := flag.String("d", DEFAULT_CONFIG, "web api configuration file path")
	flag.Parse()
	cfg, err := ini.Load(*cfgFilePath) // 编译之后的执行文件所在位置的相对位置
	if err != nil {
		fmt.Printf("Fail to read file: %v", err)
		os.Exit(1)
	}

	WorkerCount, _ =  cfg.Section("file_server").Key("worker_count").Int()

	FILE_BASE_URL = cfg.Section("upload_file").Key("save_path_url").String()

	//fastdfs
	TrackerServerAddr = cfg.Section("fastdfs").Key("tracker_server").String()
	MaxConn, _ = cfg.Section("fastdfs").Key("maxConns").Int()

	TimeLayout = cfg.Section("time").Key("layout").String()

	LogLevel = cfg.Section("log").Key("level").String()

	PprofAddr = cfg.Section("pprof").Key("pprof_addr").String()

	PttMsgKey = cfg.Section("ptt").Key("ptt_msg_key").String()

	PttWaitTime, _ = cfg.Section("ptt").Key("wait_time").Int()

	consumerAddrS := cfg.Section("kafka").Key("consumer_addr").String()
	ConsumerAddrs = strings.Split(consumerAddrS, " ")

	ConsumerGroup = cfg.Section("kafka").Key("consumer_group").String()
	ConsumerTopic = cfg.Section("kafka").Key("consumer_topic").String()

	consumerPartition, _ := cfg.Section("kafka").Key("consumer_partition").Int()
	ConsumerPartition = int32(consumerPartition)

	producerAddrS := cfg.Section("kafka").Key("producer_addr").String()
	ProducerAddrs = strings.Split(producerAddrS, " ")

	ProducerTopic = cfg.Section("kafka").Key("producer_topic").String()
}
