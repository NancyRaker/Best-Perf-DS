/*
@Time : 2019/11/5 17:00 
@Author : yanKoo
@File : web_gateway_impl
@Software: GoLand
@Description:
*/
package web_ips

import (
	pb "bo-server/api/proto"
	"bo-server/api/server/im"
	"bo-server/dao/user"
	"context"
	"net/http"
)

type WebIpsServiceServerImpl struct {
}

// 通知调度员和设备，ips已经更新
func (*WebIpsServiceServerImpl) NotifyIpsChanged(ctx context.Context, req *pb.NotifyIpsChangedReq) (*pb.NotifyIpsChangedResp, error) {
	// 查询当前调度员名下有多少设备，然后去发通知
	deviceAll, err := user.SelectUserByAccountId(int(req.AccountId))
	if err != nil {
		return nil, err
	}
	var receivers = []int32{req.AccountId}
	for _, device := range deviceAll {
		receivers = append(receivers, int32(device.Id))
	}

	// 发送通知
	im.SendSingleNotify(receivers, req.AccountId, &pb.StreamResponse{
		Uid:      req.AccountId,
		DataType: im.IPS_CHANGED_NOTIFY,})

	return &pb.NotifyIpsChangedResp{Res:&pb.Result{Code:http.StatusOK}}, nil
}
