package cache

import (
	"context"
	"shortURL/internal/model"
	"time"

	"github.com/redis/go-redis/v9"
)

type Cache interface {
	CreateURL(ctx context.Context, req *model.CreateURLRequest) error
	RedirectURL(ctx context.Context, shortURL string) (*model.RedirectURLResponse, error)
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
	// 键: shorturl 值: longurl
	err := r.rdb.Set(ctx, req.SelfShortCode, req.LongURL, time.Hour*time.Duration(*req.ExpireTime)).Err()
	if err != nil {
		return err
	}
	return nil
}

// RedirectURL
// 从缓存中获取短链对应的长链
func (r *RDBCache) RedirectURL(ctx context.Context, shortURL string) (*model.RedirectURLResponse, error) {
	rep := &model.RedirectURLResponse{}
	var err error

	rep.ShortURL = shortURL

	// 获取长链
	rep.LongURL, err = r.rdb.Get(ctx, shortURL).Result()
	if err != nil {
		return nil, err
	}

	// 获取过期时间
	et, err := r.rdb.TTL(ctx, shortURL).Result()
	if err != nil {
		rep.ExpireAt = time.Time{} // 返回默认空值
	} else {
		if et > 0 {
			rep.ExpireAt = time.Now().Add(et)
		} else {
			rep.ExpireAt = time.Time{}
		}
	}

	return rep, nil
}
