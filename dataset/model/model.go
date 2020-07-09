/*
@Time : 2019/12/4 13:33 
@Author : yanKoo
@File : model
@Software: GoLand
@Description:
*/
package model

type CompanyListNode struct {
	Overseas    string
	Domestic    string
	Citizenship string
	AckDate     string
}

type WorkTaskNode struct {
	Index int
	CompanyListNode *CompanyListNode
}
