/*
@Time : 2019/9/21 14:04 
@Author : yanKoo
@File : core
@Software: GoLand
@Description:
*/
package server

import (
	"bo-server/api/server/app"
	"bo-server/api/server/location"
	"bo-server/api/server/nfc"
	"bo-server/api/server/web"
	"bo-server/api/server/web_ips"
	"bo-server/logger"
	"flag"
	"github.com/ceshihao/ratelimiter/tokenbucket"
	"github.com/grpc-ecosystem/go-grpc-middleware"
	"github.com/grpc-ecosystem/go-grpc-middleware/ratelimit"
	"net"
	"net/http"
	"time"

	pb "bo-server/api/proto"
	cfgGs "bo-server/conf"
	//"github.com/ceshihao/ratelimiter/tokenbucket"
	//"github.com/grpc-ecosystem/go-grpc-middleware"
	//"github.com/grpc-ecosystem/go-grpc-middleware/ratelimit"
	"google.golang.org/grpc"
	_ "net/http/pprof"
)

type boServerNode struct {
	boServer *grpc.Server
}

func NewBoServer(opt ...grpc.ServerOption) *boServerNode {
	var options []grpc.ServerOption
	for _, o := range opt {
		options = append(options, o)
	}
	return &boServerNode{boServer: grpc.NewServer(options...)}
}

// 初始化一些辅助的服务
func (bsn *boServerNode) InitAssistServer() *boServerNode {
	flag.Parse()

	go func() {
		_ = http.ListenAndServe(cfgGs.PprofAddr, nil)
	}()
	return bsn
}

// 注册rpc方法
func (bsn *boServerNode) RegisterServer() *boServerNode {
	// 注册app相关服务
	pb.RegisterTalkCloudServer(bsn.boServer, &app.TalkCloudServiceImpl{})

	// 注册apk升级服务
	pb.RegisterApkUpdateServiceServer(bsn.boServer, &app.ApkUpdateServiceImpl{})

	// 注册设备位置相关服务
	pb.RegisterTalkCloudLocationServer(bsn.boServer, &location.TalkCloudLocationServiceImpl{})

	// 注册web操作服务
	pb.RegisterWebServiceServer(bsn.boServer, &web.WebServiceServerImpl{})

	// 注册NFC服务
	pb.RegisterNFCServiceServer(bsn.boServer, &nfc.NFCServiceServerImpl{})

	// 注册web-gateway服务
	pb.RegisterWebIPsServiceServer(bsn.boServer, &web_ips.WebIpsServiceServerImpl{})

	return bsn
}

// 运行服务
func (bsn *boServerNode) Run() {
	var (
		lis net.Listener
		err error
	)
	if lis, err = net.Listen("tcp", ":"+cfgGs.GrpcSPort); err != nil {
		logger.Errorf("net listen err: %v", err)
	}

	logger.Infof("listing port: %s", cfgGs.GrpcSPort)
	if err = bsn.boServer.Serve(lis); err != nil {
		logger.Errorf("listen %s fail", cfgGs.GrpcSPort)
	} else {
		logger.Error("listing")
	}
}

// 请求限制
func limiter() {
	// 限制请求数
	unaryRateLimiter := tokenbucket.NewTokenBucketRateLimiter(1*time.Second, 10, 10, 10*time.Second)
	streamRateLimiter := tokenbucket.NewTokenBucketRateLimiter(1*time.Second, 5, 5, 5*time.Second)
	talkCloudServer := grpc.NewServer(
		grpc_middleware.WithUnaryServerChain(
			ratelimit.UnaryServerInterceptor(unaryRateLimiter),
		),
		grpc_middleware.WithStreamServerChain(
			ratelimit.StreamServerInterceptor(streamRateLimiter),
		),
	)
	defer talkCloudServer.GracefulStop()
}
