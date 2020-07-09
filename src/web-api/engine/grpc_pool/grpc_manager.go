/*
@Time : 2019/9/3 18:09 
@Author : yanKoo
@File : grpc_manager
@Software: GoLand
@Description:用来管理grpc连接
*/
package grpc_pool

import (
	"context"
	"google.golang.org/grpc"
	"math"
	"strconv"
	"strings"
	"sync"
	cfgWs "web-api/config"
	"web-api/logger"
	"web-api/model"
	"web-api/utils"
)

type grpcConnNode struct {
	ip       string
	ConnPool *Pool // grpc连接池
	idMap    *utils.BitMap
	m        sync.Mutex
}

type gRPCNodeManager struct {
	connNodes []*grpcConnNode
}

var GRPCManager gRPCNodeManager

func init() {
	GRPCManager = gRPCNodeManager{}
}

func (gm *gRPCNodeManager) GetGRPCConnClientById(id int, opts ...bool) *ClientConn {
	var check = false
	for i, opt := range opts {
		if i == 0 {
			check = opt
		}
	}
	return gm.getImGRPCConnWithCheck(id, check)
}

func (gm *gRPCNodeManager) ChangeDispatcherServerId(id int, serverAddr string) {
	gm.ChangeDispatcherServerId(id, serverAddr)
}
func (gm *gRPCNodeManager) changeDispatcherServerId(id int, serverAddr string) {
	addr := strings.Split(serverAddr, ":")
	if addr == nil || len(addr) != 2 {
		return
	}
	for _, grpcNode := range gm.connNodes {
		if grpcNode.idMap.IsExist(uint(id)) {
			grpcNode.idMap.Remove(uint(id))
			logger.Debugf("gm will remove %d from %s", id, grpcNode.ip)
		}
	}

	for i, node := range gm.connNodes {
		if node.ip == addr[0] {
			gm.connNodes[i].m.Lock()
			defer gm.connNodes[i].m.Unlock()
			gm.connNodes[i].idMap.Add(uint(id))
			logger.Debugf("gm will change %d to %s", id, gm.connNodes[i].ip)
			return
		}
	}

	// 有可能这个服务节点这里没有缓存过这时候就需要增加节点，然后注册
	gm.createServerNodeAndGetConn(id, &model.ServerGroup{
		Grpc: &model.GrpcServer{
			ServerNode: model.ServerNode{
				Addr: model.Addr{
					Ip: addr[0], Port: addr[1],
				},
			},
		},
	})
	return
}

// 调度员获取连接池客户端
func (gm *gRPCNodeManager) getImGRPCConnWithCheck(id int, check bool) *ClientConn {
	// 1. 通过bitmap从连接池中获取连接，获取不到再去web-gateway获取
	if cli, isEmpty := gm.getConnById(id, check); cli != nil || isEmpty {
		return cli
	}
	// 2. 如果找不到有两种情况，
	// 第一种是grpc的connect已经有了，但是这个调度员没有来访问过，所以bitmap里没有，第二种是这个grpc的connect都没有
	serverAddr := gm.queryServerAddr(id)
	if serverAddr.Server == nil {
		return nil
	}
	logger.Debugf("gm can't find # %d, so query result grpc ip:%s", id ,serverAddr.Server.Grpc.Addr.Ip)
	// 2. 检查这个query的ip服务有没有注册过服务节点
	if gm.checkRegisterServerNode(serverAddr.Server) {
		// 如果有这个bitmap就直接取， 但是调度员没有来访问过，没有访问过就注册
		return gm.registerAndGetConn(id, serverAddr.Server.Grpc.Addr.Ip)
	}

	// 3. 如果这个服务节点都没有访问过，那么先创建服务节点，在返回连接
	return gm.createServerNodeAndGetConn(id, serverAddr.Server)
}

func (gm *gRPCNodeManager) queryServerAddr(id int) *model.WebGateWayResp {
	return utils.WebGateWay{
		Url: cfgWs.DispatcherServerAddrUrl,
	}.GetDispatcherServerAddr(map[string]string{
		//utils.Account_NAME: "test002",
		utils.Account_ID: strconv.Itoa(id),
	})
}

// 检查这个服务节点有没有在这里注册过 true 是注册过
func (gm *gRPCNodeManager) checkRegisterServerNode(serverAddr *model.ServerGroup) bool {
	for _, connNode := range gm.connNodes {
		if connNode.ip == serverAddr.Grpc.Addr.Ip {
			return true
		}
	}
	return false
}

// 通过id获取grpc连接
func (gm *gRPCNodeManager) getConnById(id int, check bool) (*ClientConn, bool) {
	for _, connNode := range gm.connNodes {
		if connNode.idMap.IsExist(uint(id)) {
			logger.Debugf("gm find #%d with ip: %s", id, connNode.ip)
			if !check {
				return gm.getConnFromPool(connNode.ConnPool)
			}
			// 如果要检查服务地址
			serverAddr := gm.queryServerAddr(id)
			if serverAddr.Server.Grpc.Addr.Ip == connNode.ip {
				return gm.getConnFromPool(connNode.ConnPool)
			} else {
				gm.changeDispatcherServerId(id, serverAddr.Server.Grpc.Addr.Ip + ":" + serverAddr.Server.Grpc.Addr.Port)
				return gm.getConnById(id, false)
			}
		}
	}
	return nil, false
}

func (gm *gRPCNodeManager) getConnFromPool(p *Pool) (*ClientConn, bool) {
	if p.Available() <= 0 {
		logger.Debugln("gm pool is empty")
		return nil, true
	}
	poolClient, _ := p.Get(context.TODO())
	return poolClient, false
}

// 检查这个服务节点有没有在这里注册过 true 是注册过
func (gm *gRPCNodeManager) registerAndGetConn(id int, ip string) *ClientConn {
	for i := range gm.connNodes {
		if gm.connNodes[i].ip == ip {
			logger.Debugf("gm registered and return # %d , ip: %s", id, ip)
			gm.connNodes[i].idMap.Add(uint(id))
			cli, _ := gm.getConnFromPool(gm.connNodes[i].ConnPool)
			return cli
		}
	}
	return nil
}

func (gm *gRPCNodeManager) createServerNodeAndGetConn(id int, serverAddr *model.ServerGroup) *ClientConn {
	// 1. 先创建新连接
	logger.Debugf("gm will create ServerNode ip:%s and add %d", serverAddr.Grpc.Addr.Ip, id)
	pool, err := New(func() (*grpc.ClientConn, error) {
		return grpc.Dial(serverAddr.Grpc.Addr.Ip+":"+serverAddr.Grpc.Addr.Port, grpc.WithInsecure())
	}, 10, 30, 0)
	if err != nil {
		logger.Debugf("gm can't create grpc conn")
		return nil
	}
	// 2. 添加服务节点
	gm.connNodes = append(gm.connNodes, &grpcConnNode{
		ip:       serverAddr.Grpc.Addr.Ip,
		ConnPool: pool,
		idMap:    utils.NewBitMap(math.MaxInt32),
	})
	// 3. 把id注册到这个bitmap上
	return gm.registerAndGetConn(id, serverAddr.Grpc.Addr.Ip)
}
