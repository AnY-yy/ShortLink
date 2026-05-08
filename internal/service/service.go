package service

import (
	"context"
	"shortURL/internal/model"

	"go.uber.org/zap"
)

// SnowFlakeIDGenerator 雪花ID生成器接口
type SnowFlakeIDGenerator interface {
}

// SBloomFilter 自定义布隆过滤器接口
type SBloomFilter interface {
}

// Cache 缓存接口
type Cache interface {
}

// Repository 数据仓库接口
type Repository interface {
	LongURLIsExist(ctx context.Context, longURL string) (bool, error)
}

// Logger 日志接口
type Logger interface {
	Info(msg string, fields ...zap.Field)
	Error(msg string, fields ...zap.Field)
	Warn(msg string, fields ...zap.Field)
}

// Service 服务核心结构体
type Service struct {
	logger               Logger
	repo                 Repository
	cache                Cache
	snowFlakeIDGenerator SnowFlakeIDGenerator
	sbloomFilter         SBloomFilter
}

// NewService 返回服务核心结构体实例
// 外部接口依赖 导出Service类型
func NewService(logger Logger, repo Repository, cache Cache) *Service {
	return &Service{
		logger: logger,
		repo:   repo,
		cache:  cache,
	}
}

// CreateURL 创建短链服务
func (s *Service) CreateURL(ctx context.Context, req *model.CreateURLRequest) (*model.CreateURLReponse, error) {
	var exist bool
	var err error
	var rep = &model.CreateURLReponse{}
	var urlParam = &model.URLParams{}

	urlParam.LongURL = req.LongURL
	// 进一步校验请求参数
	// 查询长链是否生成了短链
	exist, err = s.repo.LongURLIsExist(ctx, req.LongURL)
	if err != nil {
		s.logger.Error("在数据库中查询"+req.LongURL+"的短链信息出现错误!", zap.Error(err))
		return nil, err
	}
	if exist { // 如果长链接已经生成了短链 直接返回响应

	}
	// 生成唯一短链的逻辑: 生成全局唯一雪花ID -> 是否自定义短码 -> 是否设置过期时间 -> 存入数据库、缓存、布隆过滤器 -> 返回响应
	// 生成全局唯一雪花ID

	// 是否自定义短码

	// 是否设置过期时间

	// 写入数据库

	// 写入缓存

	// 写入布隆过滤器

	// 返回响应参数
	return rep, nil
}
