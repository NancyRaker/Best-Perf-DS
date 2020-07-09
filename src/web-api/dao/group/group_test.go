package group

import (
	"web-api/engine/cache"
	"web-api/engine/db"
	"web-api/model"
	"strings"
	"testing"
)

func testDeleteGroup(t *testing.T) {
	if err := DeleteGroup(&model.GroupInfo{
		Id: 101,
	}); err != nil {
		t.Logf("Test Delete Group error : %v", err)
	} else {
		t.Logf("sccess delete")
	}
}

func TestCreateGroup(t *testing.T) {
	s := strings.Split("123,abc,157,1235", ",")
	t.Log(s)
	/*nums := []int{1278, 1279, 1280}
	ds := make([]interface{}, 0)
	ds = append(ds, &model.User{
		Id: 333,
		UserName:  "123466",
		PassWord:  "123456",
		AccountId: 0, // TODO 默认给谁  普通用户默认是0
		IMei:      "123467894512365",
		UserType: 1,
	})

	g := &model.GroupInfo{GroupName: "天津组2", AccountId: 334, Id: 334} // 用户id
	gl := &model.GroupListNode{DeviceIds: nums, GroupInfo: g, DeviceInfo:ds}
	if _, err := CreateGroup(gl, 0); err != nil {   // 1是管理员创建， 0是普通用户创建
		t.Errorf("create group test error: %v", err)
	}*/
}

//func testGetGroupList(t *testing.T) {
//	if res, err := GetGroupListFromDB(uint64(333),db.DBHandler); err != nil {
//		t.Log("Get GroupListNode Error : ", err )
//	} else {
//		t.Log(*res)
//	}
//}

func testAddGroupCache(t *testing.T) {
	nums := []int{9000, 10000, 11000, 12000, 13000}
	ds := make([]interface{}, 0)

	for i := 1; i < 101; i++ {
		ds = append(ds, map[string]interface{}{
			"id":       i,
			"UserName": "123466",
			"PassWord": "123456",
			//AccountId"": 0, // TODO 默认给谁  普通用户默认是0
			//IMei:      "123467894512365",
			//UserType: 1,
		})
	}

	g := &model.GroupInfo{GroupName: "天津组2", AccountId: 2, Id: 101} // 用户id
	gl := &model.GroupListNode{DeviceIds: nums, GroupInfo: g, DeviceInfo: ds}
	if err := AddGroupAndUserInCache(gl, cache.GetRedisClient()); err != nil {
		t.Logf("Add GroupCache error: %v", err)
	} else {
		//t.Logf("res:%v", res)
	}
}

func testGetGroupList(t *testing.T) {
	//res, err := user.GetGroupListFromDB(uint64(1000), cache.GetRedisClient())
	//if err != nil && err != cache.NofindInCacheError {
	//	t.Logf("Add GroupCache error: %v", err)
	//} else if err == cache.NofindInCacheError {
	//	t.Logf("Can't find in cache")
	//} else {
	//	t.Logf("find group list:%-v", res)
	//}
}

func testSearchGroup(t *testing.T) {
	res, err := SearchGroup("雷坤", db.DBHandler)
	if err != nil {
		t.Logf("Add GroupCache error: %v", err)
	} else {
		t.Logf("find group list:%-v", res)
	}

}
