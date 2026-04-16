package redisutil

import (
	"github.com/go-redis/redis/v8"
)

func New(addr, password string, db int) *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})
}
