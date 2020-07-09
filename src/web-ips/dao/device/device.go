/**
* @Author: yanKoo
* @Date: 2019/8/27 15:29
* @Description:
 */
package device

import (
	"database/sql"
	"web-ips/db"
	"web-ips/logger"
)

// 获取用户
func GetAllDevice(accountId int) ([]string, error) {
	var stmtOut *sql.Stmt
	var stmtErr error
	var res []string
	stmtOut, stmtErr = db.DBHandler.Prepare("SELECT imei FROM `user` WHERE cid = ?")
	if stmtErr != nil {
		logger.Debugf("GetAllCustomer stmt err :%+v", stmtErr)
		return nil, stmtErr
	}
	defer stmtOut.Close()


	// 查询数据
	rows, err := stmtOut.Query(accountId)
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		var (
			imei string
		)
		if err := rows.Scan(&imei); err != nil {
			return res, err
		}
		res = append(res, imei)
	}

	return res, nil
}
