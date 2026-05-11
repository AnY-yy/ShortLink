package service

import (
	"context"
	"errors"
	"fmt"
	"shortURL/internal/model"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
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
	RedirectURL(ctx context.Context, shortURL string) (*model.RedirectURLResponse, error)
}

// Repository 数据仓库接口
type Repository interface {
	LongURLIsExist(ctx context.Context, longURL string) (bool, error)
	GetShortCode(ctx context.Context, longURL string) (*model.CreateURLReponse, error)
	ShortURLIsExist(ctx context.Context, shortURL string) (bool, error)
	CreateURL(url *model.URLParams) error
	RedirectURL(ctx context.Context, shortURL string) (*model.RedirectURLResponse, error)
}

// Logger 日志接口
type Logger interface {
	Info(msg string, fields ...zap.Field)
	Error(msg string, fields ...zap.Field)
	Warn(msg string, fields ...zap.Field)
}

type Service interface {
	CreateURL(ctx context.Context, req *model.CreateURLRequest) (*model.CreateURLReponse, error)
	getShortCode(n int, snowFlakeID int64) (string, error)
	RedirectURL(ctx context.Context, req *model.RedirectURLRequest) (*model.RedirectURLResponse, error)
}

// ServiceModel 服务核心结构体
type ServiceModel struct {
	logger               Logger
	repo                 Repository
	cache                Cache
	snowFlakeIDGenerator SnowFlakeIDGenerator
	shortCodeGenerator   ShortCodeGenerator
	bloomFilter          SBloomFilter
}

// NewService 返回服务核心结构体实例
// 外部接口依赖 导出Service类型
func NewService(logger Logger, repo Repository, cache Cache, generator SnowFlakeIDGenerator, base62Generator ShortCodeGenerator, bloomFilter SBloomFilter) Service {
	return &ServiceModel{
		logger:               logger,
		repo:                 repo,
		cache:                cache,
		snowFlakeIDGenerator: generator,
		shortCodeGenerator:   base62Generator,
		bloomFilter:          bloomFilter,
	}
}

// CreateURL 创建短链服务
func (s *ServiceModel) CreateURL(ctx context.Context, req *model.CreateURLRequest) (*model.CreateURLReponse, error) {
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

	// 写入数据库
	if errRepo := s.repo.CreateURL(urlParam); errRepo != nil {
		s.logger.Error(req.LongURL+"写入数据库失败", zap.Error(errRepo))
		return nil, fmt.Errorf("写入数据库失败:%v", errRepo)
	}
	s.logger.Info(req.LongURL + "写入数据库成功")

	// 写入缓存
	req.SelfShortCode = rep.ShortCode // 将生成的短码赋值给请求参数 缓存操作需要使用
	errCache := s.cache.CreateURL(ctx, req)
	if errCache != nil {
		s.logger.Error(req.LongURL+"写入缓存失败", zap.Error(errCache))
		return nil, fmt.Errorf("写入缓存失败:%v", errCache)
	}
	s.logger.Info(req.LongURL + "写入缓存成功")

	// 写入布隆过滤器
	s.bloomFilter.AddBloomFilterElem([]byte(rep.ShortCode))
	s.logger.Info(req.LongURL + "写入布隆过滤器成功")

	// 返回响应参数
	s.logger.Info("返回响应参数", zap.Any("响应体", rep))
	return rep, nil
}

// getShortCode
// 生成短码中间商 校验短码是否唯一 直到生成5次退出
func (s *ServiceModel) getShortCode(n int, snowFlakeID int64) (string, error) {
	if n > 5 {
		return "", errors.New("生成短码失败,已生成5次")
	}
	shortCode := s.shortCodeGenerator.GenerateShortCode(snowFlakeID)
	s.logger.Info(strconv.Itoa(n) + "次生成短码为" + shortCode)
	exist, err := s.repo.ShortURLIsExist(context.Background(), shortCode)
	if err != nil {
		s.logger.Error("在数据库中检查"+shortCode+"的短链是否存在出现错误!", zap.Error(err))
		return "", fmt.Errorf("service/createurl.go/getShortCode 生成短码时数据交互错误:%v", err)
	}
	if exist {
		return s.getShortCode(n+1, snowFlakeID)
	}
	s.logger.Info("生成短码"+shortCode+"即将返回", zap.String("ShortCode", shortCode))
	return shortCode, nil
}

// RedirectURL
// 查找短码对应的长链
func (s *ServiceModel) RedirectURL(ctx context.Context, req *model.RedirectURLRequest) (*model.RedirectURLResponse, error) {
	var rep = &model.RedirectURLResponse{}
	var err error
	var exist bool
	s.logger.Info("开始寻找" + req.ShortURL + "对应的长链")

	// 布隆过滤器
	exist = s.bloomFilter.IsExistData([]byte(req.ShortURL))
	if !exist {
		s.logger.Info(req.ShortURL + "不在布隆过滤器中,其对应的长链一定不存在")
		return nil, errors.New(req.ShortURL + "不存在")
	}

	// 缓存
	rep, err = s.cache.RedirectURL(ctx, req.ShortURL)
	if err != nil && errors.Is(err, redis.Nil) {
		err = nil
		rep = &model.RedirectURLResponse{}
	}
	if err != nil {
		s.logger.Error(req.ShortURL+"查询缓存时错误", zap.Error(err))
		return nil, errors.New("查询缓存时错误")
	}
	if rep.LongURL != "" {
		s.logger.Info("在缓存中查询到对应长链", zap.String(rep.ShortURL, rep.LongURL))
		return rep, nil
	}

	// 数据库
	rep, err = s.repo.RedirectURL(ctx, req.ShortURL)
	if err != nil {
		s.logger.Error(req.ShortURL+"查询数据库时错误", zap.Error(err))
		return nil, errors.New("数据层出现错误")
	}
	if rep == nil {
		s.logger.Error(req.ShortURL + "没有对应的长链信息")
		return nil, errors.New(req.ShortURL + "不存在对应的长链信息,无效的请求")
	}

	// 如果数据库存在而缓存中不存在 则需要将数据同步到缓存中
	expireTime := int(rep.ExpireAt.Sub(time.Now()).Hours())
	err = s.cache.CreateURL(ctx, &model.CreateURLRequest{
		LongURL:       rep.LongURL,
		SelfShortCode: rep.ShortURL,
		ExpireTime:    &expireTime,
	})
	if err != nil {
		s.logger.Warn("数据库与缓存数据同步失败", zap.Any("LinkData", rep))
		return rep, errors.New("数据库与缓存数据同步失败")
	}
	s.logger.Info("数据库与缓存数据同步成功", zap.Any("LinkData", rep))

	return rep, nil
}
