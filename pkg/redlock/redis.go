package redlock

import (
	"github.com/go-redis/redis"
)

// NewRedisClient return a new redis client
func NewRedisClient(addr string) (*redis.Client, error) {
	opts, err := redis.ParseURL(addr)
	if err != nil {
		return nil, err
	}
	client := redis.NewClient(opts)
	return client, nil
}
