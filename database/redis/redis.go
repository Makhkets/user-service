package redis

import (
	"context"
	"fmt"
	"github.com/go-redis/redis/v8"
)

func isnit() {
	// создаем экземпляр клиента Redis
	rdb := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // если нет пароля, оставьте пустым
		DB:       0,  // номер БД по умолчанию
	})

	// проверяем, что соединение установлено
	pong, err := rdb.Ping(context.Background()).Result()
	fmt.Println(pong, err)

	// записываем значение в Redis
	err = rdb.Set(context.Background(), "key", "value", 0).Err()
	if err != nil {
		panic(err)
	}

	// получаем значение из Redis по ключу
	val, err := rdb.Get(context.Background(), "key").Result()
	if err != nil {
		panic(err)
	}
	fmt.Println("key:", val)
}
