/**
 * Copyrights 2019. All rights reserved.
 * Author: tesion
 * Date: March 22th 2019
 */

package user_friend

import (
	"database/sql"
	"errors"
	"fmt"
	pb "web-api/proto/talk_cloud"
	"web-api/logger"
)

func CheckFriendExist(uid, fuid int32, db *sql.DB) (bool, error) {
	if db == nil {
		return false, fmt.Errorf("db is nil")
	}

	sql := fmt.Sprintf("SELECT COUNT(*) FROM user_friend WHERE (uid=%[1]d AND friend_uid=%[2]d AND stat!=1) AND (uid=%[2]d AND friend_uid=%[1]d AND stat!=1)", uid, fuid)

	rows, err := db.Query(sql)
	if err != nil {
		return false, fmt.Errorf("query(%s) error(%s)", sql, err)
	}
	defer rows.Close()

	for rows.Next() {
		var cnt int
		err = rows.Scan(&cnt)
		if err != nil {
			return false, fmt.Errorf("query scan error(%s)", err)
		}

		if 0 >= cnt {
			return false, nil
		}
	}

	return true, nil
}

// 添加好友
func AddFriend(uid, fuid int32, db *sql.DB) (bool, error) {
	//ret, err := CheckFriendExist(uid, fuid, db)
	//if err != nil {
	//	return false, err
	//}
	//
	//if ret {
	//	return false, fmt.Errorf("user friend exists")
	//}

	sql := fmt.Sprintf("INSERT INTO user_friend(uid, friend_uid) VALUES(%d, %d)", uid, fuid)
	_, err := db.Query(sql)
	if err != nil {
		return false, fmt.Errorf("query(%s), error(%s)", sql, err)
	}

	return true, nil
}

// 删除好友
func RemoveFriend(uid, fuid int32, db *sql.DB) (bool, error) {
	if db == nil {
		return false, fmt.Errorf("db is nil")
	}

	sql := fmt.Sprintf("DELETE FROM user_friend WHERE uid=%d AND friend_uid=%d", uid, fuid)
	_, err := db.Query(sql)

	if err != nil {
		return false, err
	}

	return true, nil
}

// 获取好友请求列表
func GetFriendReqList(uid int32, db *sql.DB) (*pb.FriendsRsp, *map[int32]string, error) {
	if db == nil {
		return nil, nil, errors.New("db is nil")
	}

	sql := fmt.Sprintf("SELECT d.id, d.name, d.imei FROM user AS d RIGHT JOIN user_friend AS u ON u.friend_uid=d.id WHERE u.uid=%d", uid)
	rows, err := db.Query(sql)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()

	friends := &pb.FriendsRsp{Uid: uid, FriendList: nil}
	friendsMap := map[int32]string{}
	for rows.Next() {
		friend := new(pb.FriendRecord)
		err = rows.Scan(&friend.Uid, &friend.Name, &friend.Imei)
		if err != nil {
			return nil, nil, err
		}
		friendsMap[friend.Uid] = friend.Name
		friends.FriendList = append(friends.FriendList, friend)
	}

	return friends, &friendsMap, nil
}

// 查找用户
func SearchUserByName(_ int32, target string, db *sql.DB) (*pb.UserSearchRsp, error) {
	if db == nil {
		return nil, errors.New("db is nil")
	}

	rows, err := db.Query("SELECT id, name FROM user WHERE name LIKE ?", "%"+target+"%")
	if err != nil {
		logger.Debugf("query user error(%s)", err)
		return nil, err
	}
	logger.Debugln(rows)
	defer rows.Close()

	users := &pb.UserSearchRsp{UserList: nil}
	for rows.Next() {
		user := new(pb.UserRecord)

		err = rows.Scan(&user.Uid, &user.Name)
		if err != nil {
			logger.Debugf("scan rows error: %s", err)
			return nil, err
		}
		user.IsFriend = false
		users.UserList = append(users.UserList, user)
	}

	return users, nil
}

func SearchFriendByUserId(uid uint64, db *sql.DB) (*pb.UserSearchRsp, error) {
	if db == nil {
		return nil, fmt.Errorf("db is nil")
	}

	rows, err := db.Query("SELECT id, name FROM user_friend WHERE uid = ?", uid)
	if err != nil {
		logger.Debugf("query user error(%s)", err)
		return nil, err
	}
	logger.Debugln(rows)
	defer rows.Close()

	users := &pb.UserSearchRsp{UserList: nil}

	for rows.Next() {
		user := new(pb.UserRecord)

		err = rows.Scan(&user.Uid, &user.Name)
		if err != nil {
			logger.Debugf("scan rows error: %s", err)
			return nil, err
		}
		users.UserList = append(users.UserList, user)
	}

	return users, nil
}
