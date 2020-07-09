package cache

import (
	"context"
	"errors"
	"fmt"
	"github.com/gomodule/redigo/redis"
	"strings"
	"time"
	"file-server/conf"
	"file-server/logger"
)

var RedisPool *redis.Pool
var NofindInCacheError = errors.New("no find in Cache Error")

func CheckRedisRole(c redis.Conn, t time.Time) error {
	if !TestRole(c, "master") {
		return fmt.Errorf("Role check failed")
	} else {
		return nil
	}
}

func NewRedisPool(redisCfg *conf.RedisConfig) (*redis.Pool, error) {
	if redisCfg == nil {
		return nil, fmt.Errorf("config is nil")
	}
	pool := &redis.Pool{}

	if redisCfg.SentinelAddr == "" {
		redisUrl := fmt.Sprintf("%s:%d", redisCfg.Host, redisCfg.Port)
		pool = &redis.Pool{
			MaxIdle:     redisCfg.MaxIdle,
			MaxActive:   redisCfg.MaxActive,
			IdleTimeout: time.Duration(redisCfg.IdleTimeout) * time.Second,
			Wait:        true,
			Dial: func() (redis.Conn, error) {
				con, err := redis.Dial("tcp", redisUrl,
					redis.DialPassword(redisCfg.Password),
					redis.DialDatabase(redisCfg.DB),
					redis.DialConnectTimeout(time.Duration(redisCfg.Timeout)*time.Second),
					redis.DialReadTimeout(time.Duration(redisCfg.Timeout)*time.Second),
					redis.DialWriteTimeout(time.Duration(redisCfg.Timeout)*time.Second))

				if err != nil {
					return nil, err
				}

				return con, nil
			}, // Dial end
			TestOnBorrow: CheckRedisRole,
		}
	} else {
		sntnl := &Sentinel{
			Addrs:      strings.Split(redisCfg.SentinelAddr, ","),
			MasterName: "mymaster",
			Dial: func(addr string) (redis.Conn, error) {
				timeout := 500 * time.Millisecond
				c, err := redis.Dial("tcp", addr,
					redis.DialConnectTimeout(timeout),
					redis.DialReadTimeout(timeout),
					redis.DialWriteTimeout(timeout))
				if err != nil {
					return nil, err
				}
				return c, nil
			},
		}

		pool = &redis.Pool{
			MaxIdle:     redisCfg.MaxIdle,
			MaxActive:   redisCfg.MaxActive,
			IdleTimeout: time.Duration(redisCfg.IdleTimeout) * time.Second,
			Wait:        true,
			Dial: func() (redis.Conn, error) {
				masterAddr, err := sntnl.MasterAddr()
				if err != nil {
					return nil, err
				}
				con, err := redis.Dial("tcp", masterAddr,
					redis.DialPassword(redisCfg.Password),
					redis.DialDatabase(redisCfg.DB),
					redis.DialConnectTimeout(time.Duration(redisCfg.Timeout)*time.Second),
					redis.DialReadTimeout(time.Duration(redisCfg.Timeout)*time.Second),
					redis.DialWriteTimeout(time.Duration(redisCfg.Timeout)*time.Second))

				if err != nil {
					return nil, err
				}

				return con, nil
			}, // Dial end
			TestOnBorrow: CheckRedisRole,
		}
	}

	return pool, nil
}

//var count = 0
func GetRedisClient() redis.Conn {
	if RedisPool == nil {
		fmt.Println("pool nil")
		return nil
	}

	var (
		ctx    context.Context
		cancel context.CancelFunc
	)

	ctx, cancel = context.WithTimeout(context.Background(), time.Second*3) // 3s取不到连接就返回
	defer cancel()

	conn, err := RedisPool.GetContext(ctx)
	if err != nil {
		//count++
		fmt.Printf("can't get conn from redis pool # %s\n", err.Error())
		return nil
	}

	return conn
}

func init() {
	var err error
	cfg := conf.NewRedisConfig()
	if err := cfg.LoadConfig("redis", conf.DEFAULT_CONFIG); err != nil {
		RedisPool = nil
		logger.Debugf("Init NewRedisConfig LoadConfig fail with error: %+v", err)
		return
	}

	RedisPool, err = NewRedisPool(cfg)
	if err != nil {
		logger.Debugf("Init NewRedisPool fail with error:%+v", err)
	}
}
