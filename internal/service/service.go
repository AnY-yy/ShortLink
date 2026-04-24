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
	FindLongURL(shortURL string) (*model.RedirectURLResponse, error)
}

// Cache 缓存服务接口
type Cache interface {
	CreateURL(ctx context.Context, req *model.CreateURLRequest) error
	GetURL(ctx context.Context, shortURL string) (*model.RedirectURLResponse, error)
}

// SBloomFilter 布隆过滤器接口
type SBloomFilter interface {
	AddBloomFilterElem(data []byte)
	IsExistElem(data []byte) bool
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
		bloomFilter:        bootstrap.Application.BloomFilter,
	}
}

type URLService struct { // URL服务接口
	repo               Repository
	cache              Cache
	shortCodeGenerator ShortCodeGenerator
	snowFlakeGenerator SnowFlakeGenerator
	bloomFilter        SBloomFilter
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
		urlParams.ShortURL, err = us.getShortURL(0, urlParams.ID)
		rep.ShortURL = urlParams.ShortURL
		if err != nil {
			bootstrap.Application.Logger.Error("生成短码失败", zap.Error(err))
			return nil, err
		}
		if urlParams.ShortURL == "" {
			bootstrap.Application.Logger.Error("获取到的短码为空", zap.Error(err))
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
		// 也要给req.ExpireTime传入值 用于缓存操作
		// *req.ExpireTime = 1 致命错误 这样是使用了空指针
		defaultTime := 1
		req.ExpireTime = &defaultTime
	}

	// 调用repo接口 将数据写入数据库
	urlParams.CreatedAt = time.Now()
	errRepo := us.repo.CreateURL(urlParams)
	if errRepo != nil {
		bootstrap.Application.Logger.Error("数据库插入失败", zap.Error(errRepo))
		return nil, errRepo
	}

	// 将随机生成的短链与过期时间也写入req中 用于缓存操作
	req.SelfShortUrl = urlParams.ShortURL
	// 调用cache接口 将数据写入缓存中
	errCache := us.cache.CreateURL(ctx, req)
	if errCache != nil {
		bootstrap.Application.Logger.Error("缓存插入失败", zap.Error(errCache))
		return nil, errCache
	}
	// 写入布隆过滤器
	us.bloomFilter.AddBloomFilterElem([]byte(urlParams.ShortURL))

	return rep, nil
}

// RedirectURL 重定向服务 主要根据短链去查询对应的长链 返回给api调用
func (us *URLService) RedirectURL(ctx context.Context, req *model.RedirectURLRequest) (*model.RedirectURLResponse, error) {
	var rep = &model.RedirectURLResponse{}
	var err error
	// 布隆过滤器 防止缓存穿透问题 过滤掉绝对不存在的短链
	existBloom := us.bloomFilter.IsExistElem([]byte(req.ShortURL))
	if !existBloom {
		// 布隆过滤器中不存在的直接退出
		return nil, fmt.Errorf("短链不存在")
	}

	// 缓存
	rep, err = us.cache.GetURL(ctx, req.ShortURL)
	if err != nil {
		bootstrap.Application.Logger.Error("查询缓存失败")
		return nil, fmt.Errorf("查询缓存失败: %w", err)
	}
	if rep.LongURL != "" {
		bootstrap.Application.Logger.Info("缓存中存在该短链", zap.String("shortURL", rep.ShortURL))
		return rep, nil
	}

	// 如果缓存中不存在 则需要查询数据库
	rep, err = us.repo.FindLongURL(req.ShortURL)
	if err != nil {
		return nil, fmt.Errorf("查询数据库失败: %w", err)
	}
	if rep == nil {
		return nil, fmt.Errorf("数据库中不存在该短链")
	}

	// 如果数据库存在但缓存不存在 则需要同步到缓存中
	remainingHours := int(rep.ExpireAt.Sub(time.Now()).Hours())
	err = us.cache.CreateURL(ctx, &model.CreateURLRequest{
		LongURL:      rep.LongURL,
		ExpireTime:   &remainingHours,
		SelfShortUrl: rep.ShortURL,
	})
	if err != nil {
		return rep, fmt.Errorf("缓存同步失败")
	}

	return rep, nil
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
