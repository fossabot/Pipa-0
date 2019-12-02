package redis

import (
	"github.com/garyburd/redigo/redis"
	"pipa/helper"
	"time"
)

var (
	Pool      *redis.Pool
	redisConn redis.Conn
)

func Initialize() {

	options := []redis.DialOption{
		redis.DialConnectTimeout(time.Duration(helper.CONFIG.RedisConnectTimeout) * time.Second),
		redis.DialReadTimeout(time.Duration(helper.CONFIG.RedisReadTimeout) * time.Second),
		redis.DialWriteTimeout(time.Duration(helper.CONFIG.RedisWriteTimeout) * time.Second),
	}

	if helper.CONFIG.RedisPassword != "" {
		options = append(options, redis.DialPassword(helper.CONFIG.RedisPassword))
	}

	Pool = &redis.Pool{
		MaxIdle:     helper.CONFIG.RedisPoolMaxIdle,
		IdleTimeout: time.Duration(helper.CONFIG.RedisPoolIdleTimeout) * time.Second,
		Dial: func() (redis.Conn, error) {
			c, err := redis.Dial("tcp", helper.CONFIG.RedisAddress, options...)
			if err != nil {
				return nil, err
			}
			return c, err
		},
		TestOnBorrow: func(c redis.Conn, t time.Time) error {
			_, err := c.Do("PING")
			return err
		},
	}

	redisConn = Pool.Get()
}

func Close() {
	err := redisConn.Close()
	if err != nil {
		helper.Logger.Info("Cannot close redis connection:", err)
	}
	err = Pool.Close()
	if err != nil {
		helper.Logger.Info("Cannot close redis pool:", err)
	}
}

func Strings() ([]string, error) {
	return redis.Strings(redisConn.Do("BLPOP", "taskQueue", 0))
}