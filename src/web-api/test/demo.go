///*
//@Time : 2019/7/10 15:28
//@Author : yanKoo
//@File : demo
//@Software: GoLand
//@Description:
//*/
package main

import (
	"fmt"
	"web-api/test/jose_loop"
)

func generator(n int) <-chan int {
	outCh := make(chan int)
	go func() {
		for i := 0; i < n; i++ {
			outCh <- i
		}
	}()
	return outCh
}
func main() {
	jr := jose_loop.NewJoseLoop(7, 1)
	//jr.printElement()
	fmt.Println(jr.StartKill())
}
func splitGroupMember(olders []int64, newers []int32) (addMem []int32, revMem []int32) {
	var (
		olderMap = make(map[int32]bool)
		newerMap = make(map[int32]bool)
	)

	fmt.Printf("older: %+v ", olders)
	fmt.Printf("newers: %+v\n", newers)

	for _, v := range newers {
		newerMap[int32(v)] = true
	}

	for _, v := range olders {
		id := int32(v)
		if !newerMap[id] {
			revMem = append(revMem, id)
		}
		olderMap[id] = true
	}

	for id := range newerMap {
		if !olderMap[id] {
			addMem = append(addMem, id)
		}
	}

	fmt.Printf("add:%+v ", addMem)
	fmt.Printf("rev:%+v\n", revMem)
	return
}


/*
sudo sh ./run.sh -e canal.auto.scan=false \
		  -e canal.destinations=test \
		  -e canal.instance.master.address=127.0.0.1:3506  \
		  -e canal.instance.dbUsername=canal  \
		  -e canal.instance.dbPassword=canal  \
		  -e canal.instance.connectionCharset=UTF-8 \
		  -e canal.instance.tsdb.enable=true \
		  -e canal.instance.gtidon=false  \
*/