/*
@Time : 2019/4/4 14:18
@Author : yanKoo
@File : web_device_impl
@Software: GoLand
@Description: 实现web端需要用到的关于设备管理需要用到的GRpc接口
*/
package web

import (
	pb "bo-server/api/proto"
	"bo-server/api/server/im"
	cfgGs "bo-server/conf"
	td "bo-server/dao/device"
	tfi "bo-server/dao/file_info"
	tg "bo-server/dao/group"         // table group
	tgm "bo-server/dao/group_member" // table group_device
	tlc "bo-server/dao/location"
	"bo-server/dao/pub"
	tu "bo-server/dao/user"
	tuc "bo-server/dao/user_cache"
	tw "bo-server/dao/wifi"
	twc "bo-server/dao/wifi_cache"
	"bo-server/engine/cache"
	"bo-server/logger"
	"bo-server/model"
	"bo-server/utils"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"time"
)

var (
	SAVE_WIFI_INFO   int32 = 1
	UPDATE_WIFI_INFO int32 = 2
	DELETE_WIFI_INFO int32 = 3
)
// 批量导入设备
func (wssu *WebServiceServerImpl) ImportDeviceByRoot(ctx context.Context, req *pb.ImportDeviceReq) (*pb.ImportDeviceResp, error) {
	// 设备串号和账户id进行校验
	logger.Error("start Import DeviceByRoot")
	for _, v := range req.Devices {
		if v == nil || v.Imei == "" {
			return &pb.ImportDeviceResp{
				Result: &pb.Result{
					Msg:  "Import device Imei can't be empty! please try again later.",
					Code: http.StatusUnprocessableEntity,
				},
			}, errors.New("import device imei is empty")
		}
	}

	// 只有root用户有权限导入设备 TODO 鉴权 应该有一个root独有的签名
	/*
		return &pb.ImportDeviceResp{
			Result: &pb.Result{
				Msg:  "Only the root account can import devices.",
				Code: http.StatusUnprocessableEntity,
			},
		}, errors.New("account is unauthorized")
	*/

	devices := make([]*model.User, 0)
	var imeis []string
	for _, v := range req.Devices {
		imeis = append(imeis, v.Imei)
		devices = append(devices, &model.User{
			IMei:       v.Imei,
			UserName:   v.Imei,
			NickName:   string([]byte(v.Imei)[12:len(v.Imei)]),
			PassWord:   string([]byte(v.Imei)[9:len(v.Imei)]),
			AccountId:  int(req.GetAccountId()),
			ParentId:   "1", // 设备不存在pid这个字段，没有意义，所以写成超级管理员也无妨
			DeviceType: v.DeviceType,
			ActiveTime: v.ActiveTime,
			SaleTime:   v.SaleTime,
		})
	}

	if err := td.ImportDevice(&devices); err != nil {
		return &pb.ImportDeviceResp{
			Result: &pb.Result{
				Msg:  "Import device error, please try again later.",
				Code: http.StatusInternalServerError,
			},
		}, err
	}

	// 设备更新到缓存 TODO
	for _, device := range devices {
		_ = tuc.AddUserDataInCache(int32(device.Id), []interface{}{
			pub.USER_Id, int32(device.Id),
			pub.IMEI, device.IMei,
			pub.USER_NAME, device.IMei,
			pub.ACCOUNT_ID, device.AccountId,
			pub.NICK_NAME, device.NickName,
			pub.USER_TYPE, 1,             // 导入设备默认都是1
			pub.ONLINE, pub.USER_OFFLINE, // 加载数据默认全部离线
			pub.DEVICE_TYPE, device.DeviceType,
		}, cache.GetRedisClient())
	}

	// 发送post请求给web-gateway,注册设备前缀树 TODO 错误处理
	addrs := strings.Split(cfgGs.WebGatewayAddrs, " ")
	for _, addr := range addrs {
		utils.WebGateWay{Url: addr + cfgGs.AddDeviceUrl}.ImportDevicePost(&utils.DeviceImportReq{
			Imeis:     imeis,
			AccountId: req.AccountId,
		})
	}

	return &pb.ImportDeviceResp{
		Result: &pb.Result{
			Msg:  "import device successful.",
			Code: http.StatusOK,
		},
	}, nil
}

//修改设备信息
func (wssu *WebServiceServerImpl) UpdateDeviceInfo(ctx context.Context, req *pb.UpdDInfoReq) (*pb.UpdDInfoResp, error) {
	// TODO 实在是毫无扩展性，烂代码
	logger.Debugf("update Device Info with %+v", req)
	if req == nil || req.DeviceInfo == nil {
		return nil, errors.New("invalid params")
	}

	if err := td.UpdateDeviceInfo(req.DeviceInfo); err != nil {
		return &pb.UpdDInfoResp{
			Res: &pb.Result{Msg: "Update DeviceInfo device error, please try again later.",
				Code: http.StatusInternalServerError,
			},
		}, err
	}

	// 更新缓存
	var updatePair []interface{}
	if req.DeviceInfo.NickName != "" {
		updatePair = append(updatePair, pub.NICK_NAME)
		updatePair = append(updatePair, req.DeviceInfo.NickName)
	}
	if req.DeviceInfo.StartLog != "" {
		startLog, _ := strconv.Atoi(req.DeviceInfo.StartLog)
		updatePair = append(updatePair, pub.START_LOG)
		updatePair = append(updatePair, startLog)
		// 通知设备开启或者关闭日志
		// 通知web
		msg := &pb.StreamResponse{
			Uid: req.DeviceInfo.Id,
		}
		if startLog == 0 {
			// 关闭
			msg.DataType = im.LOG_TURN_OFF_NOTIFY
		} else if startLog == 1 {
			// 打开
			msg.DataType = im.LOG_TURN_ON_NOTIFY
		}
		go im.SendSingleNotify([]int32{req.DeviceInfo.Id}, req.DeviceInfo.Id, msg)
	}

	if err := tuc.AddUserDataInCache(req.DeviceInfo.Id, updatePair, cache.GetRedisClient()); err != nil {
		return &pb.UpdDInfoResp{
			Res: &pb.Result{
				Msg:  "Update DeviceInfo device error, please try again later.",
				Code: http.StatusInternalServerError,
			},
		}, err
	}
	go func() {
		if req.DeviceInfo.NickName != "" {
			im.SendSingleNotify([]int32{req.DeviceInfo.Id}, req.DeviceInfo.LoginId, &pb.StreamResponse{
				DataType: im.DEVICE_NICKNAME_CHANGE_MSG,
				Notify:   &pb.LoginOrLogoutNotify{UserInfo: &pb.DeviceInfo{NickName: req.DeviceInfo.NickName, Id: req.DeviceInfo.Id}}})
		}
	}()

	return &pb.UpdDInfoResp{Res: &pb.Result{Msg: "Update DeviceInfo device successful!", Code: http.StatusOK,},}, nil

}

func (wssu *WebServiceServerImpl) SelectDeviceByImei(ctx context.Context, req *pb.ImeiReq) (*pb.ImeiResp, error) {
	id, err := td.SelectDeviceByImei(&model.User{IMei: req.Imei})
	resp := &pb.ImeiResp{
		Res: &pb.Result{
			Msg:  "SelectDeviceByImei error, please try again later.",
			Code: http.StatusInternalServerError,
		},
	}
	resp.Id = id
	if err != nil {
		return resp, err
	}
	return resp, nil
}

func (wssu *WebServiceServerImpl) GetDevicesInfo(ctx context.Context, req *pb.DeviceInfosReq) (*pb.DeviceInfosResp, error) {
	var resp = &pb.DeviceInfosResp{Devices: make([]*pb.DeviceInfo, 0), Res: &pb.Result{Code: http.StatusOK, Msg: "query gps info success"}}
	// 1. 获取所有用户设备
	resp.Devices = tuc.GetUserBaseInfo(req.DevicesIds)
	// 2. GPS数据
	err := tlc.GetUsersLocationFromCache(&resp.Devices, cache.GetRedisClient())
	if err != nil {
		return &pb.DeviceInfosResp{Res: &pb.Result{Code: http.StatusInternalServerError, Msg: "query gps info fail"}}, nil
	}

	// 3. 在线状态
	var statMap = make(map[int32]bool)
	err = tuc.QueryDevicesStatus(&resp.Devices, &statMap)
	if err != nil {
		return &pb.DeviceInfosResp{Res: &pb.Result{Code: http.StatusInternalServerError, Msg: "query gps info fail"}}, nil
	}
	return resp, nil
}

// web 保存，修改，删除wifi信息
func (wssu *WebServiceServerImpl) PostWifiInfo(ctx context.Context, req *pb.WifiInfoReq) (*pb.WifiInfoResp, error) {
	fmt.Printf("receive data from client: %+v", req)
	var (
		res *pb.WifiInfoResp
		//errResp = &pb.WifiInfoResp{Res:&pb.Result{Msg:"process error, please try again later"}}
		err error
	)
	// 1. TODO 校验数据
	if len(req.Wifis) == 0 || req.Wifis == nil {
		return &pb.WifiInfoResp{Res: &pb.Result{Code: http.StatusOK}}, nil
	}
	switch req.Ops {
	case SAVE_WIFI_INFO:
		// 1. TODO 校验数据

		// 2. 保存数据
		res, err = saveWifiInfo(req)
		if err != nil {
			logger.Debugf("post save wifi info error: %+v", err)
			return nil, err
		}
	case UPDATE_WIFI_INFO:
		// 1. TODO 校验数据

		// 2. 保存数据
		res, err = updateWifiInfo(req)
		if err != nil {
			logger.Debugf("post del wifi info error: %+v", err)
			return nil, err
		}
	case DELETE_WIFI_INFO:
		// 1. TODO 校验数据

		// 2. 保存数据
		res, err = delWifiInfo(req)
		if err != nil {
			logger.Debugf("post del wifi info error: %+v", err)
			return nil, err
		}

	}
	return res, nil
}

// 更新wifi信息 TODO 这种方式是有很大问题的，不能保证mysql和redis一定都正确操作成功
func updateWifiInfo(wifiInfo *pb.WifiInfoReq) (*pb.WifiInfoResp, error) {
	// 1. 更新mysql
	if err := tw.UpdateWifiInfo(wifiInfo); err != nil {
		logger.Errorf("tw.SaveWifiInfo to mysql fail with error : %+v", err)
		return nil, err
	}

	// 2. 更新缓存
	if err := twc.UpdateWifiInfoToCache(wifiInfo); err != nil {
		logger.Errorf("twc.SaveWifiInfoToCache to redis fail with error : %+v", err)
		return nil, err
	}
	return &pb.WifiInfoResp{Res: &pb.Result{Code: http.StatusOK}}, nil
}

// 保存wifi信息 TODO 这种方式是有很大问题的，不能保证mysql和redis一定都正确操作成功
func saveWifiInfo(wifiInfo *pb.WifiInfoReq) (*pb.WifiInfoResp, error) {
	// 0. 校验bssid的唯一性，查重
	for _, w := range wifiInfo.Wifis {
		if wifi, err := twc.CheckWifiInfoFromCache(w.BssId); err != nil {
			logger.Errorf("tw.SaveWifiInfo duplicate fail with error : %+v", err)
			return nil, err
		} else {
			if wifi {
				return &pb.WifiInfoResp{Res: &pb.Result{Code: http.StatusBadRequest, Msg: "mac addr " + w.BssId + " is already saved"}}, nil
			}
		}
	}

	// 1. 保存到mysql
	if err := tw.SaveWifiInfo(wifiInfo); err != nil {
		logger.Errorf("tw.SaveWifiInfo to mysql fail with error : %+v", err)
		return nil, err
	}

	// 2. 添加到缓存
	if err := twc.SaveWifiInfoToCache(wifiInfo); err != nil {
		logger.Errorf("twc.SaveWifiInfoToCache to redis fail with error : %+v", err)
		return nil, err
	}
	return &pb.WifiInfoResp{Res: &pb.Result{Code: http.StatusOK}}, nil
}

// TODO 这种方式是有很大问题的，不能保证mysql和redis一定都正确操作成功
func delWifiInfo(wifiInfo *pb.WifiInfoReq) (*pb.WifiInfoResp, error) {

	// 1. 从mysql删除
	if err := tw.DelWifiInfo(wifiInfo); err != nil {
		logger.Errorf("tw.DelWifiInfo to mysql fail with error : %+v", err)
		return nil, err
	}

	// 2. 从缓存删除
	if err := twc.DelWifiInfoToCache(wifiInfo); err != nil {
		logger.Errorf("twc.DelWifiInfoToCache to redis fail with error : %+v", err)
		return nil, err
	}
	return &pb.WifiInfoResp{Res: &pb.Result{Code: http.StatusOK}}, nil
}

// 调度员获取wifi信息
func (wssu *WebServiceServerImpl) GetWifiInfo(ctx context.Context, req *pb.GetWifiInfoReq) (*pb.GetWifiInfoResp, error) {
	if req.AccountId < 0 {
		return nil, errors.New("account id is invalid")
	}

	var (
		wifis = make([]*pb.Wifi, 0)
		err   error
	)

	if wifis, err = twc.GetWifiInfoByAccount(req.AccountId); err != nil {
		return nil, err
	}

	if wifis != nil && len(wifis) > 0 {
		randomQuickSort(&wifis, 0, len(wifis))
	}

	return &pb.GetWifiInfoResp{Wifis: wifis}, nil
}

func randomQuickSort(list *[]*pb.Wifi, start, end int) {
	if end-start > 1 {
		// get the pivot
		mid := randomPartition(list, start, end)
		randomQuickSort(list, start, mid)
		randomQuickSort(list, mid+1, end)
	}
}

func randomPartition(list *[]*pb.Wifi, begin, end int) int {
	// 生成真随机数
	r := randInt(begin, end)
	// 下面这行是核心部分，随机选择主元，如果没有此次交换，就是普通快排
	(*list)[r], (*list)[begin] = (*list)[begin], (*list)[r]
	return partition(list, begin, end)
}

func partition(list *[]*pb.Wifi, begin, end int) (i int) {
	cValue := (*list)[begin].Id
	i = begin
	for j := i + 1; j < end; j++ {
		if (*list)[j].Id < cValue {
			i++
			(*list)[j], (*list)[i] = (*list)[i], (*list)[j]
		}
	}
	(*list)[i], (*list)[begin] = (*list)[begin], (*list)[i]
	return i
}

// 真随机数
func randInt(min, max int) int {
	rand.Seed(time.Now().UnixNano())
	return min + rand.Intn(max-min)
}

// 查询日志信息
func (wssu *WebServiceServerImpl) GetDeviceLogInfo(c context.Context, req *pb.GetDeviceLogInfoReq) (*pb.GetDeviceLogInfoResp, error) {
	resp := &pb.GetDeviceLogInfoResp{LogUrl: ""}
	apkInfo, err := tfi.GetFileInfo(req.Uid)
	if err != nil && err != sql.ErrNoRows {
		return resp, err
	}
	if err == sql.ErrNoRows {
		return resp, nil
	}
	resp.LogUrl = cfgGs.FILE_BASE_URL + apkInfo.FileFastId
	return resp, nil
}

// 获取某个设备某一段时间的经纬度
func (wssu *WebServiceServerImpl) GetGpsForTrace(c context.Context, req *pb.GpsForTraceReq) (*pb.GpsForTraceResp, error) {
	var (
		err       error
		traceData = make([]*pb.TraceInfo, 0)
		total     int64
		resp      = &pb.GpsForTraceResp{}
	)
	if req.Id == "" || req.TimesStampEnd == "" || req.TimesStampStart == "" {
		return nil, errors.New("GetGpsForTrace param is invalid")
	}

	logger.Debugf("start get trace data with param: %+v", req)
	if total, err = tlc.GetTimeStampIntervalTotal(req); err != nil {
		logger.Errorf("GetGpsForTrace GetTimeStampIntervalTotal error: %+v", err)
		return nil, errors.New("internal server error")
	}

	if traceData, err = tlc.SelectTraceDataByLocalTime(req); err != nil {
		logger.Errorf("GetGpsForTrace SelectTraceDataByLocalTime error: %+v", err)
		return nil, errors.New("internal server error")
	}

	if total <= int64(len(traceData)) {
		resp.Whole = true
	} else {
		resp.Whole = false
	}
	resp.TraceData = traceData

	return resp, nil
}

//  批量转移设备
func (wssu *WebServiceServerImpl) MultiTransDevice(ctx context.Context, req *pb.TransDevices) (*pb.Result, error) {
	// 1. 校验参数是否合法
	if req == nil || req.ReceiverId <= 0 || req.SenderId <= 0 || len(req.Imeis) <= 0 {
		return nil, errors.New("invalid params")
	}

	var (
		groups    []*model.GroupInfo
		err       error
		imeiMap   = make(map[string]bool)
		imeiIdMap = make(map[int32]bool)
	)
	// 3.0 使用map来装imei
	for _, v := range req.Imeis {
		imeiMap[v] = true
	}
	// 2. 批量转移设备
	//go func() {
	deviceAll, err := tu.SelectUserByAccountId(int(req.SenderId))
	if err != nil {
		logger.Debugf("MultiTransDevice SelectUserByAccountId: %s", err)
	}
	// 需要更新设备调度员
	for _, device := range deviceAll {
		if imeiMap[device.IMei] {
			imeiIdMap[int32(device.Id)] = true
		}
	}
	logger.Debugf("will update id: %+v to account id # %d", imeiIdMap, req.ReceiverId)
	// 4.0 更新设备在缓存中的调度员
	for key := range imeiIdMap {
		if err := tuc.AddUserDataInCache(key, []interface{}{
			pub.ACCOUNT_ID, req.ReceiverId,
			pub.LOCK_GID, pub.DEFAULT_LOCK_GROUP_ID, // 转移设备就把默认janus锁定组置为0
		}, cache.GetRedisClient()); err != nil {
			//if err := tuc.UpdateUserInfoInCache(key, pub.ACCOUNT_ID, req.ReceiverId, cache.GetRedisClient()); err != nil {
			logger.Debugf("multi trans device UpdateGroup is fail : %+v", err)
			//return nil, errors.New("multi trans device UpdateGroup is fail, please try again later")
		}
	}
	//}()

	if err = td.MultiUpdateDevice(req); err != nil {
		logger.Debugf("MultiUpdateDevice db error : %s", err)
		return nil, errors.New("multi trans device is fail, please try again later")
	}

	// 3. 如果该用户名下有群组（也就是调度员）就移除这些设备
	logger.Debugf("will trans device for senderId:%d", req.SenderId)
	// 3.1 获取该账户下所有的群组
	if groups, err = tg.SelectGroupsByAccountId(int(req.SenderId)); err != nil {
		// 会出现转移了，但是群组没移走的情况
		return nil, errors.New("multi trans device is fail, please try again later")
	}
	logger.Debugf("will check groups: %+v", groups)

	// 3.2 遍历群组，找出需要移除的设备
	var updateGroups = make(map[int32][]int32, len(groups))
	for _, g := range groups {
		// 群内成员
		var devideIds = make([]int32, 0)
		devices, err := tgm.SelectDevicesByGroupId(g.Id)
		if err != nil {
			logger.Debugf("Error in Get Group devices: %s", err)
		}
		//log.log.Debugf("MultiTransDevice SelectDevicesByGroupId: %+v", devices)
		for _, d := range devices {
			logger.Debugf("MultiTransDevice SelectDevicesByGroupId: %+v", d)
			if imeiMap[d.IMei] {
				devideIds = append(devideIds, int32(d.Id))
			}
		}
		updateGroups[int32(g.Id)] = devideIds
	}
	logger.Debugf("will remove group mem: %+v", updateGroups)

	// 3.3 移除设备
	var wssi = WebServiceServerImpl{}
	for gId, devices := range updateGroups {
		if _, err := wssi.UpdateGroup(context.TODO(), &pb.UpdateGroupReq{
			RemoveDeviceInfos: devices,
			GroupInfo:         &pb.Group{Id: gId}}); err != nil {
			return nil, errors.New("multi trans device UpdateGroup is fail, please try again later")
		}
	}

	// 5.0 发送post请求给web-gateway,更新设备前缀树 TODO 错误处理
	addrs := strings.Split(cfgGs.WebGatewayAddrs, " ")
	for _, addr := range addrs {
		utils.WebGateWay{Url: addr + cfgGs.UpdateDeviceUrl}.TransDevicePost(&utils.DeviceImportReq{
			Imeis:     req.Imeis,
			AccountId: req.ReceiverId,
		})
	}

	return &pb.Result{Code: http.StatusOK}, nil
}
