/*
@Time : 2019/8/22 18:01 
@Author : yanKoo
@File : account
@Software: GoLand
@Description:
*/
package service

import (
	"context"
	"errors"
	"github.com/gomodule/redigo/redis"
	"strconv"
	tc "web-api/dao/customer" // table customer
	"web-api/dao/pub"
	"web-api/engine/cache"
	"web-api/logger"
	"web-api/model"
)

// 获取账户等级树
func getAccountTree(input interface{}, errOut chan interface{}) interface{} {
	if input == nil {
		errOut <- errors.New("getAccountTree input is nil")
		return nil
	}
	account := input.(*model.Account)
	resElem, err := tc.SelectChildByPId(account.Id)
	if err != nil {
		logger.Debugf("db error : %s", err)
		errOut <- model.ErrorDBError
	}

	cList := make([]*model.AccountClass, 0)
	for i := 0; i < len(resElem); i++ {
		child, err := tc.GetAccount((*resElem[i]).Id)
		if err != nil {
			logger.Debugf("db error : %s", err)
			errOut <- model.ErrorDBError
		}

		cList = append(cList, &model.AccountClass{
			Id:              child.Id,
			AccountName:     child.Username,
			AccountNickName: child.NickName,
		})
	}

	//fmt.Printf("end get device tree info: %d\n", time.Now().UnixNano())
	return &model.AccountClass{
		Id:              account.Id,
		AccountName:     account.Username,
		AccountNickName: account.NickName,
		Children:        cList,
	}
}

// 获取单级账户树
func GetAccountTree(ctx context.Context, account interface{}, accountTreeChan, errOut chan interface{}) {
	go ProcessAccountInfo(ctx, account, accountTreeChan, errOut, getAccountTree)
}

// 获取
func GetJuniorAccount(accountId int32)([]*model.JuniorAccount, error){
	var (
		uIds []int32
		err error
		rd = cache.GetRedisClient()
		res  []*model.JuniorAccount
	)
	if rd == nil {
		return nil, nil
	}
	defer rd.Close()


	if uIds, err = tc.SelectJuniorAccount(accountId); err != nil {
		return nil, err
	}

	// 3.0 批量查询uid的信息和在线状态
	for _, uId := range uIds {
		_ = rd.Send("HMGET", pub.MakeUserDataKey(int32(uId)),
			pub.USER_Id, pub.NICK_NAME,	pub.USER_NAME)
	}
	_ = rd.Flush()
	for _, uId := range uIds {
		source, err := redis.Strings(rd.Receive())
		user := &model.JuniorAccount{}
		if source != nil {
			id, _ := strconv.Atoi(source[0])
			user.Id = id
			user.AccountNickName = source[1]
			user.AccountName = source[2]
			res = append(res, user)
		}
		logger.Debugf("from redis # %d users source: %+v with err: %+v", uId, source, err)
	}

	return res, nil
}

