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
	pb "web-api/proto/talk_cloud"
)

const (
	WebCreateGroup     = "WebCreateGroup"
	UpdateGroup        = "UpdateGroup"
	UpdateGroupInfo    = "UpdateGroupInfo"
	DeleteGroup        = "DeleteGroup"
	ImportDeviceByRoot = "ImportDeviceByRoot"
	UpdateDeviceInfo   = "UpdateDeviceInfo"
	SelectDeviceByImei = "SelectDeviceByImei"
	GetDevicesInfo     = "GetDevicesInfo"
	DeleteAccount      = "DeleteAccount"
	UpdateGroupManager = "UpdateGroupManager"
	PostWifiInfo       = "PostWifiInfo"
	GetWifiInfo        = "GetWifiInfo"
	GetDeviceLogInfo   = "GetDeviceLogInfo"
	GetGpsForTrace     = "GetGpsForTrace"
	MultiTransDevice   = "MultiTransDevice"
	ChangeDBEngine     = "ChangeDBEngine"
	CreateAccount      = "CreateAccount"

	/*nfc*/
	PostTagsInfo = "PostTagsInfo"
	GetTagsInfo  = "GetTagsInfo"

	PostTagTasksList   = "PostTagTasksList"
	QueryTasksByDevice = "QueryTasksByDevice"
	QueryTaskDetail    = "QueryTaskDetail"
	SetReportInfo      = "SetReportInfo"
	QueryReportInfo    = "QueryReportInfo"
	DeviceClockStatus  = "DeviceClockStatus"

	// app rpc func:
	RemoveGrp        = "RemoveGrp"
	ImMessagePublish = "ImMessagePublish"
	Login            = "Login"
	DataPublish      = "DataPublish"
)

func GrpcWebRpcCall(id int, ctx context.Context, req interface{}, fName string, opts ...grpc.CallOption) (interface{}, error) {
	if cli := GRPCManager.GetGRPCConnClientById(id); cli == nil {
		return nil, errors.New("can't call rpc func because pool client is nil")
	} else {
		conn := cli.ClientConn
		defer cli.Close()
		webCli := pb.NewWebServiceClient(conn)
		nfcCli := pb.NewNFCServiceClient(conn)
		switch fName {
		case WebCreateGroup:
			return webCli.WebCreateGroup(ctx, req.(*pb.WebCreateGroupReq))
		case UpdateGroup:
			return webCli.UpdateGroup(ctx, req.(*pb.UpdateGroupReq))
		case UpdateGroupInfo:
			return webCli.UpdateGroupInfo(ctx, req.(*pb.GroupInfo))
		case DeleteGroup:
			return webCli.DeleteGroup(ctx, req.(*pb.GroupDelReq))
		case ImportDeviceByRoot:
			return webCli.ImportDeviceByRoot(ctx, req.(*pb.ImportDeviceReq))
		case UpdateDeviceInfo:
			return webCli.UpdateDeviceInfo(ctx, req.(*pb.UpdDInfoReq))
		case SelectDeviceByImei:
			return webCli.SelectDeviceByImei(ctx, req.(*pb.ImeiReq))
		case GetDevicesInfo:
			return webCli.GetDevicesInfo(ctx, req.(*pb.DeviceInfosReq))
		case DeleteAccount:
			return webCli.DeleteAccount(ctx, req.(*pb.DeleteAccountReq))
		case UpdateGroupManager:
			return webCli.UpdateGroupManager(ctx, req.(*pb.UpdGManagerReq))
		case PostWifiInfo:
			return webCli.PostWifiInfo(ctx, req.(*pb.WifiInfoReq))
		case GetWifiInfo:
			return webCli.GetWifiInfo(ctx, req.(*pb.GetWifiInfoReq))
		case GetDeviceLogInfo:
			return webCli.GetDeviceLogInfo(ctx, req.(*pb.GetDeviceLogInfoReq))
		case GetGpsForTrace:
			return webCli.GetGpsForTrace(ctx, req.(*pb.GpsForTraceReq))
		case MultiTransDevice:
			return webCli.MultiTransDevice(ctx, req.(*pb.TransDevices))
		case ChangeDBEngine:
			return webCli.ChangeDBEngine(ctx, req.(*pb.Empty))
		case CreateAccount:
			return webCli.CreateAccount(ctx, req.(*pb.CreateAccountReq))
		/******************************* NFC RPC Start **********************************/
		case PostTagsInfo:
			return nfcCli.PostTagsInfo(ctx, req.(*pb.TagsInfoReq))
		case GetTagsInfo:
			return nfcCli.GetTagsInfo(ctx, req.(*pb.GetTagsInfoReq))
		case PostTagTasksList:
			return nfcCli.PostTagTasksList(ctx, req.(*pb.TagTasksListReq))
		case QueryTasksByDevice:
			return nfcCli.QueryTasksByDevice(ctx, req.(*pb.DeviceTasksReq))
		case QueryTaskDetail: // 查询任务详细信息
			return nfcCli.QueryTaskDetail(ctx, req.(*pb.TaskDetailReq))
		case SetReportInfo: // 设置邮箱，周报月报
			return nfcCli.SetReportInfo(ctx, req.(*pb.ReportInfoReq))
		case QueryReportInfo: // 查询邮箱，周报月报
			return nfcCli.QueryReportInfo(ctx, req.(*pb.ReportSetInfoReq))
		case DeviceClockStatus: // 查询设备打卡情况
			return nfcCli.DeviceClockStatus(ctx, req.(*pb.ClockStatusReq))
			
		}
		/******************************* NFC RPC End **********************************/
	}

	return nil, errors.New("no is func")
}

func GrpcAppRpcCall(id int, ctx context.Context, req interface{}, fName string, opts ...bool /*grpc.CallOption*/) (interface{}, error) {
	var check = false
	for i, opt := range opts {
		if i == 0 {
			check = opt
		}
	}
	if cli := GRPCManager.GetGRPCConnClientById(id, check); cli == nil {
		return nil, errors.New("can't call rpc func because pool client is nil")
	} else {
		conn := cli.ClientConn
		defer cli.Close()
		appCli := pb.NewTalkCloudClient(conn)
		switch fName {
		case RemoveGrp:
			return appCli.RemoveGrp(ctx, req.(*pb.GroupDelReq))
		case ImMessagePublish:
			return appCli.ImMessagePublish(ctx, req.(*pb.ImMsgReqData))
		}
	}

	return nil, errors.New("no is func")
}
