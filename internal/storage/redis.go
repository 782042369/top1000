package storage

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
	"top1000/internal/config"
	"top1000/internal/model"
)

const (
	dataKeySuffix = "data"
	dialTimeout   = 10 * time.Second
	readTimeout   = 5 * time.Second
	writeTimeout  = 5 * time.Second
	poolSize      = 3
	minIdleConns  = 1

	// Redis TTL 特殊返回值
	ttlKeyNotExist = -2 * time.Second // key 不存在（已过期删除）
	ttlKeyNoExpire = -1 * time.Second // key 存在但没有过期时间

	// 时间格式常量
	timeFormat = "2006-01-02 15:04:05" // 数据时间字段格式
)

var (
	redisClient *redis.Client
	isUpdating   bool
	updateMutex sync.Mutex
)

// InitRedis 连接Redis
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

	log.Println("✅ Redis连接成功")
	return nil
}

// CloseRedis 关闭Redis连接
func CloseRedis() error {
	if redisClient != nil {
		return redisClient.Close()
	}
	return nil
}

// SaveData 存储数据到Redis
func SaveData(data model.ProcessedData) error {
	cfg := config.Get()

	if err := data.Validate(); err != nil {
		log.Printf("❌ 数据验证失败，拒绝保存: %v", err)
		return fmt.Errorf("数据验证失败: %w", err)
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("序列化数据失败: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), writeTimeout)
	defer cancel()

	key := cfg.RedisKeyPrefix + dataKeySuffix
	// 不设置TTL，数据永久存储
	if err := redisClient.Set(ctx, key, jsonData, 0).Err(); err != nil {
		log.Printf("❌ 保存数据到Redis失败: %v", err)
		return fmt.Errorf("保存数据到Redis失败: %w", err)
	}

	log.Printf("✅ 数据已保存到Redis（永久存储，过期判断基于数据time字段）")
	return nil
}

// LoadData 从Redis读取数据
func LoadData() (*model.ProcessedData, error) {
	cfg := config.Get()
	key := cfg.RedisKeyPrefix + dataKeySuffix

	ctx, cancel := context.WithTimeout(context.Background(), readTimeout)
	defer cancel()

	jsonData, err := redisClient.Get(ctx, key).Bytes()
	if err != nil {
		if err == redis.Nil {
			return nil, fmt.Errorf("数据不存在")
		}
		return nil, fmt.Errorf("从Redis读取数据失败: %w", err)
	}

	var data model.ProcessedData
	if err := json.Unmarshal(jsonData, &data); err != nil {
		return nil, fmt.Errorf("解析JSON失败: %w", err)
	}

	log.Printf("✅ 从Redis加载数据成功（共 %d 条记录）", len(data.Items))
	return &data, nil
}

// IsDataExpired 检查数据是否过期（基于数据time字段）
func IsDataExpired() (bool, error) {
	// 读取数据
	data, err := LoadData()
	if err != nil {
		return true, nil // 数据不存在或读取失败，认为过期
	}

	// 解析时间字段
	dataTime, err := time.Parse(timeFormat, data.Time)
	if err != nil {
		log.Printf("⚠️ 解析数据时间失败: %v", err)
		return true, nil // 解析失败，认为过期，强制更新
	}

	// 计算时间差并判断
	cfg := config.Get()
	age := time.Now().Sub(dataTime)
	isExpired := age > cfg.DataExpireDuration

	// 统一日志输出
	logDataStatus(data.Time, age.Round(time.Minute), isExpired, cfg.DataExpireDuration)
	return isExpired, nil
}

// logDataStatus 记录数据状态日志
func logDataStatus(dataTime string, age time.Duration, isExpired bool, threshold time.Duration) {
	if isExpired {
		log.Printf("⚠️ 数据过期了（数据时间: %v, 距今: %v，阈值: %v）", dataTime, age, threshold)
	} else {
		log.Printf("✅ 数据还新鲜（数据时间: %v, 距今: %v）", dataTime, age)
	}
}

// DataExists 检查数据是否存在
func DataExists() (bool, error) {
	cfg := config.Get()
	key := cfg.RedisKeyPrefix + dataKeySuffix

	ctx, cancel := context.WithTimeout(context.Background(), readTimeout)
	defer cancel()

	exists, err := redisClient.Exists(ctx, key).Result()
	if err != nil {
		return false, fmt.Errorf("检查数据存在性失败: %w", err)
	}

	return exists > 0, nil
}

// Ping 测试Redis连接
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
	updateMutex.Lock()
	defer updateMutex.Unlock()
	return isUpdating
}

// SetUpdating 设置更新标记
func SetUpdating(updating bool) {
	updateMutex.Lock()
	defer updateMutex.Unlock()
	isUpdating = updating
}
