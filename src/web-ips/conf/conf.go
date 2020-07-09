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
	"web-ips/model"
)

var (
	WebHttpPort    string // web api的监听端口
	WebHttpsPort    string // web api的监听端口 https
	GrpcAddr   string // grpc服务端的地址（包含端口）， 主要用于web模块调用grpc服务
	CertFile   string // https pem文件名
	KeyFile    string // https key文件
	HttpWay    string
	HttpsWay    string
	PprofAddr  string // pprof监听的地址
	TimeLayout string // 时间模板
	LogLevel   string // 日志级别

	ServerNodesBean = &model.ServerInitNodes{}
)

func init() {
	cfgFilePath := flag.String("d", "conf.ini", "web gateway configuration file path")
	flag.Parse()
	cfg, err := ini.Load(*cfgFilePath) // 编译之后的执行文件所在位置的相对位置
	if err != nil {
		fmt.Printf("Fail to read file: %v", err)
		os.Exit(1)
	}
	WebHttpPort = cfg.Section("web_gateway").Key("port1").String()
	WebHttpsPort = cfg.Section("web_gateway").Key("port2").String()

	GrpcAddr = cfg.Section("grpc").Key("addr").String()

	// https
	CertFile = cfg.Section("https").Key("cert_file").String()
	KeyFile = cfg.Section("https").Key("key_file").String()

	HttpWay = cfg.Section("http_way").Key("way1").String()
	HttpsWay = cfg.Section("http_way").Key("way2").String()

	TimeLayout = cfg.Section("time").Key("layout").String()

	LogLevel = cfg.Section("log").Key("level").String()

	PprofAddr = cfg.Section("pprof").Key("pprof_addr").String()

	ServerNodesBean.IPs = strings.Split(cfg.Section("server_nodes").Key("ips").String(), "|")
	ServerNodesBean.ServerIps = strings.Split(cfg.Section("server_nodes").Key("server_ips").String(), "|")
	ServerNodesBean.ServerCodes = strings.Split(cfg.Section("server_nodes").Key("sever_codes").String(), "|")
	ServerNodesBean.ServerNames = strings.Split(cfg.Section("server_nodes").Key("names").String(), "|")
	ServerNodesBean.Ports = strings.Split(cfg.Section("server_nodes").Key("ports").String(), "|")
	ServerNodesBean.JanusDomains = strings.Split(cfg.Section("server_nodes").Key("janus_domains").String(), "|")
}
