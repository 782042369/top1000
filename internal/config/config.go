package config

import (
	"time"
)

// Config 应用程序配置
type Config struct {
	// 服务器配置
	Port string

	// 数据文件配置
	DataDir       string
	DataFilePath  string
	WebDistDir    string

	// API配置
	Top1000APIURL string

	// 缓存配置
	CacheDuration time.Duration

	// 数据过期时间
	DataExpireDuration time.Duration
}

var appConfig *Config

// Load 加载配置
func Load() *Config {
	if appConfig != nil {
		return appConfig
	}

	appConfig = &Config{
		Port:              	"7066",
		DataDir:            "./public",
		DataFilePath:       "./public/top1000.json",
		WebDistDir:         "./web-dist",
		Top1000APIURL:      "https://api.iyuu.cn/top1000.php",
		CacheDuration:      24 * time.Hour,
		DataExpireDuration: 24 * time.Hour,
	}

	return appConfig
}

// Get 获取配置实例
func Get() *Config {
	if appConfig == nil {
		return Load()
	}
	return appConfig
}
