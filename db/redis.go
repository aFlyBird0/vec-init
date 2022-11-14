package db

import (
	"fmt"

	"github.com/go-redis/redis/v8"

	"vec/config"
)

func initRedis() *redis.Client {
	rdb := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", config.Get().RedisConfig.Host, config.Get().RedisConfig.Port),
		Password: config.Get().RedisConfig.Password,
		DB:       config.Get().RedisConfig.Database,
	})

	return rdb
}
