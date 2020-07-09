package customer

import (
	"testing"
)

//func testUpdateAccount(t *testing.T) {
//	if err := UpdateAccount(&model.AccountUpdate{
//		LoginId:  "9",
//		Id:       "31",
//		NickName: "ZZZZZZZZ",
//		Email:    "948162@qq.com",
//		Phone:    "123456789",
//		Address:  "株洲",
//		Remark:   "",
//	}); err != nil {
//		t.Log("Test add account error : ", err)
//	}
//}

func TestSelectChildByPId(t *testing.T) {
	res, err := SelectChildByPId(32)
	if err != nil {
		t.Log(err)
	}
	t.Logf("result: %v", res)
}
