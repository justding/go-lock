package redlock

import (
	"github.com/go-redis/redis"
)

// NewRedisClient return a pointer to a new redis client
func NewRedisClient(addr string) (*redis.Client, error) {
	opts, err := redis.ParseURL(addr)
	if err != nil {
		return nil, err
	}
	client := redis.NewClient(opts)
	return client, nil
}

// NewRedisClientPool returns an array with pointers to redis client
func NewRedisClientPool(addr []string) ([]*redis.Client, error) {
	pool := make([]*redis.Client, len(addr))

	for i, a := range addr {
		c, err := NewRedisClient(a)

		if err != nil {
			return nil, err
		}

		pool[i] = c
	}

	return pool, nil
}
