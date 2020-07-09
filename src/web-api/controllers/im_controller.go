/*
@Time : 2019/4/11 10:30
@Author : yanKoo
@File : file_controller
@Software: GoLand
@Description:
*/
package controllers

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"net/http"
	"strconv"
	"time"
	cfgWs "web-api/config"
	"web-api/dao/session"
	"web-api/engine/fdfs_client"
	"web-api/engine/grpc_pool"
	"web-api/logger"
	"web-api/model"
	pb "web-api/proto/talk_cloud"
	"web-api/utils"
)

var upGrader = websocket.Upgrader{
	//HandshakeTimeout: time.Duration(600),
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

const (
	MAX_UPLOAD_SIZE = 1024 * 1024 * 500 // 1024 byte * 1024 * 500 = 500mb

	FIRST_LOGIN_DATA                = 1  // 初次登录返回的数据。比如用户列表，群组列表，该用户的个人信息
	OFFLINE_IM_MSG                  = 2  // 用户离线时的IM数据
	IM_MSG_FROM_UPLOAD_OR_WS_OR_APP = 3  // APP和web通过httpClient上传的IM信息
	KEEP_ALIVE_MSG                  = 4  // 用户登录后，每隔interval秒向stream发送一个消息，测试能不能连通
	LOGOUT_NOTIFY_MSG               = 5  // 用户掉线之后，通知和他在一个组的其他成员
	LOGIN_NOTIFY_MSG                = 6  // 用户上线之后，通知和他在一个组的其他成员
	SOS_MSG                         = 7  // 版本不同，暂定这里
	SOS_CANCLE_MSG                  = 8  // 用户按SOS取消按键呼救
	WEB_JANUS_NOTIFY                = 9  // web用户操作群组通知app
	APP_JANUS_NOTIFY                = 10 // APP用户切换janus房间操作群组通知web

	IPS_CHANGED_NOTIFY = 20 // ips改变，重新登陆

	JANUS_TALK_QUALITY_MSG = 21 // janus通话质量消息

	IM_MSG_WORKDONE  = 1
	IM_MSG_WORKWRONG = -1

	IM_TEXT_MSG  = 1 // 普通文本
	IM_IMAGE_MSG = 2 // 图片
	IM_VOICE_MSG = 3 // 音频文件
	IM_VIDEO_MSG = 4 // 视频文件
	IM_PTT_MSG   = 5 // ptt音频文件
	LOG_FILE_MSG = 6 // 设备日志文件

	IM_UNKNOWN_TYPE_MSG = 10000
	IM_REFRESH_WEB      = 10001 // web刷新界面
	WEB_REPEATED_LOGIN  = 10002 // web重复登录
)

type worker struct {
	uId        int32
	cliStream  *pb.TalkCloud_DataPublishClient
	ws         *websocket.Conn
	Data       chan interface{}
	mt         int
	WorkerDone chan int
}

var client *fdfs_client.Client

func init() {
	var err error
	client, err = fdfs_client.NewClientWithConfig()
	if err != nil {
		logger.Errorf("Client: %+v NewClientWithConfig fastdfs error: %+v", client, err)
	}
	// defer client.Destory() TODO destory?
}

// @Summary websocket与grpc交换数据
// @Produce  json
// @Param accountId path string true "当前用户的账号Id"
// @Success 200 {string} json "{"message":"User information obtained successfully",	"account_info": ai,	"device_list" deviceAll,"group_list": gList,"tree_data":resp}"
// @Router /im-server/:accountId [get]
func ImPush(c *gin.Context) {
	uidStr := c.Param("accountId")
	uid, _ := strconv.Atoi(uidStr) // TODO 校验用户是否存在

	logger.Debugln("im push uid :", uid)

	//升级get请求为webSocket协议
	ws, err := upGrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		logger.Debugf("Upgrade ws %+v connect grpc fail with error: %s", ws, err.Error())
		return
	}
	defer ws.Close()
	sInfo, err := session.GetSessionValue("", uid)
	if err != nil {
		logger.Debugf("ImPush GetSessionValue Start connect grpc fail with error: %s", err.Error())
		return
	}

	var (
		header metadata.MD // variable to store header and trailer
		conn   *grpc.ClientConn
	)
	if cli := grpc_pool.GRPCManager.GetGRPCConnClientById(uid); cli == nil {
		logger.Debugf("ImPush getGRPCConnClientById Start require grpc fail with error: %s", err.Error())
		return

	} else {
		conn = cli.ClientConn
		defer cli.Close()
	}
	userClient := pb.NewTalkCloudClient(conn)
	res, err := userClient.Login(context.Background(), &pb.LoginReq{Name: sInfo.UserName, Passwd: sInfo.UserPwd,
		AppVersion: "web", GrpcServer: "127.0.0.1"}, grpc.Header(&header))
	if err != nil {
		logger.Debugf("connect grpc login fail with error: %s", err.Error())
		return
	}

	fmt.Println(res, err, "\n", header.Get("session-id"))
	var sIds []string
	if sIds = header.Get("session-id"); sIds == nil || len(sIds) <= 0 {
		logger.Debugln("connect grpc fail login with error")
		return
	}
	ctx := metadata.AppendToOutgoingContext(context.Background(), "session-id", sIds[0])
	// 调用调用GRPC接口，转发数据

	webCliStream, err := userClient.DataPublish(ctx)
	if err != nil {
		logger.Debugf("Start connect grpc fail with webCliStream occur error: %s", err.Error())
		return
	}

	imWorker := &worker{
		uId:        int32(uid),
		cliStream:  &webCliStream,
		ws:         ws,
		Data:       make(chan interface{}, 1),
		WorkerDone: make(chan int, 1),
	}

	// 有两种跳出循环的情况：
	// 1、web端主动关闭连接，grpc也就要不再接受数据，
	// 2、web端重复登录，TODO 放在这里判断重复登录有点不妥当，不过如果前面的登录做得好，这里不会出现这种情况，以防万一吧。
	logger.Debugln(strconv.FormatInt(int64(imWorker.uId), 10) + " ws grpc start")

	ctx, cancel := context.WithCancel(context.Background())

	// 接收web端的消息，转发给grpc
	go pushImMessage(imWorker, ctx)

	// 发送ws消息
	go sendImMessage(imWorker, ctx)

	if wd := <-imWorker.WorkerDone; wd == IM_MSG_WORKWRONG {
		_ = imWorker.ws.WriteMessage(websocket.TextMessage,
			[]byte("The connection with id:"+strconv.FormatInt(int64(imWorker.uId), 10)+" has been disconnected, please reconnect"))
		logger.Debugln("websocket exit...")
		cancel()
		return
	} else {
		// TODO grpc服务主动拒绝连接（重复登录）
	}
	//}(imWorker, &wg)
	//wg.Wait()
}

func pushImMessage(imw *worker, ctx context.Context) {
	// 发送给GRPC
	if err := (*imw.cliStream).Send(&pb.StreamRequest{
		Uid:      imw.uId,
		DataType: OFFLINE_IM_MSG,
	}); err != nil {
		imw.WorkerDone <- IM_MSG_WORKWRONG
		logger.Debugln("im message send error: ", err)
		return
	}

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
				// 读取ws中的数据，这里的数据，默认只有文本数据
				mt, message, err := imw.ws.ReadMessage()
				imw.mt = mt
				if err != nil {
					// 客户端关闭连接时也会进入
					logger.Debugf("%d WS message read error: %s", imw.uId, err.Error())

					// 通知im ws主动断开连接
					if err := sendToGrpc(imw, &model.ImMsgData{Id: int(imw.uId), MsgType: IM_REFRESH_WEB}); err != nil {
						imw.WorkerDone <- IM_MSG_WORKWRONG
						logger.Debugln("ws cancel grpc im message send error: ", err)
						return
					}
					imw.WorkerDone <- IM_MSG_WORKWRONG // TODO
					return
				}

				logger.Debugf("ws receive msg: %+v", message)
				wsImMsg := &model.ImMsgData{}
				if err := json.Unmarshal(message, wsImMsg); err != nil {
					logger.Debugf("json unmarshal fail with err :%v", err)
					// TODO  暂时忽略这条消息
					continue
				}

				if wsImMsg.MsgType != IM_REFRESH_WEB {
					// 暂时默认发过来的消息都是普通文本
					wsImMsg.MsgType = IM_TEXT_MSG
				}

				// 发送给GRPC
				logger.Debugf("ws will send to grpc: %+v", wsImMsg)
				if err := sendToGrpc(imw, wsImMsg); err != nil {
					imw.WorkerDone <- IM_MSG_WORKWRONG
					logger.Debugln("grpc im message send error: ", err)
					return
				}

				// web主动关闭ws连接
				if wsImMsg.MsgType == 10001 {
					imw.WorkerDone <- IM_MSG_WORKWRONG
					return
				}

			}
		}
	}()

	// 发送心跳
	tick := time.NewTicker(time.Second * time.Duration(cfgWs.Interval))
	for {
		select {
		case <-ctx.Done():
			return
		case <-tick.C:
			if err := (*imw.cliStream).Send(&pb.StreamRequest{
				Uid:      int32(imw.uId),
				DataType: KEEP_ALIVE_MSG,
			}); err != nil {
				imw.WorkerDone <- IM_MSG_WORKWRONG
				logger.Debugln("grpc im heartbeat message send error: ", err)
				return
			}
		}
	}
}
func sendToGrpc(imw *worker, wsImMsg *model.ImMsgData) error {
	if err := (*imw.cliStream).Send(&pb.StreamRequest{
		Uid:      int32(imw.uId),
		DataType: IM_MSG_FROM_UPLOAD_OR_WS_OR_APP,
		ImMsg: &pb.ImMsgReqData{
			Id:           int32(wsImMsg.Id),
			SenderName:   wsImMsg.SenderName,
			ReceiverId:   int32(wsImMsg.ReceiverId),
			ReceiverName: wsImMsg.ReceiverName,
			SendTime:     wsImMsg.SendTime,
			ReceiverType: int32(wsImMsg.ReceiverType),
			ResourcePath: wsImMsg.ResourcePath, // 文本消息直接放路劲这个字段
			MsgType:      int32(wsImMsg.MsgType),
		},
	}); err != nil {
		return err
	}
	return nil
}

// 从grpc stream中接收数据通过ws转发给web
func sendImMessage(imw *worker, ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
			resp, err := (*imw.cliStream).Recv()
			if err != nil {
				imw.WorkerDone <- IM_MSG_WORKWRONG
				logger.Debugf("%d grpc recv message error: %s", imw.uId, err.Error())
				break
			}
			logger.Debugf("%d web grpc client receive : %+v", imw.uId, resp)

			// 写入ws数据 二进制返回
			if resp.DataType == IM_MSG_FROM_UPLOAD_OR_WS_OR_APP {
				// 把中文转换为utf-8,时间转换
				convertIm(resp.ImMsgData)
				logger.Debugf("web grpc client receive : %+v", resp)

				// 返回JSON字符串
				err = imw.ws.WriteJSON(resp)
				if err != nil {
					imw.WorkerDone <- IM_MSG_WORKWRONG
					logger.Debugln("WS message send error:", err)
					// break
				}

				// web重复登录，关闭ws连接
				logger.Debugf("Web # %d repeat login close ws", imw.uId)
				if resp.ImMsgData.MsgType == WEB_REPEATED_LOGIN {
					// 通知im ws主动断开连接
					if err := sendToGrpc(imw, &model.ImMsgData{Id: int(imw.uId), MsgType: IM_REFRESH_WEB}); err != nil {
						imw.WorkerDone <- IM_MSG_WORKWRONG
						logger.Debugln("ws cancel grpc im message send error: ", err)
						return
					}
					imw.WorkerDone <- IM_MSG_WORKWRONG
					return
				}
			}

			if resp.DataType == OFFLINE_IM_MSG {
				// 把中文转换为utf-8
				convertEncode(resp.OfflineImMsgResp.OfflineGroupImMsgs)
				convertEncode(resp.OfflineImMsgResp.OfflineSingleImMsgs)
				convertEncode(resp.OfflineImMsgResp.OfflineGroupPttImMsgs)
				logger.Debugf("web grpc client receive : %+v", resp)

				// 返回JSON字符串
				err = imw.ws.WriteJSON(resp)
				if err != nil {
					imw.WorkerDone <- IM_MSG_WORKWRONG
					logger.Debugln("WS message send error:", err)
					//break
				}
			}

			// 掉线通知
			if resp.DataType == LOGOUT_NOTIFY_MSG || resp.DataType == LOGIN_NOTIFY_MSG ||
				resp.DataType == WEB_JANUS_NOTIFY || resp.DataType == APP_JANUS_NOTIFY ||
				resp.DataType == IPS_CHANGED_NOTIFY {
				err = imw.ws.WriteJSON(resp)
				if err != nil {
					imw.WorkerDone <- IM_MSG_WORKWRONG
					logger.Debugln("WS message send error:", err)
				}
			}

			// sos通知
			if resp.DataType == SOS_MSG || resp.DataType == SOS_CANCLE_MSG {
				sosNotify := &model.SosNotify{DataType: resp.DataType, Info: &model.SosInfo{}}
				if resp != nil && resp.Notify != nil {
					if resp.Notify.UserInfo != nil {
						sosNotify.Info.UID = resp.Notify.UserInfo.Id
						sosNotify.Info.Imei = resp.Notify.UserInfo.Imei
						sosNotify.Info.Name = resp.Notify.UserInfo.NickName
						sosNotify.Info.Online = resp.Notify.UserInfo.Online
						sosNotify.Info.DeviceType = resp.Notify.UserInfo.DeviceType
					}

					if resp.Notify.GpsResp != nil && resp.Notify.GpsResp.GpsInfo != nil {
						sosNotify.Info.LocalTime = resp.Notify.GpsResp.GpsInfo.LocalTime
						sosNotify.Info.Longitude = resp.Notify.GpsResp.GpsInfo.Longitude
						sosNotify.Info.Latitude = resp.Notify.GpsResp.GpsInfo.Latitude
						sosNotify.Info.Speed = resp.Notify.GpsResp.GpsInfo.Speed
						sosNotify.Info.Course = resp.Notify.GpsResp.GpsInfo.Course
					}

					if resp.Notify.GpsResp != nil && resp.Notify.GpsResp.DeviceInfo != nil {
						sosNotify.Info.Battery = resp.Notify.GpsResp.DeviceInfo.Battery
					}

					if resp.Notify.GpsResp != nil && resp.Notify.GpsResp.WifiInfos != nil {
						sosNotify.Info.WifiDes = resp.Notify.GpsResp.WifiInfos.Des
					}
				}

				err = imw.ws.WriteJSON(sosNotify)
				if err != nil {
					imw.WorkerDone <- IM_MSG_WORKWRONG
					logger.Debugln("WS message send error:", err)
					//break
				}
			}

			// 通话质量类消息
			if resp.DataType == JANUS_TALK_QUALITY_MSG {
				imw.sendTalkQualityMsg(resp)
			}
		}
	}
}

func (imw *worker) sendTalkQualityMsg(resp *pb.StreamResponse) {
	qt := &model.TalkQualityNotify{DataType: resp.DataType, QualityMsg: &model.TalkQualityMsg{}}
	if resp != nil && resp.QualityMsg != nil {
		qt.QualityMsg.UserId = int(resp.QualityMsg.UserId)
		qt.QualityMsg.UserStatus = int(resp.QualityMsg.UserStatus)
		qt.QualityMsg.InLinkQuality = int(resp.QualityMsg.InLinkQuality)
		qt.QualityMsg.InMediaLinkQuality = int(resp.QualityMsg.InMediaLinkQuality)
		qt.QualityMsg.OutLinkQuality = int(resp.QualityMsg.OutLinkQuality)
		qt.QualityMsg.OutMediaLinkQuality = int(resp.QualityMsg.OutMediaLinkQuality)
	}
	err := imw.ws.WriteJSON(qt)
	if err != nil {
		imw.WorkerDone <- IM_MSG_WORKWRONG
		logger.Debugln("WS sendTalkQualityMsg message send error:", err)
		//break
	}
}

// 转换中文
func convertEncode(m []*pb.OfflineImMsg) {
	for _, msg := range m {
		if msg.ImMsgData != nil {
			for _, userMsg := range msg.ImMsgData {
				convertIm(userMsg)
			}
		}
	}
}

func convertIm(userMsg *pb.ImMsgReqData) {
	userMsg.MsgCode = userMsg.SendTime
	userMsg.SendTime = utils.UnixStrToWebTimeFormat(userMsg.SendTime)
	logger.Debugf("web conv offline im time: %s %s", userMsg.MsgCode, userMsg.SendTime)
	//userMsg.SendTime = utils.ConvertOctonaryUtf8(userMsg.SendTime)
	userMsg.ResourcePath = utils.ConvertOctonaryUtf8(userMsg.ResourcePath)
}
