package service

import (
	"context"
	"errors"
	"fmt"
	"shortURL/internal/model"
	"strconv"
	"time"

	"go.uber.org/zap"
)

// SnowFlakeIDGenerator 雪花ID生成器接口
type SnowFlakeIDGenerator interface {
	GenerateSnowFlakeID() int64
}

// SBloomFilter 自定义布隆过滤器接口
type SBloomFilter interface {
	AddBloomFilterElem(data []byte)
	IsExistData(data []byte) bool
}

type ShortCodeGenerator interface {
	GenerateShortCode(snowFlakeID int64) string
}

// Cache 缓存接口
type Cache interface {
	CreateURL(ctx context.Context, req *model.CreateURLRequest) error
}

// Repository 数据仓库接口
type Repository interface {
	LongURLIsExist(ctx context.Context, longURL string) (bool, error)
	GetShortCode(ctx context.Context, longURL string) (*model.CreateURLReponse, error)
	ShortURLIsExist(ctx context.Context, shortURL string) (bool, error)
	CreateURL(url *model.URLParams) error
}

// Logger 日志接口
type Logger interface {
	Info(msg string, fields ...zap.Field)
	Error(msg string, fields ...zap.Field)
}

// Service 服务核心结构体
type Service struct {
	logger               Logger
	repo                 Repository
	cache                Cache
	snowFlakeIDGenerator SnowFlakeIDGenerator
	shortCodeGenerator   ShortCodeGenerator
	bloomFilter          SBloomFilter
}

// NewService 返回服务核心结构体实例
// 外部接口依赖 导出Service类型
func NewService(logger Logger, repo Repository, cache Cache, generator SnowFlakeIDGenerator, base62Generator ShortCodeGenerator, bloomFilter SBloomFilter) *Service {
	return &Service{
		logger:               logger,
		repo:                 repo,
		cache:                cache,
		snowFlakeIDGenerator: generator,
		shortCodeGenerator:   base62Generator,
		bloomFilter:          bloomFilter,
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
	s.logger.Info("开始检验请求体中的参数", zap.Any("req", req))
	exist, err = s.repo.LongURLIsExist(ctx, req.LongURL)
	if err != nil {
		s.logger.Error("在数据库中检查"+req.LongURL+"的短链是否存在出现错误!", zap.Error(err))
		return nil, fmt.Errorf("数据交互错误:%v", err)
	}
	if exist {
		s.logger.Info("存在" + req.LongURL + "对应的短链记录,开始查表查询短链信息")
		rep, err = s.repo.GetShortCode(ctx, req.LongURL)
		if err != nil {
			s.logger.Error("查询"+req.LongURL+"对应的短链信息出现错误", zap.Error(err))
			return nil, fmt.Errorf("数据交互错误:%v", err)
		}
		s.logger.Info("查询到短链信息,即将返回响应", zap.String(req.LongURL, rep.ShortCode))
		return rep, nil
	}

	s.logger.Info("长链接" + req.LongURL + "未生成短链,开始执行生成短链逻辑操作")
	// 生成唯一短链的逻辑: 生成全局唯一雪花ID -> 是否自定义短码 -> 是否设置过期时间 -> 存入数据库、缓存、布隆过滤器 -> 返回响应
	s.logger.Info("开始生成全局唯一雪花ID")
	urlParam.ID = s.snowFlakeIDGenerator.GenerateSnowFlakeID()
	s.logger.Info(req.LongURL+"成功生成全局唯一雪花ID", zap.Int64("ID", urlParam.ID))

	if req.SelfShortCode != "" {
		s.logger.Info(req.LongURL + "预自定义短码为" + req.SelfShortCode + "开始校验是否已存在")
		exist, err := s.repo.ShortURLIsExist(ctx, req.SelfShortCode)
		if err != nil {
			s.logger.Error("在数据库中检查"+req.SelfShortCode+"的短链是否存在出现错误!", zap.Error(err))
			return nil, fmt.Errorf("数据交互错误:%v", err)
		}
		if exist {
			s.logger.Error("自定义短码"+req.SelfShortCode+"已存在,请重新输入", zap.String("ShortCode", req.SelfShortCode))
			return nil, fmt.Errorf("自定义短码已存在")
		}
		s.logger.Info(req.SelfShortCode + "自定义短码合法")
		urlParam.ShortURL = req.SelfShortCode
		urlParam.IsCustom = true
		urlParam.SelfShortUrl = req.SelfShortCode
		rep.ShortCode = req.SelfShortCode
	} else {
		s.logger.Info(req.LongURL + "未自定义短码,开始生成随机短码")
		urlParam.ShortURL, err = s.getShortCode(1, urlParam.ID)
		if err != nil {
			s.logger.Error("生成随机短码失败", zap.Error(err))
			return nil, fmt.Errorf("生成随机短码失败:%v", err)
		}
		rep.ShortCode = urlParam.ShortURL
		urlParam.IsCustom = false
	}
	// 是否设置过期时间
	if req.ExpireTime != nil {
		s.logger.Info(fmt.Sprintf("%s预设置过期时间为:%d小时", req.LongURL, *req.ExpireTime))
		if *req.ExpireTime == 0 { // 0为永不过期
			urlParam.ExpireAt = time.Now().AddDate(100, 0, 0)
		} else {
			urlParam.ExpireAt = time.Now().Add(time.Hour * time.Duration(*req.ExpireTime))
		}
	} else {
		s.logger.Info(req.LongURL + "未预过期时间,默认过期时间为1小时")
		urlParam.ExpireAt = time.Now().Add(time.Hour)
	}
	urlParam.CreatedAt = time.Now()
	s.logger.Info(fmt.Sprintf("%s的创建时间为为:%v,过期时间为:%v", req.LongURL, urlParam.CreatedAt.Format(time.RFC3339), urlParam.ExpireAt.Format(time.RFC3339)))

	s.logger.Info(req.LongURL + "开始写入数据库")
	if errRepo := s.repo.CreateURL(urlParam); errRepo != nil {
		s.logger.Error(req.LongURL+"写入数据库失败", zap.Error(errRepo))
		return nil, fmt.Errorf("写入数据库失败:%v", errRepo)
	}
	s.logger.Info(req.LongURL + "写入数据库成功")

	s.logger.Info(req.LongURL + "开始写入缓存")
	req.SelfShortCode = rep.ShortCode // 将生成的短码赋值给请求参数 缓存操作需要使用
	errCache := s.cache.CreateURL(ctx, req)
	if errCache != nil {
		s.logger.Error(req.LongURL+"写入缓存失败", zap.Error(errCache))
		return nil, fmt.Errorf("写入缓存失败:%v", errCache)
	}
	s.logger.Info(req.LongURL + "写入缓存成功")

	s.logger.Info(req.LongURL + "开始写入布隆过滤器")
	s.bloomFilter.AddBloomFilterElem([]byte(req.LongURL))
	s.logger.Info(req.LongURL + "写入布隆过滤器成功")

	// 返回响应参数
	s.logger.Info("返回响应参数", zap.Any("响应体", rep))
	return rep, nil
}

// getShortCode
// 生成短码中间商 校验短码是否唯一 直到生成5次退出
func (s *Service) getShortCode(n int, snowFlakeID int64) (string, error) {
	if n > 5 {
		return "", errors.New("生成短码失败,已生成5次")
	}
	shortCode := s.shortCodeGenerator.GenerateShortCode(snowFlakeID)
	s.logger.Info(strconv.Itoa(n) + "次生成短码为" + shortCode)
	exist, err := s.repo.ShortURLIsExist(context.Background(), shortCode)
	if err != nil {
		s.logger.Error("在数据库中检查"+shortCode+"的短链是否存在出现错误!", zap.Error(err))
		return "", fmt.Errorf("service/service.go/getShortCode 生成短码时数据交互错误:%v", err)
	}
	if exist {
		return s.getShortCode(n+1, snowFlakeID)
	}
	s.logger.Info("生成短码"+shortCode+"即将返回", zap.String("ShortCode", shortCode))
	return shortCode, nil
}
