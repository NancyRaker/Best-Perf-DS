/*
@Time : 2019/6/6 17:25
@Author : yanKoo
@File : user_apk
@Software: GoLand
@Description:实现了热更新apk的版本号，让app升级不影响服务
*/
package app

import (
	pb "bo-server/api/proto"
	cfgGs "bo-server/conf"
	tfi "bo-server/dao/file_info"
	tsa "bo-server/dao/server_addr"
	"bo-server/utils"
	"context"
	"errors"
	"fmt"
	"github.com/go-ini/ini"
	"net/http"
	"strconv"
	"sync"
)

// 返回apk消息
func (tcs *TalkCloudServiceImpl) GetApkInfo(ctx context.Context, req *pb.ApkInfoReq) (*pb.ApkInfoResp, error) {
	resp := &pb.ApkInfoResp{Res: &pb.Result{Msg: "Get apk info successful", Code: http.StatusOK}}
	apkInfo, err := tfi.GetFileInfo(req.Uid)
	if err != nil {
		resp.Res.Code = http.StatusInternalServerError
		resp.Res.Msg = "Get apk info fail please try again later"
		return resp, nil
	}
	resp.ApkPath = cfgGs.FILE_BASE_URL + apkInfo.FileMD5
	resp.ApkVersion = apkInfo.FileName
	return resp, nil
}

// 去mysql获取服务地址
func (tcs *TalkCloudServiceImpl) GetServerAddr(ctx context.Context, req *pb.ServerAddrReq) (*pb.ServerAddrRsp, error) {
	if req.Name == "" {
		return nil, errors.New("name can't be nil")
	}

	// TODO 返回策略
	ip, err := utils.GetClietIP(ctx)
	if err != nil {

	}
	fmt.Println("\n==================>", ip)

	res := tsa.GetServerAddr(tsa.JANUS_SERVER)
	if res != nil && len(res) > 0 {
		return &pb.ServerAddrRsp{Addr: res[0].Ip + ":" + res[0].Port}, nil
	}
	return nil, errors.New("can't get server ip")
}

// 更新apk的版本号
type ApkUpdateServiceImpl struct{}

// 更新apk的版本号
func (apkUploadServer *ApkUpdateServiceImpl) UpdateApkInfo(ctx context.Context, req *pb.ApkUpdateInfo) (*pb.Result, error) {
	// 校验参数
	if req.VersionId <= 0 || req.PathUrl == "" {
		return nil, errors.New("params is invalid")
	}

	var m = sync.RWMutex{}
	m.Lock()
	defer m.Unlock()
	cfgGs.AppVersionCode = int(req.VersionId)
	cfgGs.AppUrl = req.PathUrl

	// 修改配置文件 由web-internal做了
	pathName := "./bo_conf.ini"
	cfg, err := ini.Load(pathName)
	if err != nil {
		return nil, errors.New("open conf is fial with error : %+v" + err.Error())
	}
	cfg.Section("app").Key("version_code").SetValue(strconv.Itoa(int(req.VersionId)))
	cfg.Section("app").Key("apk_path").SetValue(req.PathUrl)
	if err := cfg.SaveTo(pathName); err != nil {
		return nil, errors.New("params is invalid")
	}

	return &pb.Result{}, nil
}
