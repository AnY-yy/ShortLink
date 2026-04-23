package service

import (
	"context"
	"fmt"
	"shortURL/internal/bootstrap"
	"shortURL/internal/model"
	"shortURL/pkg/base62"
	"time"

	"go.uber.org/zap"
)

// Repository 数据库操作接口
type Repository interface {
	CreateURL(urlParams *model.URLParams) error
	IsExistLongURL(longURL string) (bool, error)
	IsExistShortURL(shortURL string) (bool, error)
	GetShortURL(LongURL string) (*model.CreateURLResponse, error)
}

// Cache 缓存服务接口
type Cache interface {
	CreateURL(ctx context.Context, req *model.CreateURLRequest) error
}

// ShortCodeGenerator 短链生成器接口
type ShortCodeGenerator interface {
	GenerateShortCode(snowflakeID int64) string
}

// SnowFlakeGenerator 雪花ID生成器接口
type SnowFlakeGenerator interface {
	GenerateSnowFlakeID() int64
}

// NewURLService 创建URL服务实例 返回给业务逻辑层调用
func NewURLService(repo Repository, cache Cache) *URLService {
	return &URLService{
		repo:               repo,
		cache:              cache,
		shortCodeGenerator: base62.NewShortCodeGenerator(),  // 短链生成器 自定义
		snowFlakeGenerator: bootstrap.Application.SnowFlake, // 雪花ID生成器 自定义
	}
}

type URLService struct { // URL服务接口
	repo               Repository
	cache              Cache
	shortCodeGenerator ShortCodeGenerator
	snowFlakeGenerator SnowFlakeGenerator
}

// CreateURL 创建短链服务
func (us *URLService) CreateURL(ctx context.Context, req *model.CreateURLRequest) (*model.CreateURLResponse, error) {
	var exist bool
	var err error
	var rep = &model.CreateURLResponse{}
	var urlParams = &model.URLParams{}
	// 进一步校验数据
	// 检查长链是否存在 如果存在返回短链 如果不存在继续进行逻辑处理
	exist, err = us.repo.IsExistLongURL(req.LongURL)
	if err != nil {
		bootstrap.Application.Logger.Error("检查长链是否存在失败", zap.Error(err))
		return nil, err
	}
	if exist {
		rep, err = us.repo.GetShortURL(req.LongURL)
		if err != nil {
			bootstrap.Application.Logger.Error("查询短链失败", zap.Error(err))
			return nil, err
		}
		bootstrap.Application.Logger.Info("长链已存在 自动返回短链", zap.String("shortURL", rep.ShortURL))
		return rep, nil
	}
	urlParams.LongURL = req.LongURL
	// 生成唯一雪花ID
	urlParams.ID = us.snowFlakeGenerator.GenerateSnowFlakeID()

	// 是否自定义短链
	if req.SelfShortUrl != "" {
		// 自定义短链是否存在
		exist, err = us.repo.IsExistShortURL(req.SelfShortUrl)
		if err != nil {
			bootstrap.Application.Logger.Error("检查自定义短链是否存在失败", zap.Error(err))
			return nil, err
		}
		if exist {
			bootstrap.Application.Logger.Error("自定义短链已经存在!")
			return nil, fmt.Errorf("自定义短链已经存在")
		}
		urlParams.IsCustom = true
		urlParams.SelfShortUrl = req.SelfShortUrl
		urlParams.ShortURL = req.SelfShortUrl
		rep.ShortURL = req.SelfShortUrl
	} else { // 未自定义短链 生成随机短链
		urlParams.IsCustom = false
		rep.ShortURL, err = us.getShortURL(0, urlParams.ID)
		if err != nil {
			bootstrap.Application.Logger.Error("生成短码失败", zap.Error(err))
			return nil, err
		}
		if urlParams.ShortURL == "" {
			bootstrap.Application.Logger.Error("生成短码失败", zap.Error(err))
			return nil, err
		}
	}

	// 是否设置过期时间
	if req.ExpireTime != nil { // ExpireTime是int类型的指针 将过期时间转换成时间戳
		if *req.ExpireTime == 0 { // 设置为0则永不过期
			// 设置为time.Time{} 数据库存储NULL 数据库插入会失败
			// 暂时先设置为100年 作为默认值 后续再根据需求调整过期时间逻辑
			urlParams.ExpireAt = time.Now().AddDate(100, 0, 0)
		} else {
			urlParams.ExpireAt = time.Now().Add(time.Hour * time.Duration(*req.ExpireTime))
		}
	} else { // 不设置过期时间 默认过期时间为1小时
		urlParams.ExpireAt = time.Now().Add(time.Hour)
	}

	// 调用repo接口 将数据写入数据库
	urlParams.CreatedAt = time.Now()
	errRepo := us.repo.CreateURL(urlParams)
	if errRepo != nil {
		bootstrap.Application.Logger.Error("数据库插入失败", zap.Error(errRepo))
		return nil, errRepo
	}

	// 将随机生成的短链也写入req中 用于缓存操作
	req.SelfShortUrl = urlParams.ShortURL
	// 调用cache接口 将数据写入缓存中
	errCache := us.cache.CreateURL(ctx, req)
	if errCache != nil {
		bootstrap.Application.Logger.Error("缓存插入失败", zap.Error(errCache))
		return nil, errCache
	}
	// 写入布隆过滤器

	return rep, nil
}

// RedirectURL 重定向服务 主要根据短链去查询对应的长链 返回给api调用
func (us *URLService) RedirectURL(ctx context.Context, req *model.RedirectURLRequest) (*model.RedirectURLResponse, error) {

	return nil, nil
}

// getShortURL 根据雪花ID随机生成短链服务 最多尝试生成n次 如果都失败 则返回空字符串
func (us *URLService) getShortURL(n int, snowflakeID int64) (string, error) {
	if n > 5 {
		return "", fmt.Errorf("尝试生成短链次数超过5次")
	}
	shortURL := us.shortCodeGenerator.GenerateShortCode(snowflakeID)
	// 检查短链是否存在 如果存在 则递归调用
	exist, err := us.repo.IsExistShortURL(shortURL)
	if err != nil {
		bootstrap.Application.Logger.Error("检查短链是否存在失败", zap.Error(err))
		return "", err
	}
	if exist {
		return us.getShortURL(n+1, snowflakeID)
	}
	return shortURL, nil
}
