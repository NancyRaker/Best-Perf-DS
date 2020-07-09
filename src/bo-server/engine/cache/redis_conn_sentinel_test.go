/*
@Time : 2019/7/2 10:27 
@Author : yanKoo
@File : redis_conn_sentinel
@Software: GoLand
@Description:
*/
package cache
import (
	"fmt"
	"strings"
	"github.com/gomodule/redigo/redis"
	"time"

	"github.com/FZambia/sentinel"
)

var RedisConnPool *redis.Pool

func InitRedisSentinelConnPool() {
	redisAddr := "192.168.1.11:26378,192.168.1.22:26378"
	redisAddrs := strings.Split(redisAddr, ",")
	masterName := "master1" // 根据redis集群具体配置设置

	sntnl := &sentinel.Sentinel{
		Addrs:      redisAddrs,
		MasterName: masterName,
		Dial: func(addr string) (redis.Conn, error) {
			timeout := 500 * time.Millisecond
			c, err := redis.DialTimeout("tcp", addr, timeout, timeout, timeout)
			if err != nil {
				return nil, err
			}
			return c, nil
		},
	}

	RedisConnPool = &redis.Pool{
		//MaxIdle:     redisCfg.MaxIdle,
		IdleTimeout: 240 * time.Second,
		Dial: func() (redis.Conn, error) {
			masterAddr, err := sntnl.MasterAddr()
			if err != nil {
				return nil, err
			}
			c, err := redis.Dial("tcp", masterAddr)
			if err != nil {
				return nil, err
			}
			return c, nil
		},
		TestOnBorrow: CheckRedisRole,
	}
}

func CheckRedisRole(c redis.Conn, t time.Time) error {
	if !sentinel.TestRole(c, "master") {
		return fmt.Errorf("Role check failed")
	} else {
		return nil
	}
}
