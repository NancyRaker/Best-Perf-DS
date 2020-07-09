/*
@Time : 2019/8/15 11:47 
@Author : yanKoo
@File : key_factory
@Software: GoLand
@Description:
*/
package pub

import (
	"fmt"
	"strconv"
)

const (
	// device 在redis的hash结构中的filed
	USER_Id   = "id"
	IMEI      = "imei"
	USER_NAME = "username"
	NICK_NAME = "nickname"
	ONLINE    = "online"
	LOCK_GID  = "lock_gid"
	ACCOUNT_ID   ="a_id"
	USER_TYPE = "user_type"
	DEVICE_TYPE = "device_type"

	LOCAL_TIME = "local_time"
	LONGITUDE  = "lon"
	LATITUDE   = "lat"
	SPEED_GPS  = "speed"
	COURSE     = "course"
	BATTERY    = "battery"
	WIFI_INFO  = "wifi_info"
)

const (
	GRP_MEM_KEY_FMT    = "grp:%d:mem"
	GRP_DATA_KEY_FMT   = "grp:%d:data"

	USR_DATA_KEY_FMT   = "usr:%d:data"
	USR_STATUS_KEY_FMT = "usr:%d:stat"
	USR_GROUP_KEY_FMT  = "usr:%d:grps"

	USER_OFFLINE = 1 // 用户离线
	USER_ONLINE  = 2 // 用户在线
	USER_JANUS_ONLINE = 3  // JANUS在线

	GROUP_MEMBER  = 1
	GROUP_MANAGER = 2
	GROUP_OWNER   = 3
)


func MakeUserGroupKey(uid int32) string {
	return fmt.Sprintf(USR_GROUP_KEY_FMT, uid)
}

func MakeGroupMemKey(gid int32) string {
	return fmt.Sprintf(GRP_MEM_KEY_FMT, gid)
}

func MakeGroupDataKey(gid int32) string {
	return fmt.Sprintf(GRP_DATA_KEY_FMT, gid)
}

func MakeUserDataKey(uid int32) string {
	return fmt.Sprintf(USR_DATA_KEY_FMT, uid)
}

func GetRedisKey(key int32) string {
	return "app:" + strconv.FormatInt(int64(key), 10) + ":stat"
}