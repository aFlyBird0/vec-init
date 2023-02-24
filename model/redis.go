package model

import (
	"context"

	"github.com/go-redis/redis/v8"
)

func SetRedis(rdb *redis.Client, key string, value any) error {
	ctx := context.Background()
	err := rdb.Set(ctx, key, value, 0).Err()

	return err
}

func GetRedis(rdb *redis.Client, key string) (any, bool, error) {
	ctx := context.Background()
	val, err := rdb.Get(ctx, key).Result()
	if err == nil {
		return val, true, nil
	}
	if err == redis.Nil {
		return nil, false, nil
	}
	return nil, false, err
}
