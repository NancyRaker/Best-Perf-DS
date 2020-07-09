/*
@Time : 2019/3/30 11:41
@Author : yanKoo
@File : regexpUtils
@Software: GoLand
@Description:
*/
package utils

import (
	"regexp"
	"strconv"
)

var (
	phoneReg    *regexp.Regexp
	emailReg    *regexp.Regexp
	pwdReg      *regexp.Regexp
	nickNameReg *regexp.Regexp
	userNameReg *regexp.Regexp
	idReg       *regexp.Regexp
	iMeiReg     *regexp.Regexp
	bssIdReg    *regexp.Regexp
	lonReg      *regexp.Regexp
	latReg      *regexp.Regexp
)

func init() {
	phoneReg = regexp.MustCompile("^((13[0-9])|(15[^4,\\D])|(18[0,5-9]))\\d{8}$")
	emailReg = regexp.MustCompile("\\w+(\\.\\w)*@\\w+(\\.\\w{2,3}){1,3}")
	pwdReg = regexp.MustCompile("^([a-zA-Z0-9]|[_]){6,20}$")
	phoneReg = regexp.MustCompile("^((13[0-9])|(15[^4,\\D])|(18[0,5-9]))\\d{8}$")
	nickNameReg = regexp.MustCompile("^([a-zA-Z0-9]|[\u4e00-\u9fa5]){1}([a-zA-Z0-9]|[_]|[\u4e00-\u9fa5]){0,20}$")
	userNameReg = regexp.MustCompile("^[a-zA-Z]{1}([a-zA-Z0-9]|[_]){4,19}$")
	idReg = regexp.MustCompile("^[0-9]{1,10}$")
	iMeiReg = regexp.MustCompile("^[0-9]{15}$")
	bssIdReg = regexp.MustCompile("^([0-9A-Fa-f]{2}[:-]){5}([0-9A-Fa-f]{2})$")
	lonReg = regexp.MustCompile("^(\\+|-)?(?:180(?:(?:\\.0{1,6})?)|(?:[0-9]|[1-9][0-9]|1[0-7][0-9])(?:(?:\\.[0-9]{1,6})?))$")
	latReg = regexp.MustCompile("^(\\+|-)?(?:90(?:(?:\\.0{1,6})?)|(?:[0-9]|[1-8][0-9])(?:(?:\\.[0-9]{1,6})?))$")
}

// 校验手机号码
func CheckPhone(phone string) bool {
	return phoneReg.MatchString(phone)
}

// 校验邮箱
func CheckEmail(email string) bool {
	return emailReg.MatchString(email)
}

// 校验密码：只能输入6-20个字母、数字、下划线
func CheckPwd(pwd string) bool {
	return pwdReg.MatchString(pwd)
}

// 校验昵称：只能输入1-20个以字母或者数字开头、可以含中文、下划线的字串。
func CheckNickName(name string) bool {
	return nickNameReg.MatchString(name)
}

// 校验用户名：只能输入5-20个包含字母、数字或下划线的字串。
func CheckUserName(name string) bool {
	return userNameReg.MatchString(name)
}

func CheckId(id int) bool {
	if id == 0 {
		return false
	}
	return idReg.MatchString(strconv.Itoa(id))
}

// 校验IMei
func CheckIMei(imei string) bool {
	return iMeiReg.MatchString(imei)
}

// 校验bssId（mac地址）
func CheckBssId(mac string) bool {
	return bssIdReg.MatchString(mac)
}

// 校验经度
func CheckLongitude(lon string) bool {
	return lonReg.MatchString(lon)
}

// 校验纬度
func CheckLatitude(lat string) bool {
	return latReg.MatchString(lat)
}









