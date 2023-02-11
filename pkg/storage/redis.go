package storage

import (
	"context"

	"github.com/redis/go-redis/v9"
)

var ctx = context.Background()

var rdb = redis.NewClient(&redis.Options{
	Addr:     "127.0.0.1:6370",
	Username: "pepe",
	Password: "pepe", // no password set
	DB:       0,      // use default DB
})

func Save(key, value string) error {
	err := rdb.Set(ctx, key, value, 0).Err()
	return err
}

func Load(key string) (string, error) {
	val, err := rdb.Get(ctx, key).Result()
	if err != nil {
		return "", err
	}
	return val, nil
}
