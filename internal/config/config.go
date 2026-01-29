package config

import (
	"fmt"
	"os"
	"strings"
	"time"
)

// 默认值常量
const (
	DefaultPort         = "7066"
	DefaultWebDistDir   = "./web-dist"
	DefaultAPIURL       = "https://api.iyuu.cn/top1000.php"
	DefaultDataExpire   = 24 * time.Hour // 数据过期检测阈值
	DefaultRedisDB      = 0              // Redis数据库编号
	DefaultRedisKey     = "top1000:data" // Redis key（Top1000数据）
	DefaultSitesKey     = "top1000:sites" // Redis key（站点数据）
	DefaultSitesExpire  = 24 * time.Hour // 站点数据过期时间
)

// Config 应用程序配置（只保留必须从环境变量读取的配置）
type Config struct {
	RedisAddr          string // Redis地址（必须配置）
	RedisPassword      string // Redis密码（必须配置）
	RedisDB            int    // Redis数据库编号（可选，默认0）
	IYYUSign           string // IYUU签名（可选，用于调用站点API）
	InsecureSkipVerify bool   // 跳过TLS证书验证（可选，仅用于证书过期等异常情况）
}

var appConfig *Config

// Load 加载配置（单例模式）
func Load() *Config {
	if appConfig != nil {
		return appConfig
	}

	appConfig = &Config{
		RedisAddr:          getEnv("REDIS_ADDR", ""),
		RedisPassword:      getEnv("REDIS_PASSWORD", ""),
		RedisDB:            getEnvInt("REDIS_DB", DefaultRedisDB),
		IYYUSign:           getEnv("IYUU_SIGN", ""),
		InsecureSkipVerify: getEnvBool("INSECURE_SKIP_VERIFY", false),
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

// getEnvInt 获取整数环境变量，如果不存在或解析失败则返回默认值
func getEnvInt(key string, defaultValue int) int {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}

	var result int
	if _, err := fmt.Sscanf(value, "%d", &result); err == nil {
		return result
	}

	return defaultValue
}

// getEnvBool 获取布尔环境变量，如果不存在或解析失败则返回默认值
func getEnvBool(key string, defaultValue bool) bool {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	// 支持 true/false, 1/0, yes/no
	return value == "true" || value == "1" || value == "yes"
}
