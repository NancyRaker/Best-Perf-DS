/*
@Time : 2019/9/5 16:37 
@Author : yanKoo
@File : grpc_call
@Software: GoLand
@Description:
*/
package grpc_pool

import (
	"context"
	"errors"
	"google.golang.org/grpc"
	"web-ips/logger"
	pb "web-ips/proto/talk_cloud"
)

const (
	NotifyIpsChanged   = "NotifyIpsChanged"
)

type ConnPoolNode struct {
	connPool *Pool // grpc连接池
}

func NewGRpcPool(serverAddr string) *ConnPoolNode {
	return &ConnPoolNode{connPool: createServerNodeAndGetConn(serverAddr)}
}

func (gcn *ConnPoolNode) getConnFromPool(p *Pool) *ClientConn {
	if p.Available() <= 0 {
		logger.Debugln("gm pool is empty")
		return nil
	}
	poolClient, _ := p.Get(context.TODO())
	return poolClient
}

func createServerNodeAndGetConn(serverAddr string) *Pool {
	// 1. 先创建新连接
	logger.Debugf("gcn will create server addr :%s", serverAddr)
	pool, err := New(func() (*grpc.ClientConn, error) {
		return grpc.Dial(serverAddr, grpc.WithInsecure())
	}, 10, 10, 0)
	if err != nil {
		logger.Debugf("gm can't create grpc conn")
		return nil
	}
	return pool
}

func (gcn *ConnPoolNode) RpcCall(ctx context.Context, req interface{}, fName string, opts ...grpc.CallOption) (interface{}, error) {
	if cli := gcn.getConnFromPool(gcn.connPool); cli == nil {
		return nil, errors.New("can't call rpc func because pool client is nil")
	} else {
		conn := cli.ClientConn
		defer cli.Close()
		nfcCli := pb.NewWebIPsServiceClient(conn)
		switch fName {
		/******************************* IPs RPC Start **********************************/
		case NotifyIpsChanged:
			return nfcCli.NotifyIpsChanged(ctx, req.(*pb.NotifyIpsChangedReq))

		}
		/******************************* IPs RPC End **********************************/
	}
	return nil, errors.New("no is func")
}

//func GrpcWebRpcCall(id int, ctx context.Context, req interface{}, fName string, opts ...grpc.CallOption) (interface{}, error) {
//	if cli := GRPCManager.GetGRPCConnClientById(id); cli == nil {
//		return nil, errors.New("can't call rpc func because pool client is nil")
//	} else {
//		conn := cli.ClientConn
//		defer cli.Close()
//		webCli := pb.NewWebServiceClient(conn)
//		nfcCli := pb.NewNFCServiceClient(conn)
//		switch fName {
//		case WebCreateGroup:
//			return webCli.WebCreateGroup(ctx, req.(*pb.WebCreateGroupReq))
//		case UpdateGroup:
//			return webCli.UpdateGroup(ctx, req.(*pb.UpdateGroupReq))
//		case UpdateGroupInfo:
//			return webCli.UpdateGroupInfo(ctx, req.(*pb.GroupInfo))
//		case DeleteGroup:
//			return webCli.DeleteGroup(ctx, req.(*pb.GroupDelReq))
//		case ImportDeviceByRoot:
//			return webCli.ImportDeviceByRoot(ctx, req.(*pb.ImportDeviceReq))
//		case UpdateDeviceInfo:
//			return webCli.UpdateDeviceInfo(ctx, req.(*pb.UpdDInfoReq))
//		case SelectDeviceByImei:
//			return webCli.SelectDeviceByImei(ctx, req.(*pb.ImeiReq))
//		case GetDevicesInfo:
//			return webCli.GetDevicesInfo(ctx, req.(*pb.DeviceInfosReq))
//		case DeleteAccount:
//			return webCli.DeleteAccount(ctx, req.(*pb.DeleteAccountReq))
//		case UpdateGroupManager:
//			return webCli.UpdateGroupManager(ctx, req.(*pb.UpdGManagerReq))
//		case PostWifiInfo:
//			return webCli.PostWifiInfo(ctx, req.(*pb.WifiInfoReq))
//		case GetWifiInfo:
//			return webCli.GetWifiInfo(ctx, req.(*pb.GetWifiInfoReq))
//		case GetDeviceLogInfo:
//			return webCli.GetDeviceLogInfo(ctx, req.(*pb.GetDeviceLogInfoReq))
//		case GetGpsForTrace:
//			return webCli.GetGpsForTrace(ctx, req.(*pb.GpsForTraceReq))
//		case MultiTransDevice:
//			return webCli.MultiTransDevice(ctx, req.(*pb.TransDevices))
//		case ChangeDBEngine:
//			return webCli.ChangeDBEngine(ctx, req.(*pb.Empty))
//		case CreateAccount:
//			return webCli.CreateAccount(ctx, req.(*pb.CreateAccountReq))
//		/******************************* NFC RPC Start **********************************/
//		case PostTagsInfo:
//			return nfcCli.PostTagsInfo(ctx, req.(*pb.TagsInfoReq))
//		case GetTagsInfo:
//			return nfcCli.GetTagsInfo(ctx, req.(*pb.GetTagsInfoReq))
//		case PostTagTasksList:
//			return nfcCli.PostTagTasksList(ctx, req.(*pb.TagTasksListReq))
//		case QueryTasksByDevice:
//			return nfcCli.QueryTasksByDevice(ctx, req.(*pb.DeviceTasksReq))
//		case QueryTaskDetail: // 查询任务详细信息
//			return nfcCli.QueryTaskDetail(ctx, req.(*pb.TaskDetailReq))
//		case SetReportInfo: // 设置邮箱，周报月报
//			return nfcCli.SetReportInfo(ctx, req.(*pb.ReportInfoReq))
//		case QueryReportInfo: // 查询邮箱，周报月报
//			return nfcCli.QueryReportInfo(ctx, req.(*pb.ReportSetInfoReq))
//		case DeviceClockStatus: // 查询设备打卡情况
//			return nfcCli.DeviceClockStatus(ctx, req.(*pb.ClockStatusReq))
//
//		}
//		/******************************* NFC RPC End **********************************/
//	}
//
//	return nil, errors.New("no is func")
//}
//
//func GrpcAppRpcCall(id int, ctx context.Context, req interface{}, fName string, opts ...bool /*grpc.CallOption*/) (interface{}, error) {
//	var check = false
//	for i, opt := range opts {
//		if i == 0 {
//			check = opt
//		}
//	}
//	if cli := GRPCManager.GetGRPCConnClientById(id, check); cli == nil {
//		return nil, errors.New("can't call rpc func because pool client is nil")
//	} else {
//		conn := cli.ClientConn
//		defer cli.Close()
//		appCli := pb.NewTalkCloudClient(conn)
//		switch fName {
//		case RemoveGrp:
//			return appCli.RemoveGrp(ctx, req.(*pb.GroupDelReq))
//		case ImMessagePublish:
//			return appCli.ImMessagePublish(ctx, req.(*pb.ImMsgReqData))
//		}
//	}
//
//	return nil, errors.New("no is func")
//}
