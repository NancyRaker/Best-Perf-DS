/*
@Time : 2019/5/9 20:57
@Author : yanKoo
@File : misc
@Software: GoLand
@Description:
*/
package utils

import (
	"bytes"
	"context"
	"encoding/gob"
	"errors"
	"google.golang.org/grpc/peer"
	cfgComm "bo-server/conf"
	"net"
	"strconv"
	"strings"
	"time"
)

func ConvertTimeUnix(t uint64) string {
	return time.Unix(int64(t), 0).Format(cfgComm.TimeLayout)
}

// 时间戳转换为模板时间
func UnixStrToTimeFormat(tStr string) string {
	t, _ := strconv.ParseInt(tStr, 10, 128)
	return time.Unix(t, 0).Format(cfgComm.TimeLayout)
}

func FormatStrength(first, second, third, fourth int32) string {
	return strconv.FormatInt(int64(first), 10) + "," +
		strconv.FormatInt(int64(second), 10) + "," +
		strconv.FormatInt(int64(third), 10) + "," +
		strconv.FormatInt(int64(fourth), 10)
}

func FormatWifiInfo(bssid string, level int32) string {
	return strconv.FormatInt(int64(level), 10) + "," + bssid
}

// 获取context中客户端的ip
func GetClietIP(ctx context.Context) (string, error) {
	pr, ok := peer.FromContext(ctx)
	if !ok {
		err := errors.New("[getClinetIP] invoke FromContext() failed")
		return "", err
	}
	if pr.Addr == net.Addr(nil) {
		err := errors.New("[getClientIP] peer.Addr is nil")
		return "", err
	}
	addSlice := strings.Split(pr.Addr.String(), ":")
	return addSlice[0], nil
}


func DeepCopy(dst, src interface{}) error {
	var buf bytes.Buffer
	if err := gob.NewEncoder(&buf).Encode(src); err != nil {
		return err
	}
	return gob.NewDecoder(bytes.NewBuffer(buf.Bytes())).Decode(dst)
}

func StringToINT32(s string) int32 {
	a, _ := strconv.Atoi(s)
	return int32(a)
}

