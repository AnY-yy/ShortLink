package service

import (
	"context"
	"fmt"
	"shortURL/internal/bootstrap"
	"shortURL/internal/model"

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
}

// ShortCodeGenerator 短链生成器接口
type ShortCodeGenerator interface {
}

// NewURLService 创建URL服务实例 返回给业务逻辑层调用
func NewURLService(repo Repository, cache Cache) *URLService {
	return &URLService{
		repo:  repo,
		cache: cache,
		// shortCodeGenerator :shortCodeGenerator,  // 短链生成器 自定义
	}
}

type URLService struct { // URL服务接口
	repo               Repository
	cache              Cache
	shortCodeGenerator ShortCodeGenerator
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
		return rep, nil
	}

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
		rep.ShortURL = req.SelfShortUrl
		urlParams.SelfShortUrl = req.SelfShortUrl
		urlParams.ShortURL = req.SelfShortUrl
	}

	// 生成唯一雪花ID

	return rep, nil
}

func (us *URLService) RedirectURL(ctx context.Context, req *model.RedirectURLRequest) (*model.RedirectURLResponse, error) {

	return nil, nil
}
