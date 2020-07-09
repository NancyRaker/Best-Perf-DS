package customer

import (
	"web-api/logger"
	"web-api/model"
	"testing"
)

func testAddAccount(t *testing.T) {
	if _, err := AddAccount(&model.CreateAccount{

		Pid:         9,
		Username:    "panda4",
		NickName:    "winna",
		Pwd:         "123456789",
		Email:       "948162@qq.com",
		Phone:       "123456789",
		Address:     "株洲",
		Remark:      "熊猫",
		PrivilegeId: 1,
		RoleId:      2,
	}); err != nil {
		t.Log("Test add account error : ", err)
	}
}

func testGetAccount(t *testing.T) {
	if res, err := GetAccount(1); err != nil {
		t.Log("Test error : ", err)
	} else {
		logger.Debugln(res)
	}
}

func testUpdateAccount(t *testing.T) {
	if err := UpdateAccount(&model.AccountUpdate{
		LoginId:  "9",
		Id:       "31",
		NickName: "ZZZZZZZZ",
		Email:    "948162@qq.com",
		Phone:    "123456789",
		Address:  "株洲",
		Remark:   "",
	}); err != nil {
		t.Log("Test add account error : ", err)
	}
}

func testSelectChildByPId(t *testing.T) {
	res, err := SelectChildByPId(32)
	if err != nil {
		t.Log(err)
	}
	t.Logf("result: %v", res)
}

func TestGetAccountTree(t *testing.T) {
	res , err := GetLowerAccount(int32(1))
	t.Log(res, "\nerr:", err)
}
