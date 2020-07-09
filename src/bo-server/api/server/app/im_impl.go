/*
@Time : 2019/8/10 16:36 
@Author : yanKoo
@File : im_impl
@Software: GoLand
@Description:实现了im相关的三个rpc调用
*/
package app

import (
	pb "bo-server/api/proto"
	"bo-server/api/server/im"
	"context"
)

// DataPublish rpc 方法实现
func (tcs TalkCloudServiceImpl) DataPublish(srv pb.TalkCloud_DataPublishServer) error {
	return im.ServerWorker{}.DataPublish(srv)
}

// ImMessagePublish rpc 方法实现
func (tcs TalkCloudServiceImpl) ImMessagePublish(ctx context.Context, req *pb.ImMsgReqData) (*pb.ImMsgRespData, error) {
	return im.ServerWorker{}.ImMessagePublish(ctx, req)
}

// ImSosPublish rpc 方法实现
func (tcs TalkCloudServiceImpl) ImSosPublish(ctx context.Context, req *pb.ReportDataReq) (*pb.ImMsgRespData, error) {
	return im.ServerWorker{}.ImSosPublish(ctx, req)
}
