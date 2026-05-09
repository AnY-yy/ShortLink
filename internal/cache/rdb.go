package cache

import (
	"context"
	"shortURL/internal/model"
	"time"

	"github.com/redis/go-redis/v9"
)

type Cache interface {
	CreateURL(ctx context.Context, req *model.CreateURLRequest) error
}

type RDBCache struct {
	rdb *redis.Client
}

func NewCache(rdb *redis.Client) Cache {
	return &RDBCache{
		rdb: rdb,
	}
}

func (r *RDBCache) CreateURL(ctx context.Context, req *model.CreateURLRequest) error {
	// 写入string中
	// 键: longurl 值: shorturl
	err := r.rdb.Set(ctx, req.LongURL, req.SelfShortCode, time.Hour*time.Duration(*req.ExpireTime)).Err()
	if err != nil {
		return err
	}
	return nil
}
