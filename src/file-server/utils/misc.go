/*
@Time : 2019/5/9 20:57
@Author : yanKoo
@File : misc
@Software: GoLand
@Description:
*/
package utils

import (
	"crypto/md5"
	"encoding/hex"
)

func GetMd5(fileSource []byte) string {
	if fileSource == nil || len(fileSource) == 0 {
		return ""
	}
	md5h := md5.New()
	md5h.Write(fileSource)
	return hex.EncodeToString(md5h.Sum([]byte("")))
}
