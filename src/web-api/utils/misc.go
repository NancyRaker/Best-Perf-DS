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
	"encoding/gob"
	"strconv"
	"time"
	cfgWs "web-api/config"
)

func ConvertTimeUnix(t uint64) string {
	return time.Unix(int64(t), 0).Format(cfgWs.TimeLayout)
}

func UnixStrToTimeFormat(tStr string) string {
	t, _ := strconv.ParseInt(tStr, 10, 64)
	return time.Unix(t, 0).Format(cfgWs.TimeLayout)
}

func UnixStrToWebTimeFormat(tStr string) string {
	t, _ := strconv.ParseInt(tStr, 10, 64)
	return time.Unix(t, 0).Format(cfgWs.TimeLayout)
}

func FormatStrength(first, second, third, fourth int32) string {
	return strconv.FormatInt(int64(first), 10) + "," +
		strconv.FormatInt(int64(second), 10) + "," +
		strconv.FormatInt(int64(third), 10) + "," +
		strconv.FormatInt(int64(fourth), 10)
}

func GetRedisKey(key int32) string {
	return "app:" + strconv.FormatInt(int64(key), 10) + ":stat"
}


func DeepCopy(dst, src interface{}) error {
	var buf bytes.Buffer
	if err := gob.NewEncoder(&buf).Encode(src); err != nil {
		return err
	}
	return gob.NewDecoder(bytes.NewBuffer(buf.Bytes())).Decode(dst)
}