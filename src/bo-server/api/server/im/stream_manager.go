/*
@Time : 2019/7/30 14:42 
@Author : yanKoo
@File : conn_center
@Software: GoLand
@Description:管理grpc的stream连接和负责监听的goroutine，producer可能还是会出现goroutine泄漏
*/
package im
//
//import (
//	"bo-server/engine/cache"
//	"bo-server/logger"
//	"context"
//	"errors"
//	"github.com/gomodule/redigo/redis"
//	"strconv"
//	"strings"
//	"sync"
//	"time"
//)
//
//// 全局连接管理 Gm sync.Mutex
//var GoroutineMap = goroutineMapSt{}
//
//func init()  {
//	GoroutineMap = goroutineMapSt{
//		clientContexts: make(map[int32]*clientContext),
//	}
//}
//
//// 全局map，保存stream连接 // TODO 后期修改
//type clientContext struct {
//	context *DataContext
//	source  DataSource
//}
//
//func newClientContext(dc *DataContext, ds DataSource) *clientContext {
//	return &clientContext{context: dc, source: ds}
//}
//
//type goroutineMapSt struct {
//	clientContexts map[int32]*clientContext
//	lock           sync.RWMutex
//}
//
//func (g *goroutineMapSt) GetContext(k int32) *DataContext {
//	g.lock.Lock()
//	defer g.lock.Unlock()
//
//	cc, ok := g.clientContexts[k]
//	if !ok {
//		return nil
//	}
//	return cc.context
//}
//
//// 获取stream
//func (g *goroutineMapSt) GetStream(k int32) DataSource {
//	g.lock.Lock()
//	defer g.lock.Unlock()
//
//	cc, ok := g.clientContexts[k]
//	if !ok {
//		return nil
//	}
//	return cc.source
//}
//
//// 如果存在这个stream就替换，如果不存在就保存
//func (g *goroutineMapSt) GetAndSet(k int32, cc *clientContext) {
//	g.lock.Lock()
//	defer g.lock.Unlock()
//
//	c := g.clientContexts[k]
//	if c != nil && c.context != nil && c.context.ExceptionalLogin != cc.context.ExceptionalLogin {
//		logger.Debugf("this user # %d is login already", k)
//		c.context.ExceptionalLogin <- k
//		g.clientContexts[k] = cc
//	} else if c == nil {
//		logger.Debugf("this user # %d first call dataPublish", k)
//		g.clientContexts[k] = cc
//	}
//}
//
//func (g *goroutineMapSt) Len() int {
//	g.lock.Lock()
//	defer g.lock.Unlock()
//	return len(g.clientContexts)
//}
//
//func (g *goroutineMapSt) Set(k int32, cc *clientContext) {
//	g.lock.Lock()
//	defer g.lock.Unlock()
//	g.clientContexts[k] = cc
//}
//
//// 删除这个连接
//func (g *goroutineMapSt) Del(id int32) {
//	g.lock.Lock()
//	defer g.lock.Unlock()
//	if cc := g.clientContexts[id]; cc != nil {
//		logger.Debugf("the user # %d will clean scheduler and dispatcher", id)
//		cc.context.ExceptionalLogin <- id
//	}
//	delete(g.clientContexts, id)
//}
//
//// 更新stream map 使用redis的pub和sub监听过期key的机制
//func syncStreamWithRedis() {
//	// 订阅redis过期的key
//	var expireKeyC = make(chan *redis.Message, 10)
//	go func() {
//		err := listenPubSubChannels(context.Background(), expireKeyC, "__keyevent@0__:expired")
//		if err != nil {
//			logger.Errorf("stream_manager syncStreamWithRedis has err: %+v", err)
//			return
//		}
//	}()
//	var keyQ = make([]*redis.Message, 0)
//	var dispatcherChan = createExMsgDispatcher()
//	for {
//		var activeChan chan *redis.Message
//		var activeMsg *redis.Message
//		if len(keyQ) > 0 {
//			activeChan = dispatcherChan
//			activeMsg = keyQ[0]
//		}
//
//		select {
//		case key := <- expireKeyC:
//			keyQ = append(keyQ, key)
//		case activeChan <- activeMsg:
//			keyQ = keyQ[1:]
//		}
//	}
//}
//
//func createExMsgDispatcher() chan *redis.Message {
//	tc := make(chan *redis.Message)
//	go exMsgHandler(tc)
//	return tc
//}
//
//// 处理过期的key（ usr:62:stat ）
//func exMsgHandler(mc chan *redis.Message) {
//	for msg := range mc{
//		// 解析messages
//		logger.Debugf("stream_manager parse key: %s", string(msg.Data))
//		temp := strings.Split(string(msg.Data), ":")
//
//		if temp == nil || len(temp) != 3 {
//			continue
//		}
//		id, _ := strconv.Atoi(temp[1])
//		logger.Debugf("stream_manager Will del uid: %d", id)
//
//		GoroutineMap.Del(int32(id))
//
//		go NotifyToOther(GlobalTaskQueue.Tasks, int32(id), LOGOUT_NOTIFY_MSG)
//	}
//}
//
//
//func listenPubSubChannels(ctx context.Context, expireKeyC chan *redis.Message, channels ...string) error {
//	const healthCheckPeriod = time.Second * 8
//	var err error
//	c := cache.GetRedisClient()
//	if c == nil {
//		err = errors.New("redis nil")
//		return err
//	}
//	defer c.Close()
//
//	psc := redis.PubSubConn{Conn: c}
//
//	if err := psc.Subscribe(redis.Args{}.AddFlat(channels)...); err != nil {
//		return err
//	}
//
//	done := make(chan error, 1)
//	// Start a goroutine to receive notifications from the server.
//	go func() {
//		for {
//			switch n := psc.Receive().(type) {
//			case error:
//				done <- n
//				return
//			case redis.Message:
//				expireKeyC <- &n
//			}
//		}
//	}()
//
//	ticker := time.NewTicker(healthCheckPeriod)
//	defer ticker.Stop()
//loop:
//	for err == nil {
//		select {
//		case <-ticker.C:
//			// Send ping to test health of connection and server. If
//			// corresponding pong is not received, then receive on the
//			// connection will timeout and the receive goroutine will exit.
//			if err = psc.Ping(""); err != nil {
//				break loop
//			}
//		case <-ctx.Done():
//			break loop
//		case err := <-done:
//			// Return error from the receive goroutine.
//			return err
//		}
//	}
//
//	// Signal the receiving goroutine to exit by unsubscribing from all channels.
//	psc.Unsubscribe()
//
//	// Wait for goroutine to complete.
//	return <-done
//}
