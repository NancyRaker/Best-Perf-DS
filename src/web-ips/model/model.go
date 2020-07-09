/**
* @Author: yanKoo
* @Date: 2019/3/16 15:32
* @Description:
 */
package model

const (
	Micro        = "micro"
	GongGuan     = "dg"
	MicroCode    = 1
	GongGuanCode = 2
)

// response
type Addr struct {
	Domain string `json:"domain"`
	Ip     string `json:"ip"`
	Port   string `json:"port"`
}

type ServerInfo struct {
	Name string `json:"name"`
	Addr Addr   `json:"addr"`
}

type JanusServer struct {
	ServerInfo
}

type WebApiServer struct {
	ServerInfo
}

type GrpcServer struct {
	ServerInfo
}
type ServerResp struct {
	Janus  *JanusServer  `json:"janus"`
	WebApi *WebApiServer `json:"web_api"`
	Grpc   *GrpcServer   `json:"grpc"`
}

// 用来取mysql中的调度员信息
type Customer struct {
	AccountName string `json:"account_name"`
	Id          int    `json:"id"`
	AddrCode    int    `json:"addr_code"`
}

// 用来接收改变设备的调度员
type DeviceChange struct {
	IMeis     []string `json:"imeis"`
	AccountId int      `json:"account_id"`
}

// 用来接收改变设备的调度员
type DeviceAdd struct {
	IMeis     []string `json:"imeis"`
	AccountId int      `json:"account_id"`
}

// 用来接收改变设备的调度员
type DispatcherAdd struct {
	AccountName string `json:"account_name"`
	AccountId   uint   `json:"account_id"`
	CreatorId   uint   `json:"creator_id"`
}

/**
[server_nodes]
ips = "23.98.41.159 121.14.149.182 113.108.62.203"
server_ips = "23.98.41.159;23.98.41.159;23.98.41.159 121.14.149.182;121.14.149.182;121.14.149.182 113.108.62.203;113.108.62.203;113.108.62.203"
sever_code = "1 2 3"
names = "im;web;janus im;web;janus im;web;janus"
ports = "9001;10000;9188 9001;10000;9188 9001;10000;9188"
janus_domains = "hk213.yunptt.com:8989 dg182.yunptt.com:8989 dev.yunptt.com:8989"
*/
type ServerInitNodes struct {
	IPs          []string  `json:"i_ps"`
	ServerIps    []string  `json:"server_ips"`
	ServerCodes  []string  `json:"server_codes"`
	ServerNames  [] string `json:"server_names"`
	Ports        []string  `json:"ports"`
	JanusDomains []string  `json:"janus_domains"`
}
