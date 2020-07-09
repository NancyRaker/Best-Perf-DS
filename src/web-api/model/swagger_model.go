/*
@Time : 2019/6/13 9:51
@Author : yanKoo
@File : swagger_model
@Software: GoLand
@Description:   这里的全部model都是只是为了swagger文档
*/
package model

// 更新群管理员请求
type SwagUpdGManagerReq struct {
	GId       int32   `json:"GId" example:"2"`
	UId       []int32 `json:"UId" example:"8"`
	AccountId int32   `json:"AccountId" example:"10"`
	RoleType  int32   `json:"RoleType" example:"2"`
}

// 获取账户的某些设备gps信息
type SwagDeviceInfosReq struct {
	DevicesIds []int32 `json:"DevicesIds" example:"2"`
}

// 创建用户
type SwagCreateAccount struct {
	ConfirmPwd string `json:"confirm_pwd" binding:"required" example:"123456"`
	Pid        int    `json:"pid" binding:"required" example:"13"`
	Username   string `json:"username" binding:"required" example:"disks007"`
	NickName   string `json:"nick_name" binding:"required" example:"nickDis001"`
	Pwd        string `json:"pwd" binding:"required" example:"123456"`
	RoleId     int    `json:"role_id" binding:"required,numeric,min=2,max=4" example:"2"`
	Email      string `json:"email" binding:"omitempty,email" example:"123456789@qq.com"`
	Contact    string `json:"contact"`
	Phone      string `json:"phone" binding:"omitempty,numeric" example:"12345678455"`
	Remark     string `json:"remark"`
	Address    string `json:"address"`
}

// 创建用户回复
type SwagCreateAccountResp struct {
	Result    string `json:"result" example:"success"`
	AccountId string `json:"account_id" example:"666"`
}

// 更新账户信息
type SwagAccountUpdate struct {
	Id       string `json:"id" binding:"required" example:"170"`
	LoginId  string `json:"login_id" binding:"required" example:"10"`
	NickName string `json:"nick_name" binding:"required"`
	Phone    string `json:"phone" binding:"omitempty,numeric" example:"12345671234"`
	Email    string `json:"email" binding:"omitempty,email" example:"123456789@qq.com"`
	Address  string `json:"address"`
	Remark   string `json:"remark"`
	Contact  string `json:"contact"`
}

// 更新账户信息回复
type SwagAccountUpdateResp struct {
	Result string `json:"result" example:"success"`
	Msg    string `json:"msg" example:"update account success"`
}

// 获取账户信息响应
type SwagGetAccountInfoResp struct {
	AccountInfo string           `json:"account_info"`
	DeviceList  []*Device        `json:"device_list"`
	GroupList   []*GroupListNode `json:"group_list"`
	TreeData    AccountClass     `json:"tree_data"`
}

// update pwd modal
type SwagAccountPwd struct {
	Id         string `json:"id" binding:"required" example:"10"`
	OldPwd     string `json:"old_pwd" binding:"required" example:"111111"`
	NewPwd     string `json:"new_pwd" binding:"required" example:"123123"`
	ConfirmPwd string `json:"confirm_pwd" binding:"required" example:"123456"`
}

// 更新账户信息回复
type SwagAccountPwdResp struct {
	Result string `json:"result" example:"success"`
	Msg    string `json:"msg" example:"Password changed successfully"`
}

// 更新账户信息回复
type SwagGetAccountClassResp struct {
	Result   string       `json:"result" example:"success"`
	TreeData AccountClass `json:"tree_data"`
}

// 更新账户信息回复
type SwagGetAccountDeviceResp struct {
	AccountInfo Account   `json:"account_info"`
	Devices     []*Device `json:"devices"`
}

// 获取设备gps信息响应
type SwagGetDeviceLocationResp struct {
	Devices []*SwagDeviceInfo `json:"Devices"`
}

type SwagDeviceInfo struct {
	Id         int32   `json:"Id"`
	IMei       string  `json:"IMei"`
	UserName   string  `json:"UserName"`
	PassWord   string  `json:"PassWord"`
	AccountId  int32   `json:"AccountId"`
	CreateTime string  `json:"CreateTime"`
	LLTime     string  `json:"LLTime"`
	ChangeTime string  `json:"ChangeTime"`
	Course     float32 `json:"course"`
	LocalTime  uint64  `json:"localTime"`
	Longitude  float64 `json:"longitude"`
	DeviceType string  `json:"DeviceType"`
	Latitude   float64 `json:"latitude"`
	Speed      float32 `json:"speed"`
	ActiveTime string  `json:"ActiveTime"`
	SaleTime   string  `json:"SaleTime"`
	NickName   string  `json:"NickName"`
	Battery    int32   `json:"battery"`
}

type SwagDeviceInfosResp struct {
	Devices []*SwagDeviceInfo `json:"Devices"`
	Res     *SwagResult       `json:"Res"`
}

// 结果信息
type SwagResult struct {
	Result string `json:"result" example:"success"`
	Msg    string `json:"msg" example:"delete account successfully"`
}

type SwagImportDeviceByRootResp struct {
	ErrDevices  []*SwagDevice `json:"err_devices"`
	DuliDevices []*SwagDevice `json:"duli_devices"`
	Msg         string        `json:"msg"`
}

// 更新账户信息回复
type SwagDeleteLeafNodeAccountResp struct {
	AccountInfo SwagAccount   `json:"account_info"`
	Devices     []*SwagDevice `json:"devices"`
}

// device
type SwagDevice struct {
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
}
type SwagImportDevice struct {
	Id         int    `json:"id" binding:"required"`
	IMei       string `json:"imei" binding:"required"`
	DeviceType string `json:"device_type" binding:"required"`
	ActiveTime string `json:"active_time" binding:"required"`
	SaleTime   string `json:"sale_time" binding:"required"`
}
type SwagAccount struct {
	Id          int    `json:"id"`
	Pid         int    `json:"pid"`
	Username    string `json:"username" example:"elephant"`
	NickName    string `json:"nick_name"`
	Pwd         string `json:"pwd" example:"123456"`
	Email       string `json:"email"`
	PrivilegeId int    `json:"privilege_id"`
	Contact     string `json:"contact"`
	RoleId      int    `json:"role_id"`
	State       string `json:"state"`
	LlTime      string `json:"ll_time"`
	ChangeTime  string `json:"change_time"`
	CTime       string `json:"c_time"`
	Phone       string `json:"phone"`
	Remark      string `json:"remark"`
	Address     string `json:"address"`
}

// 导入设备
type SwagAccountImportDeviceReq struct {
	Devices []*SwagImportDevice `json:"devices" binding:"required"`
	DType   string              `json:"d_type"`
}

// 转移设备
type SwagAccountDeviceTransReq struct {
	Devices  []*SwagDevice  `json:"devices"`
	Receiver DeviceReceiver `json:"receiver"`
	Sender   SwagAccount    `json:"sender"`
}

// 更新账户信息回复
type SwagTransAccountDeviceResp struct {
	Result string `json:"result" example:"success"`
	Msg    string `json:"msg" example:"trans successful"`
}

type SwagDeviceUpdate struct {
	Id         int32  `json:"Id"`
	IMei       string `json:"IMei"`
	NickName   string `json:"NickName"`
	LoginId    int32  `json:"LoginId"`
	CreateTime string `json:"CreateTime"`
	LLTime     string `json:"LLTime"`
	ChangeTime string `json:"ChangeTime"`
	LocalTime  uint64 `json:"LocalTime"`
	DeviceType string `json:"DeviceType"`
	ActiveTime string `json:"ActiveTime"`
	SaleTime   string `json:"SaleTime"`
}

// 更新设备信息回复
type SwagUpdateDeviceInfoResp struct {
	Msg string `json:"msg" example:"update device info successfully"`
}

// 保存wifi信息的数据
type SwagWifiInfoReq struct {
	Wifis     []*SwagWifi `json:"wifis"`
	Ops       int32       `json:"ops"`
	AccountId int32       `json:"account_id"`
}

type SwagWifi struct {
	Level     int32   `json:"level"`
	BssId     string  `json:"bss_id"`
	Longitude float64 `json:"longitude"`
	Latitude  float64 `json:"latitude"`
	Id        int32   `json:"id"`
	Des       string  `json:"des"`
}

type SwagGetDeviceLogResp struct {
	LogUrl string `json:"log_url"`
}