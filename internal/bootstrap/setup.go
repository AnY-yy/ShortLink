package bootstrap

import (
	"fmt"
	"shortURL/config"
	db2 "shortURL/database/db"
	rdb2 "shortURL/database/rdb"
	"shortURL/internal/cache"
	"shortURL/internal/repo"
	"shortURL/internal/service"
	"shortURL/pkg/base62"
	"shortURL/pkg/bloomFilter"
	"shortURL/pkg/logger"
	"shortURL/pkg/snowflake"
	"time"

	"go.uber.org/zap"
)

// App 应用核心结构体
// 相对与v1 v2将全部接口类型作为字段存储在App结构体中 而不是接口类型变量了
// 依赖注入容器
type App struct {
	Logger               logger.Logger
	Config               config.ConfigProvider
	Repo                 repo.Repository
	Cache                cache.Cache
	SnowFlakeIDGenerator snowflake.SnowflakeGenerator
	Base62Generator      base62.ShortCodeGenerator
	BloomFilter          bloomFilter.BloomFilter
	Service              *service.Service
}

func SetUp() *App {
	// 初始化日志接口
	lg := logger.NewZapLogger()
	if lg == nil {
		fmt.Println(time.Now(), "日志接口初始化失败!")
		panic(nil)
	}

	// 初始化配置接口
	cfgProvider, err := config.NewLoadConfig()
	if err != nil {
		lg.Panic("配置接口初始化失败!", zap.Error(err))
	}

	// 初始化数据仓库接口
	db, err := db2.NewDB(cfgProvider)
	if err != nil {
		lg.Panic("数据库连接初始化失败!", zap.Error(err))
	}
	dbRepo := repo.NewRepository(db)
	if dbRepo == nil {
		lg.Panic("数据仓库接口初始化失败!")
	}

	// 初始化缓存接口
	rdb, err := rdb2.NewRDB(cfgProvider)
	if err != nil {
		lg.Panic("缓存连接初始化失败!", zap.Error(err))
	}
	rdbCache := cache.NewCache(rdb)
	if rdbCache == nil {
		lg.Panic("缓存接口初始化失败!")
	}

	// 初始化雪花ID生成器接口
	snowFlakeIDGenerator, err := snowflake.NewSnowFlake(1)
	if err != nil {
		lg.Panic("雪花ID生成器初始化失败!", zap.Error(err))
	}

	// 初始化短码生成器接口
	base62Generator := base62.NewShortGenerator()
	if base62Generator == nil {
		lg.Panic("短码生成器初始化失败!")
	}

	// 初始化布隆过滤器接口
	sbf, err := bloomFilter.NewBloomFilter(10000, 0.01)
	if err != nil {
		lg.Panic("布隆过滤器初始化失败!", zap.Error(err))
	}

	// 初始化服务接口
	serviceType := service.NewService(lg, dbRepo, rdbCache, snowFlakeIDGenerator, base62Generator, sbf)
	if serviceType == nil {
		lg.Panic("服务接口初始化失败!")
	}

	lg.Info("应用初始化完成! {配置信息 日志接口 数据仓库接口 缓存接口 雪花ID生成器接口 服务接口}")
	return &App{
		Logger:               lg,
		Config:               cfgProvider,
		Repo:                 dbRepo,
		Cache:                rdbCache,
		SnowFlakeIDGenerator: snowFlakeIDGenerator,
		Base62Generator:      base62Generator,
		BloomFilter:          sbf,
		Service:              serviceType,
	}
}
