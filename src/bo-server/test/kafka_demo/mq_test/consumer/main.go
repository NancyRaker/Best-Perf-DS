/*
@Time : 2019/9/24 17:05 
@Author : yanKoo
@File : main
@Software: GoLand
@Description:
*/
package main

import (
	"bo-server/internal/mq/mq_receiver"
	"fmt"
)

func main() {
	for {
		select {
		case msg:= <- mq_receiver.PttMessage():
			fmt.Println("mq receive ", msg)
			//go sendResp(msg)
		}
	}
}
//func sendResp(req *model.SignalMsg)  {
//	/*msg := &model.GrpcNotifyJanus{
//	MsgType: mq_sender.SignalType,
//	Ip:      "127.0.0.1",
//	SignalMsg: &model.NotifyJanus{
//		SignalType: mq_sender.CreateGroupReq,
//		CreateGroupReq: &model.CreateGroupReq{
//			GId:          999,
//			GroupName:   "xxx",
//			DispatcherId: 1000,
//		},
//	},
//}*/
//
//	msg := &model.JanusNotifyGrpc{
//		MsgType: mq_sender.SignalType,
//		Ip:      "127.0.0.1",
//		SignalMsg: &model.SignalMsg{
//			SignalType: mq_sender.CreateGroupReq,
//			CreateGroupResp: &model.CreateGroupResp{
//				//GId:          req.CreateGroupResp,
//				DispatcherId: 1000,
//			},
//			Res: &model.Res{
//				Code: 1,
//				Msg:  "",
//			},
//		},
//	}
//
//	b, _ := json.Marshal(msg)
//	mq_sender.SendMsg(b)
//}
