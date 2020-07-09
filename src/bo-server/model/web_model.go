/**
* @Author: yanKoo
* @Date: 2019/3/16 15:32
* @Description:
 */
package model

import (
	"database/sql"
)

// request
type GroupList struct {
	DeviceIds  []int         `json:"device_ids"`
	DeviceInfo []interface{} `json:"device_infos"`
	GroupInfo  *GroupInfo    `json:"group_info"`
}

// Data model
type SessionInfo struct {
	SessionId string `json:"session_id"`
	Id        int32  `json:"id"`
	Online    int32  `json:"online"`
}

// device
type Device struct {
	Id         int     `json:"id"`
	IMei       string  `json:"imei"`
	UserName   string  `json:"user_name"`
	NickName   string  `json:"nick_name"`
	PassWord   string  `json:"password"`
	AccountId  int     `json:"account_id"`
	CreateTime string  `json:"create_time"`
	LLTime     string  `json:"last_login_time"`
	ChangeTime string  `json:"change_time"`
	LocalTime  uint64  `json:"local_time"`
	GPSData    *GPS    `json:"gps_data"`
	Speed      float32 `json:"speed"`
	Course     float32 `json:"course"`
	DeviceType string  `json:"device_type"`
	ActiveTime string  `json:"active_time"`
	SaleTime   string  `json:"sale_time"`
	Online     int     `json:"online"`
	WifiDes    string  `json:"wifi_des"`

	UserType  int `json:"user_type"` // 用户类型(暂定1是普通用户，2是调度员，3是经销商, 4是超级管理员)
	GroupType int `json:"group_type"`
}

type GPS struct {
	Lng float64 `json:"lng"`
	Lat float64 `json:"lat"`
}

// User
type User struct {
	Id          int    `json:"id"`
	IMei        string `json:"imei"`
	UserName    string `json:"user_name"`
	NickName    string `json:"nick_name"`
	PassWord    string `json:"password"`
	UserType    int    `json:"user_type"` // 用户类型(暂定1是普通用户，2是调度员，3是经销商, 4是超级管理员)
	StartLog    int    `json:"start_log"`  // 开启app日志 0是关闭状态，1是开启状态
	ParentId    string `json:"parent_id"`  // 如果是普通APP用户和设备注册的时候，默认是0， 如果是上级用户创建下级账户，就用来表示创建者的id
	AccountId   int    `json:"account_id"` // 只有普通用户才有这个字段，表示这个设备属于哪个账户，如果是非普通用户就是默认为0（因为customer表里面没有0号）
	LockGroupId int    `json:"lock_group_id"`
	CreateTime  string `json:"create_time"`
	LLTime      string `json:"last_login_time"`
	ChangeTime  string `json:"change_time"`
	Online      int    `json:"online"`
	DeviceType  string `json:"device_type"`
	ActiveTime  string `json:"active_time"`
	SaleTime    string `json:"sale_time"`
}

type GroupInfo struct {
	Id        int    `json:"id"`
	GroupName string `json:"group_name"`
	AccountId int    `json:"account_id"`
	Status    string `json:"status"`
	CTime     string `json:"c_time"`
	OnlineNum int    `json:"online_num"`
}

type Account struct {
	Id          int            `json:"id"`
	Pid         int            `json:"pid"`
	Username    string         `json:"username" example:"elephant"`
	NickName    string         `json:"nick_name"`
	Pwd         string         `json:"pwd" example:"123456"`
	Email       sql.NullString `json:"email"`
	PrivilegeId int            `json:"privilege_id"`
	Contact     sql.NullString `json:"contact"`
	RoleId      int            `json:"role_id"`
	State       string         `json:"state"`
	LlTime      string         `json:"ll_time"`
	ChangeTime  string         `json:"change_time"`
	CTime       string         `json:"c_time"`
	Phone       sql.NullString `json:"phone"`
	Remark      sql.NullString `json:"remark"`
	Address     sql.NullString `json:"address"`
}

type AccountUpdate struct {
	Id       string `json:"id"`
	LoginId  string `json:"login_id"`
	NickName string `json:"nick_name"`
	Username string `json:"username"`
	TypeId   string `json:"type_id"`
	Phone    string `json:"phone"`
	Email    string `json:"email"`
	Address  string `json:"address"`
	Remark   string `json:"remark"`
	Contact  string `json:"contact"`
}

type ImMsgData struct {
	Id           int    `json:"id"`
	SenderName   string `json:"SenderName"`
	ReceiverType int    `json:"ReceiverType"`
	ReceiverId   int    `json:"ReceiverId"`
	ReceiverName string `json:"ReceiverName"`
	ResourcePath string `json:"ResourcePath"`
	MsgType      int    `json:"MsgType"`
	SendTime     string `json:"SendTime"`
}

type FileContext struct {
	UserId         int
	FilePath       string
	FileType       int32
	FileParams     *ImMsgData
	FileName       string
	FileSize       int
	FileMD5        string
	FileFastId     string
	FileUploadTime string
}

// app登录时候保存登录ip结构体
type ServerInfo struct {
	Type     int    `json:"type"`
	Ip       string `json:"ip"`
	Port     string `json:"port"`
	Location string `json:"location"`
}

// 上线下线提醒
type NotifyInfo struct {
	Id         int32 `json:"id"`
	NotifyType int32 `json:"notify_type"`
}
