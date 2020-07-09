package user

import (
	"bo-server/model"
	"strconv"
	"testing"
)

func testAddDevice(t *testing.T) {
	base := "12345678541"
	for i := int64(1350); i < 1450; i++ {
		imei := base + strconv.FormatInt(i, 10)
		d := &model.User{
			Id:        int(i),
			IMei:      imei,
			UserType:  1,
			PassWord:  "123456",
			UserName:  string([]byte(imei)[9:len(imei)]),
			NickName:  ("开发" + strconv.FormatInt(i, 10) + "号"),
			AccountId: 1,
		}

		if err := AddUser(d); err != nil {
			t.Errorf("Error of AddDevice %v", err)
		}
	}
}

func testSelectUserByKey(t *testing.T) {
	if res, err := SelectUserByKey(1); err != nil {
		t.Logf("Test select user by key error: %v", err)
	} else {
		t.Log(res)
	}
}

func TestQueryDeviceOrAccountByAccountId(t *testing.T) {
	if res, err := QueryDeviceOrAccountByAccountId(90); err != nil {
		t.Logf("Test select user by key error: %v", err)
	} else {
		t.Log(res)
	}
}
//func testAddUserCache(t *testing.T) {
//	if err := AddUserForSingleGroupCache(1000, 101, cache.GetRedisClient()); err != nil {
//		t.Logf("Add user error: %v", err)
//	} else {
//	}
//}
//
//func testAddUserAddUserDataInCache(t *testing.T) {
//	if err := AddUserDataInCache(&pb.Member{
//		Id:          333,
//		IMei:        "123456789111111",
//		UserName:    "margie",
//		Online:      1,
//		LockGroupId: 9999,
//	}, cache.GetRedisClient()); err != nil {
//		t.Logf("testAddUserAddUserDataInCache error: %v", err)
//	} else {
//	}
//}
//
//func testUpdateLockGroupIdInCache(t *testing.T) {
//	if err := UpdateLockGroupIdInCache(&pb.SetLockGroupIdReq{UId: 333, GId: 9000}, cache.GetRedisClient()); err != nil {
//		t.Log("TestUpdateLockGroupIdInCache", err)
//	}
//}
//
//func testGetUserInfoFromCache(t *testing.T) {
//	if _, err := GetUserStatusFromCache(333, cache.GetRedisClient());err != nil {
//		t.Log("TestGetUserInfoFromCache error: ", err)
//	}
//}
//
//func TestGetGroupMemData(t *testing.T) {
//	res, err := GetGroupMemDataFromCache(240, cache.GetRedisClient())
//	if err != nil {
//		t.Logf("%v", err)
//	} else {
//		t.Logf("%v", res)
//	}
//}
//
//
//func testGetGroupList(t *testing.T) {
//	res, err := GetGroupList(int32(333), cache.GetRedisClient())
//	if err != nil {
//		t.Logf("Get grouplist has error: %v", err)
//	} else {
//		t.Logf("res:%v", res)
//	}
//}
