package rdb

import (
	"context"
	"fmt"
	"shortURL/config"

	"github.com/redis/go-redis/v9"
)

// NewRDB 初始化Redis连接
func NewRDB(cfgProvider config.ConfigProvider) (*redis.Client, error) {
	cfg := cfgProvider.GetConfig()

	rdb := redis.NewClient(&redis.Options{
		Addr:     cfg.RDBCfg.Addr,
		Password: cfg.RDBCfg.Password,
		DB:       cfg.RDBCfg.DB,
	})

	// 测试连接
	ctx := context.Background()
	if err := rdb.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("Redis Ping: %v ", err)
	}

	return rdb, nil
}
