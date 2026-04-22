package cache

import (
	"shortURL/internal/bootstrap"

	"github.com/redis/go-redis/v9"
)

type Cache struct {
	rdb *redis.Client
}

// NewCache 创建缓存实例
func NewCache() *Cache {
	return &Cache{
		rdb: bootstrap.Application.RDB,
	}
}
