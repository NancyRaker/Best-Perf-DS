/*
@Time : 2019/9/27 16:28 
@Author : yanKoo
@File : mq_model
@Software: GoLand
@Description: 与janus通信过程中会用到的一些结构体
*/
package model

import "time"

type MsgObject struct {
	Option *SendMsgOption
	Msg    interface{}
}

type SendMsgOption struct {
	Timeout time.Duration
}

type JanusNotifyGrpc struct {
	MsgType    int             `json:"msg_type"`    // 1 是对讲消息， 2 是消息处理回复类消息， 3 是链路质量类消息
	Ip         string          `json:"ip"`          // 消息生产者的服务ip
	PttMsg     *InterphoneMsg  `json:"ptt_msg"`     // 语音对讲消息
	SignalMsg  *SignalMsg      `json:"signal_msg"`  // 消息处理回复类消息
	QualityMsg *TalkQualityMsg `json:"quality_msg"` // 通话质量类消息
}

type TalkQualityMsg struct {
	UserId              int `json:"user_id"`
	UserStatus          int `json:"user_status"`
	InLinkQuality       int `json:"in_link_quality"`
	InMediaLinkQuality  int `json:"in_media_link_quality"`
	OutLinkQuality      int `json:"out_link_quality"`
	OutMediaLinkQuality int `json:"out_media_link_quality"`
	Rtt                 int `json:"rtt"`
}

type InterphoneMsg struct {
	Uid       string `json:"uid"`
	MsgType   string `json:"m_type"`
	Md5       string `json:"md5"`
	GId       string `json:"grp_id"`
	FilePath  string `json:"file_path"`
	Timestamp string `json:"timestamp"`
	Duration  string `json:"duration"`
}

type SignalMsg struct {
	SignalType           int                   `json:"signal_type"`                      // 消息处理回复类消息的消息类型 目前1是创建房间的回复
	CreateGroupResp      *CreateGroupResp      `json:"create_group_resp,omitempty"`      // 如果signal_type
	AlterParticipantResp *AlterParticipantResp `json:"alter_participant_resp,omitempty"` // 改变设备在房间的角色
	Res                  *Res                  `json:"res"`
}

type CreateGroupResp struct {
	GId          int `json:"gid"`
	DispatcherId int `json:"dispatcher_id"`
	ResCode      int `json:"res_code"`
}

// 改变调度员名下某个设备在房间里的等级
type AlterParticipantResp struct {
	Gid          int `json:"gid"`
	Uid          int `json:"uid"`
	Role         int `json:"role"`
	DispatcherId int `json:"dispatcher_id"`
}
type Res struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
}

/******************************************************************/
// Grpc通知janus
type GrpcNotifyJanus struct {
	MsgType   int          `json:"msg_type"` // janus消费的，grpc通知创建房间等消息的类型 // 其实默认只有2，因为grpc不会发送ptt消息
	Ip        string       `json:"ip"`
	SignalMsg *NotifyJanus `json:"signal_msg"` // 消息处理回复类消息
}

// grpc通知janus的信令类消息
type NotifyJanus struct {
	SignalType          int                  `json:"signal_type"`                     // 消息处理回复类消息的消息类型 目前1是创建房间的回复
	CreateGroupReq      *CreateGroupReq      `json:"create_group_req,omitempty"`      // 如果signal_type
	DestroyGroupReq     *DestroyGroupReq     `json:"destroy_group_req,omitempty"`     // 删除群组通知消息
	AlterParticipantReq *AlterParticipantReq `json:"alter_participant_req,omitempty"` // 改变设备在房间的角色
}

// 创建房间
type CreateGroupReq struct {
	GId          int    `json:"gid"`
	GroupName    string `json:"group_name"`
	DispatcherId int    `json:"dispatcher_id"`
}

// 删除房间
type DestroyGroupReq struct {
	GId          int `json:"gid"`
	DispatcherId int `json:"dispatcher_id"`
}

// 改变调度员名下某个设备在房间里的等级
type AlterParticipantReq struct {
	Gid          int `json:"gid"`
	Uid          int `json:"uid"`
	Role         int `json:"role"`
	DispatcherId int `json:"dispatcher_id"`
}
