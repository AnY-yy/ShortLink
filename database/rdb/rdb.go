package rdb

import (
	"context"
	"fmt"
	"shortURL/config"

	"github.com/redis/go-redis/v9"
)

// NewRDB 创建Redis连接实例
func NewRDB() *redis.Client {
	rdb := redis.NewClient(&redis.Options{
		Addr:     config.AppCfg.RDBCfg.HostPort,
		Password: config.AppCfg.RDBCfg.Password,
		DB:       config.AppCfg.RDBCfg.DB,
	})

	// 测试连接
	ctx := context.Background()
	_, err := rdb.Ping(ctx).Result()
	if err != nil {
		panic(fmt.Errorf(" Redis连接失败:%v", err))
	}

	return rdb
}
