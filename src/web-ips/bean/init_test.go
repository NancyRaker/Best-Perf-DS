/*
@Time : 2019/8/27 15:30 
@Author : yanKoo
@File : init_test
@Software: GoLand
@Description:
*/
package bean

import "testing"

func TestTrieNode_Insert(t *testing.T) {
	root := NewTrieMap()
	root.Insert("123456789654123", "dg")
	root.Insert("012124548054123", "dg")
	root.Insert("123456780004513", "micro")
	t.Log(root.Search("123456780004513"))
	t.Log(root.Search("000"))

}
