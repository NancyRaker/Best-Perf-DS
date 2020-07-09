/*
@Time : 2019/9/27 16:28 
@Author : yanKoo
@File : mq_model
@Software: GoLand
@Description:
*/
package model


type InterphoneMsg struct {
	Uid       string `json:"uid"`
	MsgType   string `json:"m_type"`
	Md5       string `json:"md5"`
	GId       string `json:"grp_id"`
	FilePath  string `json:"file_path"`
	Timestamp string `json:"timestamp"`
	Duration  string `json:"duration"`
}

type JanusNotifyGrpc struct {
	MsgType   int            `json:"msg_type"`   // 1 是对讲消息， 2 是消息处理回复类消息
	Ip        string         `json:"ip"`         // 消息生产者的服务ip
	PttMsg    *InterphoneMsg `json:"ptt_msg"`    // 语音对讲消息
	SignalMsg *SignalMsg     `json:"signal_msg"` // 消息处理回复类消息
}

type SignalMsg struct {
	SignalType      int              `json:"signal_type"`       // 消息处理回复类消息的消息类型 目前1是创建房间的回复
	CreateGroupResp *CreateGroupResp `json:"create_group_resp"` // 如果signal_type
	Res            *Res            `json:"res"`
}

type CreateGroupResp struct {
	GId          int    `json:"gid"`
	DispatcherId int    `json:"dispatcher_id"`
	ResCode      int    `json:"res_code"`
}

type Res struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
}

/******************************************************************/
// Grpc通知janus
type GrpcNotifyJanus struct {
	MsgType   int         `json:"msg_type"` // janus消费的，grpc通知创建房间等消息的类型 // 其实默认只有2，因为grpc不会发送ptt消息
	Ip        string      `json:"ip"`
	SignalMsg *NotifyJanus `json:"signal_msg"` // 消息处理回复类消息
}

// grpc通知janus的信令类消息
type NotifyJanus struct {
	SignalType     int            `json:"signal_type"`      // 消息处理回复类消息的消息类型 目前1是创建房间的回复
	CreateGroupReq *CreateGroupReq `json:"create_group_req"` // 如果signal_type
}

type CreateGroupReq struct {
	GId          int `json:"gid"`
	GroupName    string `json:"group_name"`
	DispatcherId int `json:"dispatcher_id"`
}

