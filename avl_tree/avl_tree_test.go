/*
@Time : 2019/11/13 20:53
@Author : yanKoo
@File : avl_tree_test
@Software: GoLand
@Description: avl tree test
*/
package avl_tree

import (
	"fmt"
	"sort"
	"strconv"
	"testing"
)

func TestNewAVLTree(t *testing.T) {
	avl := NewAVLTree()
	arr := []int{12, 45, 78, 1, 2, 6, 43, 871, 44}
	sort.Ints(arr)
	for i := range arr {
		fmt.Print(arr[i], " ")
		avl.Add(arr[i], "value is "+strconv.Itoa(arr[i]))
	}
	fmt.Println()

	t.Logf("bst? %v\n", avl.IsBST())
	t.Logf("avl? %v", avl.IsBalance())
	t.Logf("avl len %v", avl.Size())
	avl.Remove(44)
	t.Logf("bst? %v\n", avl.IsBST())
	t.Logf("avl? %v", avl.IsBalance())
	t.Logf("avl len %v", avl.Size())
	t.Logf("get %d: %v", 43, avl.Get(43))

}
