/*
@Time : 2019/8/28 11:56 
@Author : yanKoo
@File : server_addr_service
@Software: GoLand
@Description:
*/
package service

import (
	"fmt"
	"math"
	"strings"
	"web-ips/bean"
	"web-ips/conf"
	"web-ips/grpc_pool"
	"web-ips/model"
	"web-ips/utils"
)

type serverNode struct {
	Ip           string                 `json:"ip"`
	Code         string                 `json:"code"`
	ServerInfo   *model.ServerResp      `json:"server_info"`  // 服务的id
	BitMapNode   *utils.BitMap          `json:"bit_map_node"` // 存储了这个服务器节点有哪些调度员id
	ConnPoolNode *grpc_pool.ConnPoolNode `json:"conn_pool_node"`
}

var ServerNodeList []*serverNode

func init() {
	initServerNodeListByConf()
}

func initServerNodeListByConf() {
	for i := 0; i < len(conf.ServerNodesBean.IPs); i++ {
		node := &serverNode{
			Ip:         conf.ServerNodesBean.IPs[i],
			Code:       conf.ServerNodesBean.ServerCodes[i],
			BitMapNode: utils.NewBitMap(math.MaxInt32),
		}
		serverIps := strings.Split(conf.ServerNodesBean.ServerIps[i], " ")
		names := strings.Split(conf.ServerNodesBean.ServerNames[i], " ")
		ports := strings.Split(conf.ServerNodesBean.Ports[i], " ")

		fmt.Println(serverIps)
		fmt.Println(names)
		fmt.Println(ports)
		node.ServerInfo = &model.ServerResp{
			Grpc: &model.GrpcServer{
				ServerInfo: model.ServerInfo{
					Addr: model.Addr{
						Ip:   serverIps[0],
						Port: ports[0],
					},
					Name: names[0],
				},
			},
			WebApi: &model.WebApiServer{
				ServerInfo: model.ServerInfo{
					Addr: model.Addr{
						Ip:   serverIps[1],
						Port: ports[1],
					},
					Name: names[1],
				},
			},
			Janus: &model.JanusServer{
				ServerInfo: model.ServerInfo{
					Addr: model.Addr{
						Ip:     serverIps[2],
						Port:   ports[2],
						Domain: conf.ServerNodesBean.JanusDomains[i],
					},
					Name: names[2],
				},
			},
		}

		// 注册连接池
		node.ConnPoolNode = grpc_pool.NewGRpcPool(node.ServerInfo.Grpc.Addr.Ip  + ":" + node.ServerInfo.Grpc.Addr.Port)

		ServerNodeList = append(ServerNodeList, node)
	}

	// TODO 默认code是从小到大排序，就不排序了
}

func NotifyIpsIsChanged()  {
	
}


func ChooseServerIp(imei string) *model.ServerResp {
	return queryDispatcherServer(DeviceTrie.Search(imei))
}

func chooseServerIpForDispatcherByName(name string) *model.ServerResp {
	return queryDispatcherServer(bean.AccountMap.Get(name))
}

func chooseServerIpForDispatcherById(accountId int) *model.ServerResp {
	if accountId <= 0 {
		return nil
	}
	return queryDispatcherServer(accountId)
}

func ChooseServerIpForDispatcher(key interface{}) *model.ServerResp {
	switch key.(type) {
	case int: // 根据id查询
		return chooseServerIpForDispatcherById(key.(int))
	case string: // 根据账号查询
		return chooseServerIpForDispatcherByName(key.(string))
	default:
		return nil
	}
}

/*func ChooseServerIpForDispatcher(name string) *model.ServerResp {
	return queryDispatcherServer(bean.AccountMap.Get(name))
}*/
func queryDispatcherServer(id int) *model.ServerResp {
	if id == -1 {
		return nil
	}

	for i := 0; i < len(ServerNodeList); i++ {
		if ServerNodeList[i].BitMapNode.IsExist(uint(id)) {
			return ServerNodeList[i].ServerInfo
		}
	}

	return nil
}
