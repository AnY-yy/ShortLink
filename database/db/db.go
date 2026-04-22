package db

import (
	"fmt"
	"shortURL/config"
	"shortURL/internal/model"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// NewDB 初始化数据库连接
func NewDB() *gorm.DB {
	dbStr := fmt.Sprintf("%s:%s@(%s:%s)/%s?charset=utf8&parseTime=True&loc=Local",
		config.AppCfg.DBCfg.User,
		config.AppCfg.DBCfg.Password,
		config.AppCfg.DBCfg.Host,
		config.AppCfg.DBCfg.Port,
		config.AppCfg.DBCfg.DB,
	)

	// 新版GORM初始化
	db, dbError := gorm.Open(mysql.Open(dbStr), &gorm.Config{})

	if dbError != nil {
		panic(fmt.Errorf("数据库连接失败: %v", dbError))
	}

	// 完成数据库迁移 创建表结构
	if dbError = db.AutoMigrate(&model.URLParams{}); dbError != nil {
		panic(fmt.Errorf("数据库迁移失败: %v", dbError))
	}

	return db
}
