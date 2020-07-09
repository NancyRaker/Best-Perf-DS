/*
@Time : 2019/9/18 15:27
@Author : yanKoo
@File : download
@Software: GoLand
@Description:
*/
package main

import "fmt"

func main() {
	v := []int{1, 2, 3}
	for i := range v {
		v = append(v, i)
		fmt.Println(v)
	}
}









