/*
@Time : 2019/9/24 17:05 
@Author : yanKoo
@File : main
@Software: GoLand
@Description:
*/
package main

import (
	"bo-server/internal/mq/mq_sender"
	"bo-server/model"
	"encoding/json"
	"time"
)

func main() {
	for i := 0; i < 10; i++ {
		go func() {
			//time.Sleep(time.Second * 3)
			msg := &model.JanusNotifyGrpc{
				MsgType: 3,
				//Ip:      "127.0.0.1",
				//PttMsg: &model.InterphoneMsg{
				//	Uid:       "62",
				//	MsgType:   "ptt",
				//	Md5:       "md5",
				//	//GId:       "528",
				//	GId:       "1492",
				//	Timestamp: "1568968731",
				//	FilePath:  "https://dev.yunptt.com:82/group1/M00/00/00/wKhkBl2Jd8SAVwq-AABVgIc4lZw066.mp3",
				//},
				QualityMsg:&model.TalkQualityMsg{
					UserId:62,
					UserStatus:2,
					InLinkQuality:100,
					InMediaLinkQuality:0,
					OutLinkQuality:i,
					OutMediaLinkQuality:2,
					Rtt:10,
				},
			}
			/*msg := &model.GrpcNotifyJanus{
				MsgType: mq_sender.SignalType,
				Ip:      "127.0.0.1",
				SignalMsg: &model.NotifyJanus{
					SignalType: mq_sender.CreateGroupReq,
					CreateGroupReq: &model.CreateGroupReq{
						GId:          999,
						GroupName:   "xxx",
						DispatcherId: 1000,
					},
				},
			}*/

			/*msg := &model.JanusNotifyGrpc{
				MsgType: mq_sender.SignalType,
				Ip:      "127.0.0.1",
				SignalMsg: &model.SignalMsg{
					SignalType: mq_sender.CreateGroupReq,
					CreateGroupResp: &model.CreateGroupResp{
						GId:          999,
						DispatcherId: 1000,
					},
					Res: &model.Res{
						Code: 1,
						Msg:  "",
					},
				},
			}*/

			b, _ := json.Marshal(msg)
			mq_sender.SendMsg(&model.MsgObject{Msg: b})
		}()
	}
	time.Sleep(time.Hour * 1)
}
