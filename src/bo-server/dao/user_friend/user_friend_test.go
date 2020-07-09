/*
@Time : 2019/4/3 18:02
@Author : yanKoo
@File : user_friend_test
@Software: GoLand
@Description:
*/
package user_friend

import (
	"bo-server/engine/db"
	"testing"
)

func testSearchUserByName(t *testing.T) {
	res, err := SearchUserByName(333, "44", db.DBHandler)
	if err != nil {
		t.Logf("error: %v", err)
	} else {
		t.Logf("res:%+v", res)
	}
}

func TestAddFriend(t *testing.T) {
	if res, err := AddFriend(333, 336, db.DBHandler); err != nil {
		t.Logf("Add friend error: %v", err)
	} else {
		t.Log(res)
	}
}
