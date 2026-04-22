package config

import (
	"fmt"

	"github.com/spf13/viper"
)

type Database struct {
	Host     string
	Port     string
	User     string
	Password string
	DB       string
}

type Redis struct {
	HostPort string
	Password string
	DB       int
}

type AppConfig struct { // 结构体字段变量名要与配置文件中的字段一致 或者使用mapstructure标签指定字段名
	DBCfg  Database `mapstructure:"Database"`
	RDBCfg Redis    `mapstructure:"Redis"`
}

var AppCfg = &AppConfig{}

// InitConfig 初始化配置文件
func InitConfig() {
	// Viper读取配置文件
	viper.SetConfigName("config")
	viper.SetConfigType("yml")
	viper.AddConfigPath("./config")

	// 读取配置文件
	if err := viper.ReadInConfig(); err != nil {
		panic(fmt.Errorf("配置文件读取失败: %v", err))
	}

	// 解析配置文件 读取到变量中
	if err := viper.Unmarshal(&AppCfg); err != nil {
		panic(fmt.Errorf("配置文件解析失败: %v", err))
	}
}
