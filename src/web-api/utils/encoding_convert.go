/*
@Time : 2019/4/19 14:02
@Author : yanKoo
@File : encoding_convert
@Software: GoLand
@Description:
*/
package utils

import (
	"regexp"
	"strconv"
)

func ConvertOctonaryUtf8(in string) string {
	s := []byte(in)
	reg := regexp.MustCompile(`\\[0-7]{3}`)

	out := reg.ReplaceAllFunc(s,
		func(b []byte) []byte {
			i, _ := strconv.ParseInt(string(b[1:]), 8, 0)
			return []byte{byte(i)}
		})
	return string(out)
}
