/*
@Time : 2019/8/19 9:33 
@Author : yanKoo
@File : hungry
@Software: GoLand
@Description:
*/
package main

import (
	_ "github.com/go-sql-driver/mysql"
)

func main() {

}
//func main() {
//	data := struct {
//		Imeis     []string `json:"imeis"`
//		AccountId int      `json:"account_id"`
//	}{
//		Imeis:     []string{"122345678945612"},
//		AccountId: 115}
//
//	jsonStr, _ := json.Marshal(data)
//	url := "http://localhost:20000/server/device/add"
//	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonStr))
//	req.Header.Set("Content-Type", "application/json")
//
//	client := &http.Client{}
//	resp, err := client.Do(req)
//	if err != nil {
//		// handle error
//	}
//	defer resp.Body.Close()
//	statuscode := resp.StatusCode
//	fmt.Println(statuscode)
//
//	//runtime.GOMAXPROCS(1)
//	//
//	//go func() {
//	//	for x := 0; x < 10; x++ {
//	//		fmt.Println(x)
//	//	}
//	//}()
//	//
//	//go func() {
//	//	for {
//	//		//runtime.Gosched()
//	//		x := 1
//	//		x += 0
//	//	}
//	//}()
//	//
//	//time.Sleep(time.Hour * 9)
//}

/**
 * Definition for a binary tree node.
 * type TreeNode struct {
 *     Val int
 *     Left *TreeNode
 *     Right *TreeNode
 * }
 */
type TreeNode struct {
	Val   int
	Left  *TreeNode
	Right *TreeNode
}

func maxLevelSum(root *TreeNode) int {
	var sum []int
	var max_value int
	var res int
	if root != nil {
		sum = append(sum, 0)
		recursion(root, &sum, &max_value, &res, 0)
	}
	return res
}

// [989,null,10250,98693,-89388,null,null,null,-32127]
func recursion(node *TreeNode, sum *[]int, max_value *int, res *int, h int) {
	(*sum)[h] += node.Val
	if (*sum)[h] > *max_value {
		*max_value = (*sum)[h]
		*res = h + 1
	}

	if node.Left == nil && node.Right == nil {
		return
	}

	*sum = append(*sum, 0)
	if node.Left != nil {
		recursion(node.Left, sum, max_value, res, h+1)
	}
	if node.Right != nil {
		recursion(node.Right, sum, max_value, res, h+1)
	}
}
