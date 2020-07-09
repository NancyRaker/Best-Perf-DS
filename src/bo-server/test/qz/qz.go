/*
@Time : 2019/11/1 15:48 
@Author : yanKoo
@File : qz
@Software: GoLand
@Description:
*/
package main

import "fmt"

func main() {
	fmt.Println(appendByte(nil))
}

func appendByte(in []byte) []byte {
	var hash = []byte{'2', '6', 'a'}
	in = append(in, hash...)
	return in
}


/*
export GOROOT=/usr/local/go
export GOPATH=~/go_source/golib:~/go_source/goproject
export GOBIN=~/go_source/gobin
export PATH=$PATH:$GOROOT/bin:$GOBIN


*/