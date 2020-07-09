/*
@Time : 2019/11/7 11:51 
@Author : yanKoo
@File : import_device
@Software: GoLand
@Description:
*/
package device_pressure

import (
	pb "bo-server/api/proto"
	"context"
	"fmt"
	"google.golang.org/grpc"
	"strconv"
	"time"
)

func importTestDevice( conn *grpc.ClientConn){
	// 导入设备
	var devices []*pb.DeviceInfo
	for i := 0; i < 2997; i++ {
		time.Sleep(time.Nanosecond)
		imeiStr := strconv.FormatInt(time.Now().UnixNano(), 10)
		fmt.Println(imeiStr[2 : len(imeiStr)-2])
		imei := imeiStr[2 : len(imeiStr)-2]
		devices = append(devices, &pb.DeviceInfo{
			Imei:       imei,
			DeviceType: "JW10",
		})
	}

	cli := pb.NewWebServiceClient(conn)
	res, err := cli.ImportDeviceByRoot(context.Background(), &pb.ImportDeviceReq{
		AccountId: 220,
		Devices:devices,
	})
	fmt.Println(res, err)
}