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

	HMSet(ctx context.Context, key string, values ...interface{}) *redis.BoolCmd
	HMGet(ctx context.Context, key string, fields ...string) *redis.SliceCmd

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
	// проверяем, что соединение установлено
	_, err := rdb.Ping(context.Background()).Result()
	if err != nil {
		panic(err)
	}

	return rdb
}
