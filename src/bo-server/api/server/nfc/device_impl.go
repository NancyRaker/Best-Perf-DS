/*
@Time : 2019/10/21 14:48 
@Author : yanKoo
@File : device_impl
@Software: GoLand
@Description:
*/
package nfc

import (
	pb "bo-server/api/proto"
	"bo-server/dao/nfc_device_clock"
	"bo-server/logger"
	"context"
	"errors"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"net/http"
)

// 设备打卡 TODO
func (wssu *NFCServiceServerImpl) DeviceClockReport(ctx context.Context, req *pb.ClockData) (*pb.ClockDataResp, error) {
	// 0. 校验数据
	logger.Debugf("DeviceClockReport req: %+v", req)
	if req == nil || req.DeviceId <= 0 || req.TagId <= 0 || req.ClockTime <= 0 {
		return nil, errors.New("param is invalid")
	}

	// 1. 存储到数据库 失败也算了
	if err := nfc_device_clock.SaveNFCDeviceClockRecord(req); err != nil {
		logger.Errorf("DeviceClockReport err: %+v", err)
	}

	// 2. TODO 存储到缓存？

	return &pb.ClockDataResp{Res:&pb.Result{Code:http.StatusOK}}, nil
}

// 查询设备打卡状态这个只能去mysql查询
func (wssu *NFCServiceServerImpl) DeviceClockStatus(ctx context.Context, req *pb.ClockStatusReq) (*pb.ClockStatusResp, error) {
	// nfc_device_clock.QueryNFCDeviceClockRecord()
	// 0. 校验数据

	// 1. 查询该设备打卡时间 TODO 目前默认只是查看 [req.StartTimestamp, req.StartTimestamp + 24h)时间段内的打卡记录
	clockRecords, err := nfc_device_clock.QueryNFCDeviceClockRecord(req)
	if err != nil {

	}
	// 2. 查询打卡任务，找出该设备在这一天内要执行哪几项打卡任务，任务打卡的标签点，要打卡的时间间隔

	// 3. 遍历设备在这一天内的打卡记录，判断是否正确打卡
	for _, cr := range clockRecords {
		checkRecordsIsVaild(cr)
	}
	// 4. 返回结果

	return nil, status.Errorf(codes.Unimplemented, "method DeviceClockStatus not implemented")
}

func checkRecordsIsVaild(data *pb.ClockData) {
	
	
}

