/*
@Time : 2019/9/11 16:44 
@Author : yanKoo
@File : goruntine_lock
@Software: GoLand
@Description:
*/
package main

import (
	"context"
	"fmt"
	"math/rand"
	"sync"
)
// 消息数据来源
type DataSource interface{}

// 全局map，保存stream连接 // TODO 后期修改
type ClientContext struct {
	Dc *DataContext
	Ds DataSource
}
// 分发的任务
type Task struct {
	Receiver []int32
	SenderId int32
}

// im推送上下文
type DataContext struct {
	UId              int32
	TempId           chan int32
	Task             chan Task
	ExceptionalLogin chan int32 // 重复登录
	Ctx              context.Context
	Ctf              context.CancelFunc
}

// var Gm sync.Mutex
var GoroutineMap = GoroutineMapSt{
	ClientContexts: make(map[int32]*ClientContext),
}

type GoroutineMapSt struct {
	ClientContexts map[int32]*ClientContext
	Lock           sync.RWMutex
}

func (g GoroutineMapSt) GetContext(k int32) *DataContext {
	g.Lock.RLock()
	defer g.Lock.RUnlock()
	if cc := g.ClientContexts[k]; cc == nil {
		return nil
	}
	return g.ClientContexts[k].Dc
}

func (g GoroutineMapSt) Len() int {
	g.Lock.RLock()
	defer g.Lock.RUnlock()
	return len(g.ClientContexts)
}

func (g GoroutineMapSt) Set(k int32, cc *ClientContext) {
	g.Lock.Lock()
	defer g.Lock.Unlock()
	g.ClientContexts[k] = cc
}

// 获取stream
func (g GoroutineMapSt) GetStream(k int32) DataSource {
	//g.lock.RLock()
	//defer g.lock.RUnlock()
	if g.ClientContexts[k] == nil {
		return nil
	}
	return g.ClientContexts[k]
}

// 如果存在这个stream就替换，如果不存在就保存
func (g GoroutineMapSt) GetAndSet(k int32, cc *ClientContext) {
	g.Lock.Lock()
	defer g.Lock.Unlock()
	c := g.ClientContexts[k]
	if c != nil && c.Dc != nil && c.Dc.ExceptionalLogin != cc.Dc.ExceptionalLogin {
		fmt.Printf("this user # %d is login already", k)
		c.Dc.ExceptionalLogin <- k
		g.ClientContexts[k] = cc
	} else if c == nil {
		fmt.Printf("this user # %d first call dataPublish", k)
		g.ClientContexts[k] = cc
	}
}

// 删除这个连接
func (g GoroutineMapSt) Del(id int32) {
	g.Lock.Lock()
	defer g.Lock.Unlock()
	if cc := g.ClientContexts[id]; cc != nil {
		fmt.Printf("the user # %d will clean scheduler and dispatcher\n", id)
		cc.Dc.ExceptionalLogin <- id
	}
	delete(g.ClientContexts, id)
}


// var Gm sync.Mutex
var maps = GoroutineMapSt{
	ClientContexts: make(map[int32]*ClientContext),
}

func main() {
	for i := int32(0); i < 10; i++ {
		//maps.clientContexts[i] = &ClientContext{dataContext:&DataContext{UId:i}}
	}

	fmt.Printf("init maps:%+v", maps)

	fmt.Println("开始抢锁")

	var wg sync.WaitGroup
	for i := 100000; i > 0; i-- {
		wg.Add(1)
		go setClient(&wg)
		wg.Add(1)
		go getClient(&wg)
	}
	wg.Wait()
}

func getClient(wg *sync.WaitGroup)  {
	defer wg.Done()
	i := rand.Intn(3)
	if value := maps.GetStream(int32(i)); value != nil {
		fmt.Println(i, "get")
	}else {
		fmt.Println(i, " can't get lock")
	}
}
func setClient(wg *sync.WaitGroup) {
	defer wg.Done()
	i := rand.Intn(3)
	maps.GetAndSet(int32(i),  &ClientContext{Dc:&DataContext{UId:int32(i)}})
}
