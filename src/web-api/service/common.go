/*
@Time : 2019/8/23 17:10 
@Author : yanKoo
@File : common
@Software: GoLand
@Description:
*/
package service

import "context"

type processFunc func(accountId interface{}, errOut chan interface{}) interface{}

func ProcessAccountInfo(ctx context.Context, account interface{}, out, errOut chan interface{}, f processFunc) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
			out <- f(account, errOut)
		}
	}
}
