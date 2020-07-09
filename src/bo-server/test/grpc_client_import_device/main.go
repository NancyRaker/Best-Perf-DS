/**
* @Author: yanKoo
* @Date: 2019/3/25 14:57
* @Description:
 */
package main

import (
	"bo-server/test/grpc_client_import_device/device_pressure"
	"container/list"
	"fmt"
	"google.golang.org/grpc"
	"runtime"
	"sync"
)

const (
	FIRST_LOGIN_DATA                = 1 // 初次登录返回的数据。比如用户列表，群组列表，该用户的个人信息
	OFFLINE_IM_MSG                  = 2 // 用户离线时的IM数据
	IM_MSG_FROM_UPLOAD_OR_WS_OR_APP = 3 // APP和web通过httpClient上传的文件信息、在线时通信的im数据
	KEEP_ALIVE_MSG                  = 4 // 用户登录后，每隔interval秒向stream发送一个消息，测试能不能连通
	LOGOUT_NOTIFY_MSG               = 5 // 用户掉线之后，通知和他在一个组的其他成员
	LOGIN_NOTIFY_MSG                = 6 // 用户上线之后，通知和他在一个组的其他成员

	IM_MSG_FROM_UPLOAD_RECEIVER_IS_USER  = 1 // APP和web通过httpClient上传的IM信息是发给个人
	IM_MSG_FROM_UPLOAD_RECEIVER_IS_GROUP = 2 // APP和web通过httpClient上传的IM信息是发给群组

	USER_OFFLINE = 1 // 用户离线
	USER_ONLINE  = 2 // 用户在线

	UNREAD_OFFLINE_IM_MSG = 1 // 用户离线消息未读
	READ_OFFLINE_IM_MSG   = 2 // 用户离线消息已读

	CLIENT_EXCEPTION_EXIT = -1 // 客户端异常终止

	NOTIFY = 1 // 通知完一个
)

var maps sync.Map

type client struct {
	m   sync.RWMutex
	cli *grpc.ClientConn
	e   *list.Element
}

func init() {
	// 设置内核cpu数目
	runtime.GOMAXPROCS(runtime.NumCPU())
}

func main() {
	//host := "113.108.62.203:"
	//host := "121.14.149.182:9001"   //东莞
	//host := "127.0.0.1:9001"
	host := "10.0.10.28:9001"
	//host := "172.20.10.9:9001"
	//host := "23.101.8.213:9001"
	conn, err := grpc.Dial(host, grpc.WithInsecure(), )
	if err != nil {
		fmt.Printf("grpc.Dial err : %v", err)
	} else {
		fmt.Printf("--->%s\nstatu:%+v\n", conn.Target(), conn.GetState())
	}
	defer conn.Close()

	// 导入设备
	//importTestDevice(conn)

	// im 压测
	device_pressure.NewDeviceImPressure(0, 3000, conn).Run()
	select {}
}
