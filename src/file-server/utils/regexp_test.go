/*
@Time : 2019/4/1 13:40
@Author : yanKoo
@File : regexp_test
@Software: GoLand
@Description:
*/
package utils

import (
	"strconv"
	"testing"
)

//
//func testCheckPwd(t *testing.T) {
//	t.Log(CheckPwd("gagdfh"))
//}
//
//func testCheckNickName(t *testing.T) {
//	t.Log(CheckNickName("中"))
//}
//
//func testCheckUserName(t *testing.T) {
//	t.Log(CheckUserName("safs"))
//}
//
func testCheckId(t *testing.T) {
	t.Log(CheckBssId("14:DD:A9:4F:03:8F"))
	t.Log(strconv.FormatFloat(25.100423444, 'f', 6, 64))
	t.Log(strconv.FormatFloat(12, 'f', 6, 64))
	t.Log(CheckLongitude("25.100423444"))
	t.Log(CheckLongitude(strconv.FormatFloat(-12, 'f', 6, 64)))
}
//// 获取文件大小的接口
//type Size interface {
//	Size() int64
//}
//func testGetFileType(t *testing.T) {
//	//f, err := os.Open("C:\\Users\\Administrator\\Desktop\\api.html")
//	f, err := os.Open("C:\\Users\\Administrator\\Desktop\\335_1556418847_1556418846861_voice_1556417943423.mp3")
//	//f, err := os.Open("C:\\Users\\Public\\Music\\Sample Music\\Kalimba.mp3")
//	if err != nil {
//		t.Logf("open error: %v", err)
//	}
//
//	fSrc, err := ioutil.ReadAll(f)
//	t.Log(GetFileType(fSrc[:10]))
//	t.Logf("file len:")
//}

func testGettime(t *testing.T) {
	//t.Log(UnixStrToTimeFormat("1557475724"))
}

func estCheckIMei(t *testing.T) {
	t.Log("----------------------->",CheckIMei("123456489123456"))
}