package bootstrap

import (
	"shortURL/config"
	"shortURL/database/db"
	"shortURL/database/rdb"
	"shortURL/internal/model"
	"shortURL/pkg/logger"
	"shortURL/pkg/snowflake"
)

var Application *model.App

// Setup 初始化组件信息
func Setup() {
	Application = &model.App{}

	// 初始化配置文件
	config.InitConfig()

	Application.DB = db.NewDB()
	Application.RDB = rdb.NewRDB()
	Application.Logger = logger.NewLoger()
	Application.SnowFlake, _ = snowflake.NewSnowFlake(1)
}
