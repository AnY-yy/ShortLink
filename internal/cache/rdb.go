package cache

import "github.com/redis/go-redis/v9"

type Cache interface {
}

type RDBCache struct {
	rdb *redis.Client
}

func NewCache(rdb *redis.Client) Cache {
	return &RDBCache{
		rdb: rdb,
	}
}
