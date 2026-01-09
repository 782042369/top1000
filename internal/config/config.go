package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

// Config 应用程序配置（使用环境变量管理所有配置）
type Config struct {
	// 服务器配置
	Port string

	// 前端构建目录
	WebDistDir string

	// API配置
	Top1000APIURL string

	// 缓存配置
	CacheDuration time.Duration

	// 数据过期时间
	DataExpireDuration time.Duration

	// Redis 配置（必须配置，否则无法运行）
	RedisEnabled   bool
	RedisAddr      string
	RedisPassword  string
	RedisDB        int
	RedisKeyPrefix string
}

var appConfig *Config

// getEnv 获取环境变量，如果不存在则返回默认值
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getEnvInt 获取整数环境变量（解析失败就用默认值）
func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intVal, err := strconv.Atoi(value); err == nil {
			return intVal
		}
	}
	return defaultValue
}

// getEnvDuration 获取时长环境变量（支持秒、分钟、小时，如30s、5m、24h）
func getEnvDuration(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return defaultValue
}

// Load 加载配置（单例模式，全局只有一个配置对象）
func Load() *Config {
	if appConfig != nil {
		return appConfig
	}

	appConfig = &Config{
		// 服务器配置（支持环境变量覆盖）
		Port: getEnv("PORT", "7066"),

		// 前端构建目录
		WebDistDir: getEnv("WEB_DIST_DIR", "./web-dist"),

		// API配置
		Top1000APIURL: getEnv("TOP1000_API_URL", "https://api.iyuu.cn/top1000.php"),

		// 缓存配置
		CacheDuration:      getEnvDuration("CACHE_DURATION", 24*time.Hour),
		DataExpireDuration: getEnvDuration("DATA_EXPIRE_DURATION", 24*time.Hour),

		// Redis 配置（必须通过环境变量设置，不提供默认值，安全优先）
		RedisEnabled:   getEnv("REDIS_ENABLED", "true") == "true",
		RedisAddr:      getEnv("REDIS_ADDR", ""),
		RedisPassword:  getEnv("REDIS_PASSWORD", ""),
		RedisDB:        getEnvInt("REDIS_DB", 0),
		RedisKeyPrefix: getEnv("REDIS_KEY_PREFIX", "top1000:"),
	}

	return appConfig
}

// Validate 验证配置的有效性（启动时检查，配置不正确则直接退出）
func Validate() error {
	if appConfig.RedisEnabled {
		if appConfig.RedisAddr == "" {
			return fmt.Errorf("REDIS_ADDR 环境变量未设置，请检查.env文件或环境变量配置")
		}
		if appConfig.RedisPassword == "" {
			return fmt.Errorf("REDIS_PASSWORD 环境变量未设置，请检查.env文件或环境变量配置")
		}
	}
	return nil
}

// Get 获取配置实例（单例模式，首次调用时初始化）
func Get() *Config {
	if appConfig == nil {
		return Load()
	}
	return appConfig
}
