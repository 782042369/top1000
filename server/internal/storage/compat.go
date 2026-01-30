package storage

import (
	"context"
	"fmt"
	"log"

	"github.com/redis/go-redis/v9"
	"top1000/internal/config"
	"top1000/internal/model"
)

var (
	defaultStore      DataStore
	defaultSitesStore SitesStore
	defaultLock       UpdateLock
	redisClient       *redis.Client
)

// InitRedis 初始化 Redis 连接
func InitRedis() error {
	cfg := config.Get()
	log.Printf("正在连接Redis: %s (DB: %d)", cfg.RedisAddr, cfg.RedisDB)

	redisClient = redis.NewClient(&redis.Options{
		Addr:         cfg.RedisAddr,
		Password:     cfg.RedisPassword,
		DB:           cfg.RedisDB,
		DialTimeout:  dialTimeout,
		ReadTimeout:  readTimeout,
		WriteTimeout: writeTimeout,
		PoolSize:     poolSize,
		MinIdleConns: minIdleConns,
	})

	ctx, cancel := context.WithTimeout(context.Background(), dialTimeout)
	defer cancel()

	if err := redisClient.Ping(ctx).Err(); err != nil {
		log.Printf("Redis连接失败: %v", err)
		return fmt.Errorf("Redis连接失败: %w", err)
	}

	redisStore := NewRedisStore(redisClient)
	defaultStore = redisStore.AsDataStore()
	defaultSitesStore = redisStore.AsSitesStore()
	defaultLock = redisStore.AsUpdateLock()

	log.Println("Redis连接成功")
	return nil
}

// CloseRedis 关闭 Redis 连接
func CloseRedis() error {
	if redisClient != nil {
		return redisClient.Close()
	}
	return nil
}

// GetDefaultStore 获取默认数据存储实例
func GetDefaultStore() DataStore {
	return defaultStore
}

// GetDefaultSitesStore 获取默认站点存储实例
func GetDefaultSitesStore() SitesStore {
	return defaultSitesStore
}

// GetDefaultLock 获取默认更新锁实例
func GetDefaultLock() UpdateLock {
	return defaultLock
}

// ===== 兼容函数 =====

// SaveData 存储数据到 Redis（使用默认超时）
func SaveData(data model.ProcessedData) error {
	ctx, cancel := context.WithTimeout(context.Background(), writeTimeout)
	defer cancel()
	return SaveDataWithContext(ctx, data)
}

// SaveDataWithContext 存储数据到 Redis
func SaveDataWithContext(ctx context.Context, data model.ProcessedData) error {
	return defaultStore.SaveData(ctx, data)
}

// LoadData 从 Redis 读取数据（使用默认超时）
func LoadData() (*model.ProcessedData, error) {
	ctx, cancel := context.WithTimeout(context.Background(), readTimeout)
	defer cancel()
	return LoadDataWithContext(ctx)
}

// LoadDataWithContext 从 Redis 读取数据
func LoadDataWithContext(ctx context.Context) (*model.ProcessedData, error) {
	return defaultStore.LoadData(ctx)
}

// IsDataExpired 检查数据是否过期（使用默认超时）
func IsDataExpired() (bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), readTimeout)
	defer cancel()
	return IsDataExpiredWithContext(ctx)
}

// IsDataExpiredWithContext 检查数据是否过期
func IsDataExpiredWithContext(ctx context.Context) (bool, error) {
	return defaultStore.IsDataExpired(ctx)
}

// DataExists 检查数据是否存在（使用默认超时）
func DataExists() (bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), readTimeout)
	defer cancel()
	return DataExistsWithContext(ctx)
}

// DataExistsWithContext 检查数据是否存在
func DataExistsWithContext(ctx context.Context) (bool, error) {
	return defaultStore.DataExists(ctx)
}

// Ping 测试 Redis 连接
func Ping() error {
	if redisClient == nil {
		return fmt.Errorf("Redis客户端未初始化")
	}
	ctx, cancel := context.WithTimeout(context.Background(), dialTimeout)
	defer cancel()
	return redisClient.Ping(ctx).Err()
}

// IsUpdating 检查是否正在更新
func IsUpdating() bool {
	return defaultLock.IsUpdating()
}

// SetUpdating 设置更新标记
func SetUpdating(updating bool) {
	defaultLock.SetUpdating(updating)
}

// SaveSitesData 存储站点数据到 Redis（使用默认超时）
func SaveSitesData(data any) error {
	ctx, cancel := context.WithTimeout(context.Background(), writeTimeout)
	defer cancel()
	return SaveSitesDataWithContext(ctx, data)
}

// SaveSitesDataWithContext 存储站点数据到 Redis
func SaveSitesDataWithContext(ctx context.Context, data any) error {
	return defaultSitesStore.SaveSitesData(ctx, data)
}

// LoadSitesData 从 Redis 读取站点数据（使用默认超时）
func LoadSitesData() (any, error) {
	ctx, cancel := context.WithTimeout(context.Background(), readTimeout)
	defer cancel()
	return LoadSitesDataWithContext(ctx)
}

// LoadSitesDataWithContext 从 Redis 读取站点数据
func LoadSitesDataWithContext(ctx context.Context) (any, error) {
	return defaultSitesStore.LoadSitesData(ctx)
}

// SitesDataExists 检查站点数据是否存在（使用默认超时）
func SitesDataExists() (bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), readTimeout)
	defer cancel()
	return SitesDataExistsWithContext(ctx)
}

// SitesDataExistsWithContext 检查站点数据是否存在
func SitesDataExistsWithContext(ctx context.Context) (bool, error) {
	return defaultSitesStore.SitesDataExists(ctx)
}

// IsSitesUpdating 检查是否正在更新站点数据
func IsSitesUpdating() bool {
	return defaultLock.IsSitesUpdating()
}

// SetSitesUpdating 设置站点数据更新标记
func SetSitesUpdating(updating bool) {
	defaultLock.SetSitesUpdating(updating)
}
