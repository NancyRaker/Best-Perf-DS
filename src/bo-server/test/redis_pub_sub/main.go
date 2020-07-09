/*
@Time : 2019/9/24 14:07 
@Author : yanKoo
@File : sub_pub_demo
@Software: GoLand
@Description:
*/
package main

import (
	"bo-server/engine/cache"
	"context"
	"errors"
	"fmt"
	"github.com/gomodule/redigo/redis"
	"time"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	err := listenPubSubChannels(ctx,
		func(channel string, message []byte) error {
			fmt.Printf("channel: %s, message: %s\n", channel, message)

			// For the purpose of this example, cancel the listener's context
			// after receiving last message sent by publish().
			if string(message) == "goodbye" {
				cancel()
			}
			return nil
		},
		"__keyevent@0__:expired")

	if err != nil {
		fmt.Println(err)
		return
	}
}


func listenPubSubChannels(ctx context.Context, onMessage func(channel string, data []byte) error, channels ...string) error {

	const healthCheckPeriod = time.Second * 8
	var err error
	c := cache.GetRedisClient()
	if c == nil {
		err = errors.New("redis nil")
		return err
	}
	defer c.Close()

	psc := redis.PubSubConn{Conn: c}

	if err := psc.Subscribe(redis.Args{}.AddFlat(channels)...); err != nil {
		return err
	}

	done := make(chan error, 1)
	// Start a goroutine to receive notifications from the server.
	go func() {
		for {
			switch n := psc.Receive().(type) {
			case error:
				done <- n
				return
			case redis.Message:
				if err := onMessage(n.Channel, n.Data); err != nil {
					done <- err
					return
				}
			}
		}
	}()

	ticker := time.NewTicker(healthCheckPeriod)
	defer ticker.Stop()
loop:
	for err == nil {
		select {
		case <-ticker.C:
			// Send ping to test health of connection and server. If
			// corresponding pong is not received, then receive on the
			// connection will timeout and the receive goroutine will exit.
			if err = psc.Ping(""); err != nil {
				break loop
			}
		case <-ctx.Done():
			break loop
		case err := <-done:
			// Return error from the receive goroutine.
			return err
		}
	}

	// Signal the receiving goroutine to exit by unsubscribing from all channels.
	psc.Unsubscribe()

	// Wait for goroutine to complete.
	return <-done
}