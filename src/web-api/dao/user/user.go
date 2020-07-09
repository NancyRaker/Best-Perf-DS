package user

import (
	"database/sql"
	"errors"
	"github.com/smartwalle/dbs"
	"time"
	cfgWs "web-api/config"
	"web-api/engine/db"
	"web-api/logger"
	"web-api/model"
	pb "web-api/proto/talk_cloud"
)

// 增加设备
func AddUser(u *model.User, in ...interface{}) error {
	var stmtInsB = dbs.NewInsertBuilder()

	stmtInsB.Table("user")
	//stmtInsB.SET("id",u.Id)
	stmtInsB.SET("imei", u.IMei)
	stmtInsB.SET("name", u.UserName)
	stmtInsB.SET("passwd", u.PassWord)
	stmtInsB.SET("cid", u.AccountId)
	//stmtInsB.SET("pid", u.ParentId)
	stmtInsB.SET("nick_name", u.NickName) // 注册的时候默认把username当做昵称
	stmtInsB.SET("user_type", 1)
	t := time.Now()
	ctime := t.Format(cfgWs.TimeLayout)
	stmtInsB.SET("last_login_time", ctime)
	stmtInsB.SET("create_time", ctime)

	if _, err := stmtInsB.Exec(db.DBHandler); err != nil {
		return err
	}

	return nil
}

// 用过用户名查重，用在app GRpc注册
func GetUserByName(userName string) (int, error) {
	stmtOut, err := db.DBHandler.Prepare("SELECT count(id) FROM user WHERE name = ?")
	if err != nil {
		logger.Debugf("DB error :%s", err)
		return -1, err
	}
	defer func() {
		if err := stmtOut.Close(); err != nil {
			logger.Debugln("Statement close fail")
		}
	}()

	var res int
	if err := stmtOut.QueryRow(userName).Scan(&res); err != nil {
		logger.Debugf("err: %s", err)
		return -1, err
	}

	return res, nil
}

// 通过关键词查找用户名
func SelectUserByKey(key interface{}) (*model.User, error) {
	var stmtOut *sql.Stmt
	var err error
	switch t := key.(type) {
	case int:
		stmtOut, err = db.DBHandler.Prepare("SELECT id, name, nick_name, passwd, imei, user_type, pid, cid, lock_gid, create_time, last_login_time, change_time FROM `user` WHERE id = ?")
	case string:
		stmtOut, err = db.DBHandler.Prepare("SELECT id, name, nick_name, passwd, imei, user_type, pid, cid, lock_gid, create_time, last_login_time, change_time  FROM `user` WHERE name = ?")
	default:
		_ = t
		return nil, err
	}
	if err != nil {
		logger.Debugf("SelectUserByKey db.DBHandler.Prepare fail with error: %s", err)
		return nil, err
	}
	defer func() {
		if err := stmtOut.Close(); err != nil {
			logger.Debugln("Statement close fail")
		}
	}()

	var (
		id, userType, cId, lockGId                                    int
		pId, userName, nickName, pwd, iMei, cTime, llTime, changeTime string
	)
	err = stmtOut.QueryRow(key).Scan(&id, &userName, &nickName, &pwd, &iMei, &userType, &pId, &cId, &lockGId, &cTime, &llTime, &changeTime)
	if err != nil {
		return nil, err
	}

	res := &model.User{
		Id:          id,
		IMei:        iMei,
		UserName:    userName,
		PassWord:    pwd,
		NickName:    nickName,
		UserType:    userType,
		ParentId:    pId,
		AccountId:   cId,
		LockGroupId: lockGId,
		CreateTime:  cTime,
		LLTime:      llTime,
		ChangeTime:  changeTime,
	}

	return res, nil
}

// 查找设备
func SelectUserByAccountId(aid int) ([]*model.Device, error) {
	var stmtOut *sql.Stmt
	var err error
	stmtOut, err = db.DBHandler.Prepare(`SELECT id, imei, name, nick_name, passwd, user_type, start_log, cid, create_time,
       												last_login_time, change_time, d_type, active_time, sale_time
	                                        	FROM user WHERE cid = ?`)
	if err != nil {
		logger.Debugf("%s", err)
		return nil, err
	}
	defer func() {
		if err := stmtOut.Close(); err != nil {
			logger.Debugln("Statement close fail")
		}
	}()

	var res []*model.Device

	rows, err := stmtOut.Query(aid)
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		var device = model.Device{}
		if err := rows.Scan(&device.Id, &device.IMei, &device.UserName, &device.NickName,
			&device.PassWord, &device.UserType, &device.StartLog, &device.AccountId, &device.CreateTime, &device.LLTime, &device.ChangeTime,
			&device.DeviceType, &device.ActiveTime, &device.SaleTime); err != nil {
			return res, err
		}
		res = append(res, &device)
	}

	return res, nil
}

// 设置用户锁定默认组
func SetLockGroupId(req *pb.SetLockGroupIdReq, db *sql.DB) error {
	if db == nil {
		return errors.New("set Lock group Id error, db is nil")
	}

	updStmt, err := db.Prepare("UPDATE`user` SET lock_gid = ? WHERE id = ?")
	if err != nil {
		return errors.New("set Lock group Id error, updStmt error " + err.Error())
	}

	_, err = updStmt.Exec(req.GId, req.UId)
	if err != nil {
		return errors.New("set Lock group Id error, updStmt.Exec error " + err.Error())
	}

	return nil
}

// 找出所有的用户ID
func SelectAllUserId() ([]int32, error) {
	var stmtOut *sql.Stmt
	var err error
	stmtOut, err = db.DBHandler.Prepare("SELECT id FROM `user`")
	if err != nil {
		logger.Debugf("%s", err)
		return nil, err
	}
	defer func() {
		if err := stmtOut.Close(); err != nil {
			logger.Debugln("Statement close fail")
		}
	}()

	var res []int32

	rows, err := stmtOut.Query()
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		var id int32
		if err := rows.Scan(&id); err != nil {
			return res, err
		}
		res = append(res, id)
	}

	return res, nil
}
