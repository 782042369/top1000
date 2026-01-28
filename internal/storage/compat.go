package storage

import (
	"context"
	"fmt"
	"log"

	"github.com/redis/go-redis/v9"
	"top1000/internal/config"
	"top1000/internal/model"
)

// 全局变量（向后兼容，新代码建议使用依赖注入）
var (
	defaultStore   DataStore
	defaultSitesStore SitesStore
	defaultLock    UpdateLock
	redisClient    *redis.Client
)

// InitRedis 初始化 Redis 连接（向后兼容）
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
		log.Printf("❌ Redis连接失败: %v", err)
		return fmt.Errorf("Redis连接失败: %w", err)
	}

	// 初始化默认存储实例
	redisStore := &RedisStore{client: redisClient}
	defaultStore = redisStore
	defaultSitesStore = redisStore
	defaultLock = redisStore

	log.Println("✅ Redis连接成功")
	return nil
}

// CloseRedis 关闭 Redis 连接
func CloseRedis() error {
	if redisClient != nil {
		return redisClient.Close()
	}
	return nil
}

// ===== 向后兼容函数（委托给接口） =====

// SaveData 存储数据到 Redis（向后兼容，使用默认超时）
func SaveData(data model.ProcessedData) error {
	ctx, cancel := context.WithTimeout(context.Background(), writeTimeout)
	defer cancel()
	return defaultStore.SaveData(ctx, data)
}

// SaveDataWithContext 存储数据到 Redis（支持外部传入 context）
func SaveDataWithContext(ctx context.Context, data model.ProcessedData) error {
	return defaultStore.SaveData(ctx, data)
}

// LoadData 从 Redis 读取数据（向后兼容，使用默认超时）
func LoadData() (*model.ProcessedData, error) {
	ctx, cancel := context.WithTimeout(context.Background(), readTimeout)
	defer cancel()
	return defaultStore.LoadData(ctx)
}

// LoadDataWithContext 从 Redis 读取数据（支持外部传入 context）
func LoadDataWithContext(ctx context.Context) (*model.ProcessedData, error) {
	return defaultStore.LoadData(ctx)
}

// IsDataExpired 检查数据是否过期（向后兼容）
func IsDataExpired() (bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), readTimeout)
	defer cancel()
	return defaultStore.IsDataExpired(ctx)
}

// IsDataExpiredWithContext 检查数据是否过期（支持外部传入 context）
func IsDataExpiredWithContext(ctx context.Context) (bool, error) {
	return defaultStore.IsDataExpired(ctx)
}

// DataExists 检查数据是否存在（向后兼容）
func DataExists() (bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), readTimeout)
	defer cancel()
	return defaultStore.DataExists(ctx)
}

// DataExistsWithContext 检查数据是否存在（支持外部传入 context）
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

// SaveSitesData 存储站点数据到 Redis（向后兼容）
func SaveSitesData(data interface{}) error {
	ctx, cancel := context.WithTimeout(context.Background(), writeTimeout)
	defer cancel()
	return defaultSitesStore.SaveSitesData(ctx, data)
}

// SaveSitesDataWithContext 存储站点数据到 Redis（支持外部传入 context）
func SaveSitesDataWithContext(ctx context.Context, data interface{}) error {
	return defaultSitesStore.SaveSitesData(ctx, data)
}

// LoadSitesData 从 Redis 读取站点数据
func LoadSitesData() (interface{}, error) {
	ctx, cancel := context.WithTimeout(context.Background(), readTimeout)
	defer cancel()
	return defaultSitesStore.LoadSitesData(ctx)
}

// LoadSitesDataWithContext 从 Redis 读取站点数据（支持外部传入 context）
func LoadSitesDataWithContext(ctx context.Context) (interface{}, error) {
	return defaultSitesStore.LoadSitesData(ctx)
}

// SitesDataExists 检查站点数据是否存在
func SitesDataExists() (bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), readTimeout)
	defer cancel()
	return defaultSitesStore.SitesDataExists(ctx)
}

// SitesDataExistsWithContext 检查站点数据是否存在（支持外部传入 context）
func SitesDataExistsWithContext(ctx context.Context) (bool, error) {
	return defaultSitesStore.SitesDataExists(ctx)
}

// IsSitesUpdating 检查是否正在更新站点数据
func IsSitesUpdating() bool {
	if rs, ok := defaultStore.(*RedisStore); ok {
		return rs.IsSitesUpdating()
	}
	return false
}

// SetSitesUpdating 设置站点数据更新标记
func SetSitesUpdating(updating bool) {
	if rs, ok := defaultStore.(*RedisStore); ok {
		rs.SetSitesUpdating(updating)
	}
}
