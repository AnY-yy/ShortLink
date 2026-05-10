package config

import (
	"os"
	"sync"

	"gopkg.in/yaml.v3"
)

type DatabaseConfig struct {
	Host     string `yaml:"host"`
	Port     string `yaml:"port"`
	UserName string `yaml:"username"`
	Password string `yaml:"password"`
	DBName   string `yaml:"dbname"`
}

type RedisConfig struct {
	Addr     string `yaml:"addr"`
	Password string `yaml:"password"`
	DB       int    `yaml:"db"`
}

type IPInfoConfig struct {
	Token string `yaml:"token"`
}

type AppConfig struct {
	DBCfg     DatabaseConfig `yaml:"database"`
	RDBCfg    RedisConfig    `yaml:"redis"`
	IPInfoCfg IPInfoConfig   `yaml:"ipinfo"`
}

type ConfigLoader struct {
	Cfg *AppConfig
	mu  sync.RWMutex // 如果支持热更新 则需要加锁
}

type ConfigProvider interface {
	GetConfig() *AppConfig
}

// NewLoadConfig
// 构造函数 返回一个ConfigProvider接口
func NewLoadConfig() (ConfigProvider, error) {
	filePath := "./config/config.yml"
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	var appCfg AppConfig
	if err := yaml.Unmarshal(data, &appCfg); err != nil {
		return nil, err
	}

	return &ConfigLoader{Cfg: &appCfg}, nil
}

func (c *ConfigLoader) GetConfig() *AppConfig {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.Cfg
}
