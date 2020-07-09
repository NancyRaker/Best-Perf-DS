/**
* @Author: yanKoo
* @Date: 2019/3/11 11:16
* @Description:
 */
package customer

import (
	"database/sql"
	"web-ips/db"
	"web-ips/logger"
	"web-ips/model"
)

// 获取用户
func GetAllCustomer() ([]*model.Customer, error) {
	var stmtOut *sql.Stmt
	var stmtErr error
	var res []*model.Customer
	stmtOut, stmtErr = db.DBHandler.Prepare("SELECT `name`, uid, server_addr FROM customer LEFT JOIN `user` ON user.`id` = customer.`uid`")
	if stmtErr != nil {
		logger.Debugf("GetAllCustomer stmt err :%+v", stmtErr)
		return nil, stmtErr
	}
	defer stmtOut.Close()

	// 查询数据
	rows, err := stmtOut.Query()
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		var (
			customer = &model.Customer{}
		)
		if err := rows.Scan(&customer.AccountName, &customer.Id, &customer.AddrCode); err != nil {
			return res, err
		}
		res = append(res, customer)
	}

	return res, nil
	// 赋值给返回的结构体
}

// 修改用户登录地址编号
func UpdateCustomer(serverCode []int, accountId []int) error {
	var stmtOut *sql.Stmt
	var stmtErr error
	stmtOut, stmtErr = db.DBHandler.Prepare("UPDATE customer SET server_addr = ? WHERE customer.`uid` = ?")
	if stmtErr != nil {
		logger.Debugf("GetAllCustomer stmt err :%+v", stmtErr)
		return nil
	}
	defer stmtOut.Close()

	// 查询数据
	for i, id := range accountId {
		_, err := stmtOut.Exec(serverCode[i], id)
		if err != nil {
			return err
		}
	}
	return nil
}
