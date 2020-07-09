/**
* @Author: yanKoo
* @Date: 2019/3/11 11:16
* @Description:
 */
package customer

import (
	"database/sql"
	"strconv"
	"time"
	cfgWs "web-api/config"
	"web-api/engine/db"
	"web-api/logger"
	"web-api/model"
)

// 增加用户
func AddAccount(a *model.CreateAccount) (int, error) {
	//tx, err := db.DBHandler.Begin()
	//if err != nil {
	//	log.Log.Debug("事物开启失败")
	//}

	ctime := time.Now().Format(cfgWs.TimeLayout)
	u := &model.User{
		IMei:       "1",
		UserName:   a.Username,
		NickName:   a.NickName,
		PassWord:   a.Pwd,
		UserType:   a.RoleId,
		ParentId:   strconv.FormatInt(int64(a.Pid), 10),
		AccountId:  0,
		LLTime:     ctime,
		CreateTime: ctime,
	}

	stmtQuery := "INSERT INTO user (imei, name, passwd, cid, pid, nick_name, user_type, last_login_time, create_time)	VALUES (?, ?, ?,?, ?, ?, ?, ?, ?)"
	userRes, err := db.DBHandler.Exec(stmtQuery, u.IMei, u.UserName, u.PassWord, u.AccountId, u.ParentId, u.NickName, u.UserType, ctime, ctime)
	if err != nil {
		return -1, err
	}

	uid, err := userRes.LastInsertId()
	if err != nil {
		logger.Error("get insert AddUser Fail")
		return -1, err
	}

	// customer
	_, err = db.DBHandler.Exec("INSERT INTO customer (uid, pid, email, phone, address, remark, create_time) VALUES (?, ?, ?, ?, ?, ?, ?)",
		uid, a.Pid, a.Email, a.Phone, a.Address, a.Remark, ctime)
	if err != nil {
		return -1, err
	}
	//var affUser, affCus int64
	//
	//if userRes != nil {
	//	affUser, _ = userRes.RowsAffected()
	//}
	//if cusRes != nil {
	//	affCus, _ = cusRes.RowsAffected()
	//}
	//log.Log.Debug("create account rollback or commit start")
	//if affUser == 1 && affCus == 1 {
	//	// 提交事务
	//	if err := tx.Commit(); err != nil {
	//		return -1, err
	//	}
	//	log.Log.Debug("create account commit")
	//} else {
	//	// 回滚
	//	if err := tx.Rollback(); err != nil {
	//		return -1, err
	//	}
	//	log.Log.Debug("create account rollback")
	//}
	return int(uid), nil
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
			logger.Debugln("Statement close fail")
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

// 删除用户
func DeleteAccount(loginName string, pwd string) error {
	stmtDel, err := db.DBHandler.Prepare("DELETE FROM user WHERE user_name = ? AND password = ?")
	if err != nil {
		logger.Debugf("DeleteAccount error: %s", err)
		return err
	}
	defer func() {
		if err := stmtDel.Close(); err != nil {
			logger.Debugln("Statement close fail")
		}
	}()

	if _, err := stmtDel.Exec(loginName, pwd); err != nil {
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
			logger.Debugln("Statement close fail")
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
		logger.Debugln("no rows")
		return nil, nil
	}

	// 赋值给返回的结构体
	//log.Log.Debug("get account : ", id, "  ", username, " ", privilegeId, " ", pwd, " ", cTime)
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
	//tx, err := db.DBHandler.Begin()
	//if err != nil {
	//	log.Log.Debug("事务开启失败")
	//	return err
	//}

	userUpdStmt := "UPDATE `user` SET nick_name = ?, change_time = ? WHERE id = ?"
	cusUpdStmt := "UPDATE customer SET remark = ?, address = ?, email = ?, phone = ?, contact = ?, change_time = ? WHERE uid = ?"

	t := time.Now()
	ctime := t.Format(cfgWs.TimeLayout)

	userRes, err := db.DBHandler.Exec(userUpdStmt, a.NickName, ctime, a.Id)
	if err != nil {
		logger.Debugln("update user error : ", err)
		return err
	}
	//var userAff, cusAff int64

	if userRes != nil {
		_, err = userRes.RowsAffected()
		if err != nil {
			logger.Debugln("update user RowsAffected error : ", err)
			return err
		}
	}
	cusRes, err := db.DBHandler.Exec(cusUpdStmt,
		a.Remark,
		a.Address,
		a.Email,
		a.Phone,
		a.Contact,
		ctime, a.Id)
	if err != nil {
		logger.Debugln("update customer error : ", err)
		return err
	}

	if cusRes != nil {
		/*cusAff*/ _, err = cusRes.RowsAffected()
		if err != nil {
			logger.Debugln("update customer RowsAffected error : ", err)
			return err
		}
	}

	//log.Log.Debug(userAff, "cusAff", cusAff)
	//if userAff == 1 && cusAff == 1 {
	//	return tx.Commit()
	//} else {
	//	log.Log.Debug("rollback")
	//	return tx.Rollback()
	//}

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

// 查找下级目录
func SelectChildByPId(pId int) ([]*model.Account, error) {
	stmtOut, err := db.DBHandler.Prepare("SELECT id, `name` FROM `user` WHERE pid = ? AND user_type > 1") // 1是普通用户
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := stmtOut.Close(); err != nil {
			logger.Debugln("statement close fail.")
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

//  获取下级账户id
func SelectJuniorAccount(uId int32) ([]int32, error) {
	stmtOut, err := db.DBHandler.Prepare("SELECT customer.`uid`, customer.`pid` FROM `customer` WHERE `customer`.`pid` >=  ?") // 1是普通用户
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := stmtOut.Close(); err != nil {
			logger.Debugln("statement close fail.")
		}
	}()

	rows, err := stmtOut.Query(uId)
	if err != nil {
		return nil, err
	}

	var idMap = make(map[int32]int32)
	var list =make([][]int32, 0)
	var res []int32

	for rows.Next() {
		var uId, pId int32
		if err := rows.Scan(&uId, &pId); err != nil {
			return res, err
		}
		if _, ok := idMap[pId]; ok {
			index := idMap[pId]
			list[index] = append(list[index], uId)
		} else {
			index := len(list)
			idMap[pId] = int32(index)
			list = append(list, []int32{uId})
		}
	}

	// 根据map和list处理层级关系
	var queue = []int32{uId}
	res = append(res, uId)
	for len(queue) != 0 {
		uId := queue[0]
		queue = queue[1:]
		if key, ok := idMap[uId]; ok {
			l := list[key]
			res = append(res, list[key]...)
			queue = append(queue, l...)
		}
	}
	return res, nil
}
