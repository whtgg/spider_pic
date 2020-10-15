package utils

import (
	"github.com/go-redis/redis"
	"log"
	"sync"
)
var (
	Redis RedisClient
	one sync.Once
)
func InitRedis(addr,password string) {
	one.Do(func() {
		rdb := redis.NewClient(&redis.Options{
			Addr:addr,
			Password:password,
			DB:0,
		})
		pong,_:= rdb.Ping().Result()
		if len(pong) == 0 {
			log.Printf("redis fail")
		}
		Redis  = RedisClient{Cmd: rdb}
	})

}