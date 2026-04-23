package cache

import (
	"context"
	"shortURL/internal/bootstrap"
	"shortURL/internal/model"
	"time"

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

func (c *Cache) CreateURL(ctx context.Context, req *model.CreateURLRequest) error {
	// 短链设置为key 长链设置为value
	key := req.SelfShortUrl
	value := req.LongURL
	// 过期时间
	expireTime := time.Hour * time.Duration(*req.ExpireTime) // 将过期时间转换为时间单位 int -> time.Duration
	// 写入缓存
	err := c.rdb.Set(ctx, key, value, expireTime).Err()
	if err != nil {
		return err
	}
	return nil
}
