/*
@Time : 2019/9/5 15:26 
@Author : yanKoo
@File : grpc_manager
@Software: GoLand
@Description:
*/
package main

import (
	"fmt"
	"time"
	"web-api/engine/grpc_pool"
)

func main()  {

	go func() {
		for i := 0; i < 5; i++ {
			if cli := grpc_pool.GRPCManager.GetGRPCConnClientById(7); cli == nil {
				fmt.Println("go NULL")
			} else {
				conn := cli.ClientConn
				fmt.Println(i, " go ", conn.Target())
				cli.Close()
			}
		}
		fmt.Println("go end")
	}()

	go func() {
		for i := 0; i < 30; i++ {
			if cli := grpc_pool.GRPCManager.GetGRPCConnClientById(7); cli == nil {
				fmt.Println("NULL")
			} else {
				conn := cli.ClientConn
				fmt.Println(i, " ", conn.Target())
				defer cli.Close()
			}
		}
		fmt.Println("end")
	}()
	time.Sleep(time.Hour)
}


