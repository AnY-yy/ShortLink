package bootstrap

import (
	"fmt"
	"shortURL/config"
	db2 "shortURL/database/db"
	rdb2 "shortURL/database/rdb"
	"shortURL/internal/cache"
	"shortURL/internal/repo"
	"shortURL/internal/service"
	"shortURL/pkg/logger"
	"time"

	"go.uber.org/zap"
)

// App 应用核心结构体
// 相对与v1 v2将全部接口类型作为字段存储在App结构体中 而不是接口类型变量了
// 依赖注入容器
type App struct {
	Logger  logger.Logger
	Config  config.ConfigProvider
	Repo    repo.Repository
	Cache   cache.Cache
	Service *service.Service
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

	// 初始化服务接口
	sve := service.NewService(lg, dbRepo, rdbCache)
	if sve == nil {
		lg.Panic("服务接口初始化失败!")
	}
	lg.Info("应用初始化完成! {配置信息 日志接口 数据仓库接口 缓存接口 服务接口}")
	return &App{
		Logger:  lg,
		Config:  cfgProvider,
		Repo:    dbRepo,
		Cache:   rdbCache,
		Service: sve,
	}
}
