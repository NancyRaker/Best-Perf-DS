/*
@Time : 2019/11/7 11:48 
@Author : yanKoo
@File : pressure_login
@Software: GoLand
@Description:
*/
package device_pressure

import (
	pb "bo-server/api/proto"
	"context"
	"fmt"
	"github.com/360EntSecGroup-Skylar/excelize"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"strconv"
	"time"
)

type deviceImPressure struct {
	start   int
	end     int
	devices []*device
	conn    *grpc.ClientConn
}

type device struct {
	imei   string
	passwd string
	id     string
}

const (
	FIRST_LOGIN_DATA                = 1 // 初次登录返回的数据。比如用户列表，群组列表，该用户的个人信息
	OFFLINE_IM_MSG                  = 2 // 用户离线时的IM数据
	IM_MSG_FROM_UPLOAD_OR_WS_OR_APP = 3 // APP和web通过httpClient上传的文件信息、在线时通信的im数据
	KEEP_ALIVE_MSG                  = 4 // 用户登录后，每隔interval秒向stream发送一个消息，测试能不能连通
)

func NewDeviceImPressure(start, end int, conn *grpc.ClientConn) *deviceImPressure {
	return &deviceImPressure{
		start: start,
		end:   end,
		conn:  conn,
	}
}

// 读取imei号，用来登录
func (dip *deviceImPressure) readDeviceFormExcel() {
	f, err := excelize.OpenFile("./压测设备.xlsx")
	if err != nil {
		panic(err)
	}
	// Get all the rows in the Sheet1.
	deviceSource := f.GetRows("Sheet1")
	var devices []*device
	deviceSource = deviceSource[1:]
	for _, row := range deviceSource {
		devices = append(devices, &device{
			id:     row[0],
			imei:   row[1],
			passwd: row[2],
		})
	}
	dip.devices = devices
}
func (dip deviceImPressure) login(index int) string {
	ctx := context.Background()
	var header metadata.MD
	userClient := pb.NewTalkCloudClient(dip.conn)
	res, err := userClient.Login(ctx, &pb.LoginReq{
		Name:       dip.devices[index].imei,
		Passwd:     dip.devices[index].passwd,
		AppVersion: "test",
		GrpcServer: "127.0.0.1",
	}, grpc.Header(&header))
	fmt.Printf("res : %s\nerr: %+v\nseesionId:%+v\n", res, err, header.Get("session-id"))
	matedata := header.Get("session-id")
	if len(matedata) == 0 {
		return ""
	}
	return matedata[0]
}

func (dip deviceImPressure) imPush(index int, sessionId string) {
	i, _ := strconv.Atoi(dip.devices[index].id)

	client := pb.NewTalkCloudClient(dip.conn)

	// create a new context with some metadata
	ctx := metadata.AppendToOutgoingContext(context.Background(), "session-id", sessionId)

	allStr, err := client.DataPublish(ctx)

	fmt.Println("------------->", err)
	if err := allStr.Send(&pb.StreamRequest{
		Uid:      int32(i),
		DataType: OFFLINE_IM_MSG,
	}); err != nil {
		fmt.Println("=============>", err)
	}

	go func(i int) {
		for {
			//fmt.Println("start send heartbeat")
			if err := allStr.Send(&pb.StreamRequest{
				Uid:        int32(i),
				DataType:   KEEP_ALIVE_MSG,
				DeviceInfo: &pb.DeviceInfo{Battery: 58, Charge: 1},
			}); err != nil {
			}

			time.Sleep(time.Second * 60)
		}
	}(i)

	go func() {
		for {
			if err := allStr.Send(&pb.StreamRequest{
				Uid:      int32(i),
				DataType: IM_MSG_FROM_UPLOAD_OR_WS_OR_APP,
				ImMsg: &pb.ImMsgReqData{
					Id:           int32(i),
					SenderName:   "xiaoliu",
					ReceiverType: 1,
					ReceiverId:   int32(i - 1),
					ResourcePath: "SOSOS",
					MsgType:      6,
					ReceiverName: "xx group",
					SendTime:     strconv.FormatInt(time.Now().Unix(), 10),
				},
			}); err != nil {
			}
			time.Sleep(time.Second * 75)
		}

	}()

	go func(allStr *pb.TalkCloud_DataPublishClient) {
		i := 1
		for {
			data, _ := (*allStr).Recv()
			if data != nil {
				if data.DataType == KEEP_ALIVE_MSG {
					fmt.Printf("%d client receive: %d # %d \n", time.Now().UnixNano(), data.KeepAlive.Uid, i)
					i++
				} else if data.DataType == OFFLINE_IM_MSG {
					fmt.Println("client receive: 2", data.OfflineImMsgResp)
				} else if data.DataType == IM_MSG_FROM_UPLOAD_OR_WS_OR_APP {
					fmt.Println("client receive: 2", data.ImMsgData)
				} else {
					fmt.Printf("%+v\n", data)
				}
			}
		}
	}(&allStr)
}

func (dip deviceImPressure) Run() {
	// 读取xlsx文件，获取压测数据
	dip.readDeviceFormExcel()
	for i := dip.start; i < dip.end; i++ {
		sessionId := dip.login(i)
		dip.imPush(i, sessionId)
	}
}
