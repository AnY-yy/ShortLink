package cache

import (
	"context"
	"fmt"
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

// CreateURL 将短链与长链存入缓存中
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

// GetURL 从缓存中获取短链对应的长链
func (c *Cache) GetURL(ctx context.Context, shortURL string) (*model.RedirectURLResponse, error) {
	rep := &model.RedirectURLResponse{}
	var err error

	rep.ShortURL = shortURL

	// 获取长链接
	rep.LongURL, err = c.rdb.Get(ctx, shortURL).Result()
	if err != nil {
		return nil, fmt.Errorf("获取缓存失败: %w", err)
	}

	// 获取过期时间
	ttl, err := c.rdb.TTL(ctx, shortURL).Result()
	if err != nil {
		rep.ExpireAt = time.Time{} // 返回默认零值
	} else {
		if ttl > 0 {
			rep.ExpireAt = time.Now().Add(ttl)
		} else { // 为0时则为永不过期 返回零值
			rep.ExpireAt = time.Time{} // 返回默认零值
		}
	}
	return rep, nil
}
