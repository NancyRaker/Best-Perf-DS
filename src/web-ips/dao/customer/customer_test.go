/*
@Time : 2019/9/21 17:32 
@Author : yanKoo
@File : customer_test
@Software: GoLand
@Description:
*/
package customer

import "testing"

func TestUpdateCustomer(t *testing.T) {
	err := UpdateCustomer([]int{1},[]int{3})
	t.Log(err)
}
