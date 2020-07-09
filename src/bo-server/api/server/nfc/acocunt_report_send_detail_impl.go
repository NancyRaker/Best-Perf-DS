/*
@Time : 2019/10/22 14:13 
@Author : yanKoo
@File : acocunt_report_send_detail_impl
@Software: GoLand
@Description:
*/
package nfc

import (
	pb "bo-server/api/proto"
	"bo-server/dao/nfc_report"
	"bo-server/logger"
	"context"
	"errors"
	"net/http"
)

// web 保存，修改，删除 邮箱周报月报的信息
func (wssu *NFCServiceServerImpl) SetReportInfo(ctx context.Context, req *pb.ReportInfoReq) (*pb.ReportInfoResp, error) {
	logger.Debugf("SetReportInfo param: %+v", req)
	var (
		res                    *pb.ReportInfoResp
		err                    error
		saveReportSendDetail   int32 = 1
		updateReportSendDetail int32 = 2
		deleteReportSendDetail int32 = 3
	)
	switch req.Ops {
	case saveReportSendDetail:
		// 1. TODO 校验数据

		if req.DetailParam == nil ||
			req.DetailParam.AccountId <= 0 {
			return nil, errors.New("invalid param")
		}
		// 2. 保存数据
		res, err = saveAccountReportSendDetail(req)
		if err != nil {
			logger.Debugf("post save Tag info error: %+v", err)
			return nil, err
		}
	case updateReportSendDetail:
		// 1. TODO 校验数据

		// 2. 保存数据
		res, err = updateAccountReportSendDetail(req)
		if err != nil {
			logger.Debugf("post del Tag info error: %+v", err)
			return nil, err
		}
	case deleteReportSendDetail:
		// 1. TODO 校验数据

		// 2. 保存数据
		res, err = delTagReportSendDetailsList(req)
		if err != nil {
			logger.Debugf("post del Tag info error: %+v", err)
			return nil, err
		}

	}
	return res, nil
}

// 保存ReportSendDetail信息 TODO 这种方式是有很大问题的，不能保证mysql和redis一定都正确操作成功
func saveAccountReportSendDetail(req *pb.ReportInfoReq) (*pb.ReportInfoResp, error) {
	// 校验参数
	// 1. 保存到mysql
	if err := nfc_report.SaveDeviceClockInfo(req); err != nil {
		logger.Errorf("tw.SaveAccountReportSendDetail to mysql fail with error : %+v", err)
		return nil, err
	}

	// 2. 添加到缓存？
	return &pb.ReportInfoResp{Res: &pb.Result{Code: http.StatusOK}}, nil
}

// TODO 这种方式是有很大问题的，不能保证mysql和redis一定都正确操作成功
func delTagReportSendDetailsList(req *pb.ReportInfoReq) (*pb.ReportInfoResp, error) {
	// 1. 从mysql删除
	if err := nfc_report.DeleteDeviceClockInfo(req); err != nil {
		logger.Errorf("tw.DelTagReportSendDetailsList to mysql fail with error : %+v", err)
		return nil, err
	}

	// 2. 从缓存删除?
	return &pb.ReportInfoResp{Res: &pb.Result{Code: http.StatusOK}}, nil
}

// 更新ReportSendDetail信息 TODO 这种方式是有很大问题的，不能保证mysql和redis一定都正确操作成功
func updateAccountReportSendDetail(req *pb.ReportInfoReq) (*pb.ReportInfoResp, error) {
	// 1. 更新mysql
	if err := nfc_report.UpdateDeviceClockInfo(req); err != nil {
		logger.Errorf("tw.SaveAccountReportSendDetail to mysql fail with error : %+v", err)
		return nil, err
	}

	// 2. 更新缓存?
	return &pb.ReportInfoResp{Res: &pb.Result{Code: http.StatusOK}}, nil
}

// 查询调度员设置邮箱等的信息
func (wssu *NFCServiceServerImpl) QueryReportInfo(ctx context.Context, req *pb.ReportSetInfoReq) (*pb.ReportSetInfoResp, error) {
	// 校验参数
	logger.Debugf("QueryReportInfo param: %+v", req)
	if req.AccountId <= 0 {
		return nil, errors.New("invalid params")
	}

	// 查询
	var (
		res = &pb.ReportSetInfoResp{}
		err error
	)
	res.Detail, err = nfc_report.QueryDeviceClockInfo(req)
	if err != nil {
		logger.Debugf("QueryReportInfo QueryDeviceClockInfo err: %+v", err)
		return nil, err
	}
	return res, err
}
