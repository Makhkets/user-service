package rdb

import (
	"Makhkets/internal/configs"
	"context"
	"github.com/go-redis/redis/v8"
	"time"
)

type Client interface {
	Set(ctx context.Context, key string, value interface{}, expiration time.Duration) *redis.StatusCmd
	Get(ctx context.Context, key string) *redis.StringCmd
	Del(ctx context.Context, keys ...string) *redis.IntCmd

	SAdd(ctx context.Context, key string, members ...interface{}) *redis.IntCmd
	LRem(ctx context.Context, key string, count int64, value interface{}) *redis.IntCmd
	LRange(ctx context.Context, key string, start, stop int64) *redis.StringSliceCmd

	SMembers(ctx context.Context, key string) *redis.StringSliceCmd

	HMSet(ctx context.Context, key string, values ...interface{}) *redis.BoolCmd
	HMGet(ctx context.Context, key string, fields ...string) *redis.SliceCmd
	HGetAll(ctx context.Context, key string) *redis.StringStringMapCmd

	Expire(ctx context.Context, key string, expiration time.Duration) *redis.BoolCmd
	Exists(ctx context.Context, keys ...string) *redis.IntCmd

	Append(ctx context.Context, key, value string) *redis.IntCmd
	Ping(ctx context.Context) *redis.StatusCmd
}

func InitRedis() Client {
	cfg := configs.GetConfig()

	// создаем экземпляр клиента Redis
	rdb := redis.NewClient(&redis.Options{
		Addr:     cfg.Redis.Address,
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.Db,
	})

	_, err := rdb.Ping(context.Background()).Result()
	if err != nil {
		panic(err)
	}

	return rdb
}
