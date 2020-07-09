/**
* @Author: yanKoo
* @Date: 2019/3/11 11:16
* @Description:
 */
package customer

import (
	pb "bo-server/api/proto"
	cfgComm "bo-server/conf"
	"bo-server/engine/db"
	"bo-server/logger"
	"bo-server/model"
	"database/sql"
	"strconv"
	"time"
)

// 增加用户
func AddAccount(a *pb.CreateAccountReq) (*model.User, error) {
	ctime := time.Now().Format(cfgComm.TimeLayout)
	u := &model.User{
		IMei:       "1",
		UserName:   a.Username,
		NickName:   a.NickName,
		PassWord:   a.Pwd,
		UserType:   int(a.RoleId),
		ParentId:   strconv.FormatInt(int64(a.Pid), 10),
		AccountId:  0,
		LLTime:     ctime,
		CreateTime: ctime,
	}

	stmtQuery := "INSERT INTO user (imei, name, passwd, cid, pid, nick_name, user_type, last_login_time, create_time)	VALUES (?, ?, ?,?, ?, ?, ?, ?, ?)"
	userRes, err := db.DBHandler.Exec(stmtQuery, u.IMei, u.UserName, u.PassWord, u.AccountId, u.ParentId, u.NickName, u.UserType, ctime, ctime)
	if err != nil {
		return nil, err
	}

	uid, err := userRes.LastInsertId()
	if err != nil {
		logger.Error("get insert AddUser Fail")
		return nil, err
	}

	// customer
	_, err = db.DBHandler.Exec("INSERT INTO customer (uid, pid, email, phone, address, remark, create_time) VALUES (?, ?, ?, ?, ?, ?, ?)",
		uid, a.Pid, a.Email, a.Phone, a.Address, a.Remark, ctime)
	if err != nil {
		return nil, err
	}

	u.Id = int(uid)
	return u, nil
}

// 查找下级目录
func SelectChildByPId(pId int) ([]*model.Account, error) {
	stmtOut, err := db.DBHandler.Prepare("SELECT id, `name` FROM `user` WHERE pid = ?")
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := stmtOut.Close(); err != nil {
			logger.Error("statement close fail.")
		}
	}()

	rows, err := stmtOut.Query(pId)
	if err != nil {
		return nil, err
	}

	var res []*model.Account

	for rows.Next() {
		var id int
		var userName string
		if err := rows.Scan(&id, &userName); err != nil {
			return res, err
		}

		acc := &model.Account{Id: id, Pid: pId, Username: userName}
		res = append(res, acc)
	}

	return res, nil
}

// 删除用户
func DeleteAccount(aid int32) error {
	stmtDel, err := db.DBHandler.Prepare("DELETE customer, `user` FROM `user` LEFT JOIN customer ON `user`.id = customer.`uid`WHERE `user`.`id` = ?")
	if err != nil {
		logger.Debugf("DeleteAccount error: %s", err)
		return err
	}
	defer stmtDel.Close()

	if _, err := stmtDel.Exec(aid); err != nil {
		return err
	}
	return nil
}

// 通过账户名获取账户数 注册查重
func GetAccountByName(userName string) (int, error) {
	stmtOut, err := db.DBHandler.Prepare("SELECT count(id) FROM user WHERE name = ?")
	if err != nil {
		logger.Debugf("DB error :%s", err)
		return -1, err
	}

	var res int
	if err := stmtOut.QueryRow(userName).Scan(&res); err != nil {
		logger.Debugf("err: %s", err)
		return -1, err
	}

	defer func() {
		if err := stmtOut.Close(); err != nil {
			logger.Error("Statement close fail")
		}
	}()
	return res, nil
}

// 获取用户
func GetAccount(key interface{}) (*model.Account, error) {
	var stmtOut *sql.Stmt
	var stmtErr error
	switch t := key.(type) {
	case int:
		stmtOut, stmtErr = db.DBHandler.Prepare(`SELECT  u.id, u.pid, u.name, u.nick_name, u.passwd, u.user_type, u.last_login_time, u.create_time, u.change_time, email, phone, remark, address, contact 
													FROM user AS u LEFT JOIN customer AS c ON u.id = c.uid WHERE u.id = ?`)
	case string:
		stmtOut, stmtErr = db.DBHandler.Prepare(`SELECT  u.id, u.pid, u.name, u.nick_name, u.passwd, u.user_type, u.last_login_time, u.create_time, u.change_time, email, phone, remark, address, contact 
													FROM user AS u LEFT JOIN customer AS c ON u.id = c.uid WHERE u.name = ?`)
	default:
		_ = t
	}

	if stmtErr != nil {
		logger.Debugf("%s", stmtErr)
		return nil, stmtErr
	}

	defer func() {
		if err := stmtOut.Close(); err != nil {
			logger.Error("Statement close fail")
		}
	}()

	var (
		id, pid                                int
		username, nickname, pwd                string
		email, phone, remark, contact, address sql.NullString
		privilegeId                            int
		roleId                                 int
		stat                                   string
		llTime, cTime, changeTime              string
	)
	// 查询数据
	err := stmtOut.QueryRow(key).
		Scan(&id, &pid, &username, &nickname, &pwd, &roleId, &llTime, &cTime, &changeTime, &email, &phone, &remark, &address, &contact)

	if err != nil {
		logger.Debugf("err: %s", err)
		return nil, err
	}

	if err == sql.ErrNoRows {
		logger.Error("no rows")
		return nil, nil
	}

	// 赋值给返回的结构体
	//log.log.Error("get account : ", id, "  ", username, " ", privilegeId, " ", pwd, " ", cTime)
	res := &model.Account{
		Id:          id,
		Pid:         pid,
		Username:    username,
		NickName:    nickname,
		Pwd:         pwd,
		Email:       email,
		PrivilegeId: privilegeId,
		RoleId:      roleId,
		State:       stat,
		LlTime:      llTime,
		Contact:     contact,
		ChangeTime:  changeTime,
		CTime:       cTime,
		Phone:       phone,
		Address:     address,
		Remark:      remark,
	}
	return res, nil
}

// 更新用户
func UpdateAccount(a *model.AccountUpdate) error {
	tx, err := db.DBHandler.Begin()
	if err != nil {
		logger.Error("事务开启失败")
		return err
	}

	userUpdStmt := "UPDATE `user` SET nick_name = ?, change_time = ? WHERE id = ?"
	cusUpdStmt := "UPDATE customer SET remark = ?, address = ?, email = ?, phone = ?, contact = ?, change_time = ? WHERE uid = ?"

	t := time.Now()
	ctime := t.Format(cfgComm.TimeLayout)

	userRes, err := tx.Exec(userUpdStmt, a.NickName, ctime, a.Id)
	if err != nil {
		logger.Error("update user error : ", err)
		return err
	}
	var userAff, cusAff int64

	if userRes != nil {
		userAff, err = userRes.RowsAffected()
		if err != nil {
			logger.Error("update user RowsAffected error : ", err)
			return err
		}
	}
	cusRes, err := tx.Exec(cusUpdStmt, a.Remark, a.Address, a.Email, a.Phone, a.Contact, ctime, a.Id)
	if err != nil {
		logger.Error("update customer error : ", err)
		return err
	}

	if cusRes != nil {
		cusAff, err = cusRes.RowsAffected()
		if err != nil {
			logger.Error("update customer RowsAffected error : ", err)
			return err
		}
	}

	logger.Error(userAff, "cusAff", cusAff)
	if userAff == 1 && cusAff == 1 {
		return tx.Commit()
	} else {
		logger.Error("rollback")
		return tx.Rollback()
	}

	return nil
}

// 更新密码
func UpdateAccountPwd(pwd string, id int) error {
	stmtUpd, err := db.DBHandler.Prepare("UPDATE user SET passwd = ? WHERE id = ?")
	if err != nil {
		return err
	}

	if _, err := stmtUpd.Exec(pwd, id); err != nil {
		return err
	}

	return nil
}

// 获取用户的密码
func GetAccountPwdByKey(key interface{}) (string, error) {
	var stmtOut *sql.Stmt
	var err error
	switch t := key.(type) {
	case int:
		stmtOut, err = db.DBHandler.Prepare("SELECT passwd FROM user WHERE id = ?")
	case string:
		stmtOut, err = db.DBHandler.Prepare("SELECT passwd FROM user WHERE user_name = ?")
	default:
		_ = t
		return "", err
	}
	if err != nil {
		logger.Debugf("%s", err)
		return "", err
	}
	defer func() {
		if err := stmtOut.Close(); err != nil {
			logger.Error("Statement close fail")
		}
	}()

	var pwd string
	err = stmtOut.QueryRow(key).Scan(&pwd)
	if err != nil && err != sql.ErrNoRows {
		return "", err
	}

	if err == sql.ErrNoRows {
		return "", err
	}

	return pwd, nil
}
