package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

// 默认值常量
const (
	DefaultPort            = "7066"
	DefaultWebDistDir      = "./web-dist"
	DefaultAPIURL          = "https://api.iyuu.cn/top1000.php"
	DefaultCacheDuration   = 24 * time.Hour
	DefaultDataExpire      = 24 * time.Hour
	DefaultRedisDB         = 0
	DefaultRedisKeyPrefix  = "top1000:"
)

// Config 应用程序配置
type Config struct {
	Port               string        // 服务器端口
	WebDistDir         string        // 前端构建目录
	Top1000APIURL      string        // 数据源API地址
	CacheDuration      time.Duration // 静态文件缓存时间
	DataExpireDuration time.Duration // 数据过期检测阈值
	RedisAddr          string        // Redis地址（必须配置）
	RedisPassword      string        // Redis密码（必须配置）
	RedisDB            int           // Redis数据库编号
	RedisKeyPrefix     string        // Redis键前缀
}

var appConfig *Config

// Load 加载配置（单例模式）
func Load() *Config {
	if appConfig != nil {
		return appConfig
	}

	appConfig = &Config{
		Port:               getEnv("PORT", DefaultPort),
		WebDistDir:         getEnv("WEB_DIST_DIR", DefaultWebDistDir),
		Top1000APIURL:      getEnv("TOP1000_API_URL", DefaultAPIURL),
		CacheDuration:      getEnvDuration("CACHE_DURATION", DefaultCacheDuration),
		DataExpireDuration: getEnvDuration("DATA_EXPIRE_DURATION", DefaultDataExpire),
		RedisAddr:          getEnv("REDIS_ADDR", ""),
		RedisPassword:      getEnv("REDIS_PASSWORD", ""),
		RedisDB:            getEnvInt("REDIS_DB", DefaultRedisDB),
		RedisKeyPrefix:     getEnv("REDIS_KEY_PREFIX", DefaultRedisKeyPrefix),
	}

	return appConfig
}

// Validate 验证配置的有效性
func Validate() error {
	if appConfig.RedisAddr == "" {
		return fmt.Errorf("REDIS_ADDR 环境变量未设置")
	}
	if appConfig.RedisPassword == "" {
		return fmt.Errorf("REDIS_PASSWORD 环境变量未设置")
	}
	return nil
}

// Get 获取配置实例
func Get() *Config {
	if appConfig == nil {
		return Load()
	}
	return appConfig
}

// getEnv 获取环境变量，如果不存在则返回默认值
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getEnvInt 获取整数环境变量
func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intVal, err := strconv.Atoi(value); err == nil {
			return intVal
		}
	}
	return defaultValue
}

// getEnvDuration 获取时长环境变量（支持30s、5m、24h等格式）
func getEnvDuration(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return defaultValue
}
