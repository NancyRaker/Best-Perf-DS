/*
@Time : 2019/11/13 14:53
@Author : yanKoo
@File : bst_test
@Software: GoLand
@Description: bst test
*/
package binary_search_tree

import (
	"fmt"
	"testing"
)

func TestNewBST(t *testing.T) {
	bst := NewBST()
	arr := []int{40, 5, 42, 6, 10, 400, 77}
	for _, e := range arr {
		bst.Add(e)
	}
	fmt.Println(bst.size, len(arr))
	fmt.Println(bst.Search(4))
	fmt.Println(bst.Search(40))
	bst.Remove(40)
	fmt.Println(bst.Search(40))
	fmt.Println(bst.size, len(arr))
	fmt.Println(bst.RemoveMax())
	fmt.Println(bst.RemoveMin())

}
