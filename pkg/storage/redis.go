package storage

import (
	"context"
	"fmt"

	"os"

	"github.com/joho/godotenv"
	"github.com/redis/go-redis/v9"
)

var _ = godotenv.Load()
var ctx = context.Background()

var rdb = redis.NewClient(&redis.Options{
	Addr:     os.Getenv("DB_HOST") + os.Getenv("DB_PORT"),
	Username: os.Getenv("DB_USER"),
	Password: os.Getenv("DB_PASSWORD"),
	DB:       0, // use default DB
})

func Save(key, value string) error {
	err := rdb.Set(ctx, key, value, 0).Err()
	return err
}

func Load(key string) (string, error) {
	val, err := rdb.Get(ctx, key).Result()
	fmt.Println(rdb)
	if err != nil {
		return "", err
	}
	return val, nil
}

func Keys() ([]string, error) {
	keys, err := rdb.Keys(ctx, "*").Result()
	if err != nil {
		return keys, err
	}
	return keys, nil
}
