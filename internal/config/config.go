package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
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

// ValidationError 配置验证错误（收集所有错误）
type ValidationError struct {
	errors []string
}

// Error 实现 error 接口
func (e *ValidationError) Error() string {
	return fmt.Sprintf("配置验证失败: %s", strings.Join(e.errors, "、"))
}

// Add 添加验证错误
func (e *ValidationError) Add(field string) {
	e.errors = append(e.errors, field)
}

// IsValid 检查是否有错误
func (e *ValidationError) IsValid() bool {
	return len(e.errors) == 0
}

// Validate 验证配置的有效性（返回所有错误）
func Validate() error {
	var errs ValidationError

	if appConfig.RedisAddr == "" {
		errs.Add("REDIS_ADDR")
	}
	if appConfig.RedisPassword == "" {
		errs.Add("REDIS_PASSWORD")
	}

	if !errs.IsValid() {
		return &errs
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
