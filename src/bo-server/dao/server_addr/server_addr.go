/*
@Time : 2019/7/1 11:42 
@Author : yanKoo
@File : server_addr
@Software: GoLand
@Description:
*/
package server_addr

import (
	"database/sql"
	"bo-server/engine/db"
	"bo-server/logger"
	"bo-server/model"
	"strings"
)

const (
	DISABLE_SERVER = 1
	ENABLE_SERVER  = 2

	JANUS_SERVER = 1
	// ...
)

func GetServerAddr(serverType int) []*model.ServerInfo {
	var (
		res     = make([]*model.ServerInfo, 0)
		stmtOut *sql.Stmt
		err     error
		rows    *sql.Rows
	)
	if db.DBHandler = db.DBHandler; db.DBHandler == nil {
		logger.Debugln("GetServerAddr db.DBHandler is nil")
		return res
	}

	if stmtOut, err = db.DBHandler.Prepare("SELECT s_addr, s_loc FROM server_addr WHERE s_type = ? AND s_status = ?"); err != nil {
		logger.Debugf("GetServerAddr stmtOut error :", err)
		return res
	}

	if rows, err = stmtOut.Query(serverType, ENABLE_SERVER); err != nil {
		logger.Debugf("GetServerAddr stmtOut.Query error :", err)
		return res
	}

	for rows.Next() {
		si := &model.ServerInfo{}
		var addr string
		if err = rows.Scan(addr, si.Location); err != nil {
			return make([]*model.ServerInfo, 0)
		}
		addrs := strings.Split(addr,":")

		if addrs != nil && len(addrs) == 2{
			si.Ip = addrs[0]
			si.Port = addrs[1]
		} else {
			return make([]*model.ServerInfo, 0)
		}
		si.Type =  serverType
		res = append(res, si)
	}
	return res
}
