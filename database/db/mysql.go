package db

import (
	"fmt"
	"shortURL/config"
	"shortURL/internal/model"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// NewDB 初始化数据库MySQL连接
func NewDB(cfgProvider config.ConfigProvider) (*gorm.DB, error) {
	cfg := cfgProvider.GetConfig()
	dbStr := fmt.Sprintf("%s:%s@(%s:%s)/%s?charset=utf8&parseTime=True&loc=Local",
		cfg.DBCfg.UserName,
		cfg.DBCfg.Password,
		cfg.DBCfg.Host,
		cfg.DBCfg.Port,
		cfg.DBCfg.DBName,
	)

	// 新版GORM初始化
	db, dbError := gorm.Open(mysql.Open(dbStr), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent), // 关闭数据库的日志打印
	})

	if dbError != nil {
		return nil, fmt.Errorf("数据库链接失败! %v", dbError)
	}

	// 完成数据库迁移 创建表结构
	if dbError = db.AutoMigrate(&model.URLParams{}); dbError != nil {
		return nil, fmt.Errorf("数据库迁移失败: %v", dbError)
	}

	return db, nil
}
