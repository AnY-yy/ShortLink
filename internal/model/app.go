package model

import (
	"shortURL/pkg/snowflake"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type App struct {
	Logger    *zap.Logger
	DB        *gorm.DB
	RDB       *redis.Client
	SnowFlake *snowflake.SnowFlake
}
