/**
* @Author: yanKoo
* @Date: 2019/3/25 14:57
* @Description:
 */
package main

import (
	pb "bo-server/api/proto"
	"bufio"
	"container/list"
	"context"
	"encoding/json"
	"fmt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"io/ioutil"
	"math"
	"os"
	"runtime"
	"strconv"
	"sync"
	"time"
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

//const GROUP_PORT = ":9001"
//const GROUP_PORT = ":9999"

//const GROUP_PORT = ":9998" //qps测试

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
	//_ = flag.String("p", "23.101.8.213:9001", "grpc server addr")
	//ids := flag.String("id", "64", "device id")
	//groupIdstr := flag.String("gid", "1", "group id")
	//sendtype := flag.String("t", "s", "group id")
	//flag.Parse()
	//
	//var msgType int32
	//if *sendtype == "s" {
	//	msgType = 2
	//}
	//if *sendtype == "c"{
	//	msgType = 3
	//}
	//
	////
	//id,_ := strconv.Atoi(*ids)
	//gid, _ := strconv.Atoi(*groupIdstr)
	//host := "113.108.62.203:"
	//host := "121.14.149.182:9001"   //东莞
	//host := "127.0.0.1:9001"
	host := "10.0.10.28:9001"
	//host := "172.20.10.9:9001"
	//host := "23.101.8.213:9001"
	conn, err := grpc.Dial(host, grpc.WithInsecure(), )
	//r := etcd_lb.NewResolver(*reg, *svc)

	//resolver.Register(r)
	////ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	//// https://github.com/grpc/grpc/blob/master/doc/naming.md
	//// The gRPC client library will use the specified scheme to pick the right resolver plugin and pass it the fully qualified name string.
	//fmt.Printf("-------->%s", r.Scheme())
	//conn, err := grpc.Dial(r.Scheme()+"://authority/"+*svc, grpc.WithInsecure(), grpc.WithBalancerName(roundrobin.Name), grpc.WithBlock())
	//cancel()
	//conn, err := grpc.Dial("47.100.116.26:2379/iM_server", grpc.WithInsecure(), grpc.WithBalancerName("round_robin"))
	if err != nil {
		fmt.Printf("grpc.Dial err : %v", err)
	} else {
		fmt.Printf("--->%s\nstatu:%+v\n", conn.Target(), conn.GetState())
	}
	defer conn.Close()
	//c := client{
	//	cli: conn,
	//}

	//res, err := webCli.ImportDeviceByRoot(context.Background(), &pb.ImportDeviceReq{
	//	DeviceImei:[]string{"1234567897777777"},
	//	AccountId: 1,
	//})

	// 调用调用GRPC接口，转发数据
	/*for i := 0; i <1000; i++ {
		go func() {
			res, err := webCli.ImMessagePublish(context.Background(), &pb.ImMsgReqData{
				Id:           8,
				SenderName:   "xiaoliu",
				ReceiverType: 2,
				ReceiverId:   229,
				ResourcePath: "SOSOS",
				MsgType:      6,
				ReceiverName: "xx group",
				SendTime:     strconv.FormatInt(time.Now().Unix(), 10),
			})

			fmt.Printf("res:%+v err : %+v",res, err)
		}()
	}
	select {}*/
	/*webCli := pb.NewTalkCloudClient(conn)
	for i:= 0; i < 1500; i++ {
		time.Sleep(time.Microsecond*300)
		go func() {
			res, err := webCli.ImMessagePublish(context.Background(), &pb.ImMsgReqData{
				Id:           8,
				SenderName:   "iron man",
				ReceiverType: 2,
				ReceiverId:   229,
				ResourcePath: "SOS",
				MsgType:      3,
				ReceiverName: "boot",
				SendTime:     "hhhhh",
				MsgCode:      strconv.FormatInt(time.Now().Unix(), 10),
			})
			fmt.Printf("%+v, err:%+v",res.Result.Msg, err)
		}()
	}*/
	/*res, err := userClient.AppRegister(context.Background(), &pb.AppRegReq{
		Name:     "355172100003878",
		Password: "123456",
	})
	fmt.Printf(res.String())*/

	/*res, err := webCli.DeleteGroup(context.Background(), &pb.Group{
		Id: 102,
	})*/

	/*res, err := userClient.AddFriend(context.Background(), &pb.FriendNewReq{
		Uid:333,
		Fuid:500,
	})*/

	for i := 0; i < 0; i++ {
		webCli := pb.NewWebServiceClient(conn)
		res, err := webCli.MultiTransDevice(context.Background(), &pb.TransDevices{
			Imeis:      []string{"355172100056157"},
			ReceiverId: 7,
			SenderId:   115,
		})
		fmt.Println(res, err)
	}

	/*   565 566 567 568 569 570 571 572 */
	for i := 0; i < 0; i++ {
		uids := []int32{332, 332, 332, 332, 332, 332, 332, 332, 332, 332, 332, 332, 332, 332, 332, 332, 332, 332, 332,
			332, 332, 332, 332, 332, 332, 332, 332, 332, 332, 332, 332, 332, 332, 332, 332, 332, 332, 332, 332, 332,
			332, 332, 332, 332, 332, 332, 332, 332, 332, 332, 332, 332, 332}

		gids := []int32{2074, 2076, 2078, 2080, 2082, 2084, 2086, 2088, 2090, 2092, 2094, 2096, 2098, 2100, 2102, 2104, 2106, 2108, 2110, 2112, 2114, 2116, 2118, 2120, 2122, 2124, 2126, 2128, 2130, 2132, 2134, 2136, 2138, 2140, 2142, 2144, 2146, 2148, 2150, 2152, 2154, 2156, 2158, 2160, 2162, 2164, 2166, 2168, 2170, 2172, 2174, 217, 2178}
		cli := pb.NewTalkCloudClient(conn)
		for i := range uids {
			res, err := cli.RemoveGrp(context.Background(), &pb.GroupDelReq{
				Uid: uids[i],//211,
				Gid: gids[i],//631,
			})
			fmt.Println(res, err)
		}
	}

	for i := 0; i < 0; i++ {
		cli := pb.NewWebServiceClient(conn)
		res, err := cli.WebCreateGroup(context.Background(), &pb.WebCreateGroupReq{
			DeviceIds: []int32{65},
			GroupInfo: &pb.Group{
				Status:    1, //普通组
				AccountId: 62,
			},
		})
		fmt.Println(res, err)
	}

	//fmt.Println("---------------------------------------Login Start-------------------------------------------")
	ctx := context.Background()
	var header metadata.MD // variable to store header and trailer
	//var data = []byte{}
	for i := 0; i <= 0; i++ {
		//time.Sleep(time.Second*3)
		start := time.Now().UnixNano()
		//fmt.Printf("login start time:%d\n", start)
		userClient := pb.NewTalkCloudClient(conn)
		res, _ := userClient.Login(ctx, &pb.LoginReq{
			//Name:   "865305032789213",   // 大门
			//Passwd: "789213",
			//Name:   "454512127878963",  //224
			//Passwd: "878963",
			//Name:   "999900007777123",  //dev 186
			//Passwd: "777123",
			//Name:   "355172100001534", //dev 17
			//Passwd: "001534",
			//Name:   "989876545678888",  //dev 189
			//Passwd: "678888",
			//Name:   "351609083508690", // 218 xiang218
			//Passwd: "508690",
			//Name:   "355172100056157", //liu yang dong guan 62
			//Passwd: "056157",
			Name:           "355172100087962", //  219
			Passwd:         "087962",
			//Name:   "355172100054251", // leikun dongguan 65
			//Passwd: "054251",
			//AppVersionCode: 10,
			AppVersion: "test",
			GrpcServer: "127.0.0.1",
		}, grpc.Header(&header))

		end := time.Now().UnixNano()
		fmt.Printf("res : %s\nerr: %+v\nseesionId:%+v\n", res, err, header.Get("session-id"))
		fmt.Printf("login start time:%d; login end time:%d; used time: %d ms\n", start, end, (end-start)/1000000)
		//data = append(data, []byte(fmt.Sprintf("login start time:%d; login end time:%d; used time: %d ms\n",start, end, (end-start)/1000000))...)
	}
	//jsonRes, err := json.Marshal(data)

	//_ = ioutil.WriteFile("test_micro.log", []byte(data), 0644)
	/*fmt.Println("---------------------------------------Login Start-------------------------------------------")
	res, err := userClient.RemoveGrp(context.Background(), &pb.GroupDelReq{
		Gid:8,
	})
	fmt.Println(res)*/

	//time.Sleep(time.Second * 3)
	/*
		for _, v := range res.GroupList {
			if v.GroupManager == -1 {
				fmt.Printf("find a no mannage group: %d", v.Gid)
			}
		}*/

	//fmt.Printf("this user groups:%d, all:%+v", len(res.GroupList), res)

	/*	res, err := userClient.InviteUserIntoGroup(context.Background(), &pb.InviteUserReq{
			Uids:"1536",
			Gid:208,
		})

		fmt.Println(ress, "---------------++++++",  err)*/

	/*res, err := userClient.SearchUserByKey(context.Background(), &pb.UserSearchReq{
		Uid:uint64(333),
		Target:"",
	})*/
	/*res, err := userClient.GetFriendList(context.Background(), &pb.FriendsReq{
		Uid:333,
	})*/
	/*for i := 0; i < 3000; i++ {
		go func(i int) {
	//	*/
	for i := 0; i < 0; i++ {
		fmt.Println("---------------------------------------GetGroupList Start-------------------------------------------")
		userClient := pb.NewTalkCloudClient(conn)
		groupList, err := userClient.GetGroupList(context.Background(), &pb.GrpListReq{
			Uid: int32(24),
		})
		fmt.Printf("%+v Get group list **************************** # %+v\n", err, len([]byte(groupList.String())))
		fmt.Printf("%+v Get group list **************************** # %+v\n", err, groupList)

		groupListJson, _ := json.Marshal(groupList)
		_ = ioutil.WriteFile("group-list.json", groupListJson, 0644)
		//fmt.Println("---------------------------------------GetGroupInfo Start-------------------------------------------")
		//groupInfo, err := userClient.GetGroupInfo(context.Background(), &pb.GetGroupInfoReq{
		//	Gid: int32(1870),
		//	Uid: int32(101),
		//})
		//groupInfoJson, _ := json.Marshal(groupInfo)
		//_ = ioutil.WriteFile("group-info.json", groupInfoJson, 0644)
		//fmt.Printf("%+v Get group info **************************** # %+v\n", err, groupInfo)
	}

	/*
		}(i)

		go func() {
			ress, err := userClient.InviteUserIntoGroup(context.Background(), &pb.InviteUserReq{
				Uids:"457",
				Gid:210,
			})
			fmt.Println(ress, "*******---------++++++",  err)
			fmt.Printf("InviteUserIntoGroup **************************** # %d", i)
		}()
	}*/

	for i := 0; i < 0; i++ {
		fmt.Println("---------------------------------------Create group Start-------------------------------------------")
		userClient := pb.NewTalkCloudClient(conn)
		create, err := userClient.CreateGroup(context.Background(), &pb.CreateGroupReq{
			DeviceIds: []int32{65}, // 要邀请进群的id
			GroupInfo: &pb.Group{
				GroupName: "20191104g",
				AccountId: 62, // 当前设备id
				Status:    1,  // 暂时把这种对讲组当做普通组
			}})
		fmt.Printf("%+v>>>>>>>>>>>>>>>>>>>>>create:%+v", err, create)
	}

	for i := 0; i < 0; i++ {
		fmt.Println("---------------------------------------del group Start-------------------------------------------")
		userClient := pb.NewTalkCloudClient(conn)
		create, err := userClient.RemoveGrp(context.Background(), &pb.GroupDelReq{
			Uid: 188, // 要邀请进群的id
			Gid: 613})
		fmt.Printf("%+v>>>>>>>>>>>>>>>>>>>>>create:%+v", err, create)
	}

	for i := 0; i < 0; i++ {
		fmt.Println("---------------------------------------web del group Start-------------------------------------------")
		userClient := pb.NewWebServiceClient(conn)
		create, err := userClient.DeleteGroup(context.Background(), &pb.GroupDelReq{
			Uid: 188, // 要邀请进群的id
			Gid: 613})
		fmt.Printf("%+v>>>>>>>>>>>>>>>>>>>>>create:%+v", err, create)
	}

	//fmt.Println("---------------------------------------GetGroupList again Start-------------------------------------------")
	//glAgain, err := userClient.GetGroupList(context.Background(), &pb.GrpListReq{
	//	Uid: int32(333),
	//})
	//fmt.Printf("%+v Get group list **************************** # %+v", err, glAgain)

	/*res, err := userClient.SearchGroup(context.Background(), &pb.GrpSearchReq{
		Uid:uint64(333),
		Target:"雷坤",
	})
	*/

	/*res , err := userClient.SearchUserByKey(context.Background(), &pb.UserSearchReq{
		Uid:333,
		Target:"121422",
	})*/

	/*res, err := userClient.JoinGroup(context.Background(), &pb.GrpUserAddReq{
		Uid: 1514,
		Gid: 151,
	})*/

	//fmt.Println(res, "---------------",  err)
	for i := 0; i < 0; i++ {
		userClient := pb.NewTalkCloudClient(conn)
		res, err := userClient.SetLockGroupId(context.Background(), &pb.SetLockGroupIdReq{
			UId: 62,
			GId: 193,
		})
		fmt.Printf("%+v, %+v", res, err)
	}

	// GPS 数据
	deviceCli := pb.NewTalkCloudLocationClient(conn)

	for j := int32(0); j < 0; j++ {
		i := int32(189) //int32(id)
		wifiInfo := make([]*pb.Wifi, 0)
		//for i := 0; i < 3; i++ {
		wifiInfo = append(wifiInfo, &pb.Wifi{
			Level: int32(-1000),
			BssId: "123465681352" + strconv.Itoa(100),
		})
		wifiInfo = append(wifiInfo, &pb.Wifi{
			Level: int32(-62),
			BssId: "123465681352" + strconv.Itoa(200),
		})
		wifiInfo = append(wifiInfo, &pb.Wifi{
			Level: int32(-6),
			BssId: "12:dataContext:A9:4a:03:9F",
		})
		//}
		res, err := deviceCli.ReportGPSData(context.Background(), &pb.ReportDataReq{
			IMei:     "989876545678888",
			DataType: 1, // 1是gps
			DeviceInfo: &pb.Device{
				Id:           i,
				Battery:      78,
				DeviceType:   1,
				CurrentGroup: int32(1498),
			},
			LocationInfo: &pb.Location{
				GpsInfo: &pb.GPS{
					LocalTime: uint64(time.Now().Unix() - int64(j)*10000),
					Longitude: float64(120.13795866213755) + float64(float64(0.001)*math.Pow(-1, float64(j+1))*float64(j)),
					Latitude:  float64(28.480194593114472) + float64(float64(0.001)*math.Pow(-1, float64(j))*float64(j)),
					Course:    200,
					Speed:     float32(123.4),
				},
				BSInfo: &pb.BaseStation{
					//Country: 460,
					//Operator: 1,
					//Lac:42705,
					//Cid: 228408571,
					//FirstBs: -49,
				},
				//BtInfo: &pb.BlueTooth{
				//	//FirstBt: -93,
				//	//SecondBt: -89,
				//	//ThirdBt:-98,
				//},
				WifiInfo: wifiInfo,
			},
		})
		//if msgType == 3 {
		fmt.Printf("%+v, %s, err :%+v", res, "send cancel sos", err)
		//} else if msgType == 2 {
		//	fmt.Printf("%+v, %s, err :%+v", res, "send sos", err)
		//}
		//}
		//if err != nil {
		//	fmt.Println(err)
		//} else {
		//	fmt.Printf("%+v", len(res.GroupList))

	}
	/*client := pb.NewTalkCloudClient(c.cli)
	resp, err := client.GetServerAddr(context.TODO(), &pb.ServerAddrReq{Name: "aa"})
	fmt.Printf("%+v, %+v", resp, err)*/
	//TODO 服务端 客户端 双向流
	fmt.Println(time.Now().UnixNano(), "Start ------------------>")
	for j := 0; j <= 0; j++ {
		//i := 62 // dong guan liuyang
		//i := 65 // dongguan leikun
		//i := 218 // dev
		i := 219 // dev

		//i := 186 // dev
		//i := 189 // dev
		//i := 14
		//go func(i int) {
		//time.Sleep(time.Millisecond*100*time.Duration(i))
		//if i == 2299 || i == 300{fmt.Println("#", i, "Start at ==========>", time.Now().UnixNano())}
		//c.m.lock()
		client := pb.NewTalkCloudClient(conn)
		var m sync.Mutex
		// create a new context with some metadata
		ctx = metadata.AppendToOutgoingContext(context.Background(), "session-id", /*"cd2632f7-b5c8-4697-a0f6-10c9bf660188")//*/ (header.Get("session-id"))[0])

		//for j := 300; j < 20; j++ {
		//	i := 215
		//i := 62  // dong guan
		//	//i := 17 //
		//	//i := 1700 // 355172100003878
		//	go func(i int) {
		//time.Sleep(j*time.Millisecond)
		m.Lock()
		defer m.Unlock()
		allStr, err := client.DataPublish(ctx)

		//fmt.Printf("%d start send get offline msg", i)
		fmt.Println("------------->", err)
		if err := allStr.Send(&pb.StreamRequest{
			Uid:      int32(i),
			DataType: OFFLINE_IM_MSG,
		}); err != nil {
			fmt.Println("=============>", err)
		}
		//c.m.Unlock()

		go func(i int) {
			//for k := 0; k < 0; k++{
			for {
				//fmt.Println("start send heartbeat")
				if err := allStr.Send(&pb.StreamRequest{
					Uid:        int32(i),
					DataType:   KEEP_ALIVE_MSG,
					DeviceInfo: &pb.DeviceInfo{Battery: 58, Charge: 1},
				}); err != nil {
				}

				time.Sleep(time.Second * 5)
			}
		}(i)
		/*
						go func(i int) {

							if err := allStr.Send(&pb.StreamRequest{
								Uid:     int32(i),
								DataType: IM_MSG_FROM_UPLOAD_OR_WS_OR_APP,
								ImMsg:&pb.ImMsgReqData{
									SenderName:   "xiaoliu",
									ReceiverType: 2,
									ReceiverId:   229,
									ResourcePath: "SOSOS",
									MsgType:      6,
									ReceiverName: "xx group",
									SendTime:     strconv.FormatInt(time.Now().Unix(), 10),
								},
							}); err != nil {
							}

							time.Sleep(time.Second * 3)
						}(i)
			*/
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
		//for  {
		//	time.Sleep(time.Second*3)
		//	fmt.Println(conn.GetState())
		//}
		for {
			//读键盘
			reader := bufio.NewReader(os.Stdin)
			//以换行符结束
			str, _ := reader.ReadString('\n')
			if err := allStr.Send(&pb.StreamRequest{
				Uid:      int32(i),
				DataType: IM_MSG_FROM_UPLOAD_OR_WS_OR_APP,
				ImMsg: &pb.ImMsgReqData{
					Id:           int32(i),
					SenderName:   "xiaoliu",
					ReceiverType: 1,
					ReceiverId:   115,
					ResourcePath: str,
					MsgType:      1,
					ReceiverName: "xx group",
					SendTime:     strconv.FormatInt(time.Now().Unix(), 10),
				},
			}); err != nil {
			}

			fmt.Println(time.Now().UnixNano(), "send end------------------->")
		}
		//}(i)

	}
	//var chans chan int
	//<- chans
}
