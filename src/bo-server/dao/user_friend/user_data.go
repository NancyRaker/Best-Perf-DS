/**
 * Copyrights (c) 2019. All rights reserved.
 * User data type for the module
 * Author: tesion
 * Date: March 25th 2019
 */
package user_friend

type FriendRecord struct {
	UID  uint64 `json:"uid"`
	Name string `json:"name"`
	IMEI string `json:"imei"`
}

type UserFriendList struct {
	UID        uint64          `json:"uid"`
	FriendList []*FriendRecord `json:"friend_list"`
}
