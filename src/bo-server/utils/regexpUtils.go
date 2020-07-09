/*
@Time : 2019/3/30 11:41
@Author : yanKoo
@File : regexpUtils
@Software: GoLand
@Description:
*/
package utils

import (
	"fmt"
	"regexp"
	"bo-server/logger"
	"strconv"
)

func CheckId(id int) bool {
	if id == 0 {
		return false
	}
	reg, err := regexp.Compile("^[0-9]{1,10}$")
	if err != nil {
		fmt.Println(err)
	}
	return reg.MatchString(strconv.Itoa(id))
}

func CheckIMei(imei string) bool {
	reg, err := regexp.Compile("^[0-9]{15}$")
	if err != nil {
		logger.Debugln(err)
	}
	return reg.MatchString(imei)
}