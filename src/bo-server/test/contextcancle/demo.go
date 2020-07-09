/*
@Time : 2019/10/24 18:20 
@Author : yanKoo
@File : demo
@Software: GoLand
@Description:
*/
package main

import (
	"fmt"
	"time"
)

func main() {
	var c = make(chan struct{})
	go func() {
		time.Sleep(time.Second * 10)
		close(c)
	}()

	t := time.NewTicker(time.Second * 2)
tag:
	for {
		select {
		case <-c:
			fmt.Println("channel is close")
			 break tag
		case <-t.C:
			fmt.Println("Ticker is run")
		}
	}

	fmt.Println("done")
}
