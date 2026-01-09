package storage

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/redis/go-redis/v9"
	"top1000/internal/config"
	"top1000/internal/model"
)

var (
	// Redis客户端
	redisClient *redis.Client
	// 上下文，Redis操作需要使用
	ctx = context.Background()
	// 更新标记，防止并发更新
	isUpdating bool
)

// InitRedis 连接Redis，连接失败则无法运行
func InitRedis() error {
	cfg := config.Get()

	log.Printf("正在连接Redis: %s (DB: %d)", cfg.RedisAddr, cfg.RedisDB)

	// 创建Redis客户端，配置已写定
	redisClient = redis.NewClient(&redis.Options{
		Addr:         cfg.RedisAddr,      // Redis地址，从配置文件读取
		Password:     cfg.RedisPassword,  // 密码，禁止硬编码
		DB:           cfg.RedisDB,        // 数据库编号，默认0
		DialTimeout:  5 * time.Second,    // 连接超时5秒
		ReadTimeout:  3 * time.Second,    // 读超时3秒
		WriteTimeout: 3 * time.Second,    // 写超时3秒
		PoolSize:     10,                 // 连接池10个连接
		MinIdleConns: 5,                  // 保持5个空闲连接
	})

	// Ping测试连接是否可用
	if err := redisClient.Ping(ctx).Err(); err != nil {
		log.Printf("❌ Redis连接失败: %v", err)
		return fmt.Errorf("Redis连接失败: %v", err)
	}

	log.Println("✅ Redis连接成功")
	return nil
}

// CloseRedis 关闭Redis连接，程序退出时调用
func CloseRedis() error {
	if redisClient != nil {
		return redisClient.Close()
	}
	return nil
}

// SaveData 存数据到Redis，存储前先检查数据正确性
func SaveData(data model.ProcessedData) error {
	cfg := config.Get()

	// 先验证数据，避免存储错误数据
	if err := data.Validate(); err != nil {
		log.Printf("❌ 数据验证失败，拒绝保存: %v", err)
		return fmt.Errorf("数据验证失败: %v", err)
	}

	// 转成JSON
	jsonData, err := json.Marshal(data)
	if err != nil {
		log.Printf("❌ 序列化数据失败: %v", err)
		return fmt.Errorf("序列化数据失败: %v", err)
	}

	// Redis的key
	key := cfg.RedisKeyPrefix + "data"

	// 存进去，TTL设成2倍的更新间隔（48小时）
	expiration := 2 * cfg.DataExpireDuration
	if err := redisClient.Set(ctx, key, jsonData, expiration).Err(); err != nil {
		log.Printf("❌ 保存数据到Redis失败: %v", err)
		return fmt.Errorf("保存数据到Redis失败: %v", err)
	}

	log.Printf("✅ 数据已保存到Redis（过期时间: %v）", expiration)
	return nil
}

// LoadData 从Redis读数据，数据不存在则返回nil
func LoadData() (*model.ProcessedData, error) {
	cfg := config.Get()

	// Redis的key
	key := cfg.RedisKeyPrefix + "data"

	// 从Redis读
	jsonData, err := redisClient.Get(ctx, key).Bytes()
	if err != nil {
		if err == redis.Nil {
			return nil, fmt.Errorf("数据不存在")
		}
		log.Printf("❌ 从Redis读取数据失败: %v", err)
		return nil, fmt.Errorf("从Redis读取数据失败: %v", err)
	}

	// 解析JSON
	var data model.ProcessedData
	if err := json.Unmarshal(jsonData, &data); err != nil {
		log.Printf("❌ 解析JSON失败: %v", err)
		return nil, fmt.Errorf("解析JSON失败: %v", err)
	}

	log.Printf("✅ 从Redis加载数据成功（共 %d 条记录）", len(data.Items))
	return &data, nil
}

// IsDataExpired 检查数据是否过期，TTL小于24小时即算过期
func IsDataExpired() (bool, error) {
	cfg := config.Get()

	// Redis的key
	key := cfg.RedisKeyPrefix + "data"

	// 获取TTL
	ttl, err := redisClient.TTL(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			// key不存在，肯定过期了
			return true, nil
		}
		log.Printf("❌ 获取TTL失败: %v", err)
		return false, fmt.Errorf("获取TTL失败: %v", err)
	}

	// TTL小于阈值就当过期了
	isExpired := ttl < cfg.DataExpireDuration

	if isExpired {
		log.Printf("⚠️ 数据过期了（剩余时间: %v，阈值: %v）", ttl, cfg.DataExpireDuration)
	} else {
		log.Printf("✅ 数据还新鲜（剩余时间: %v）", ttl)
	}

	return isExpired, nil
}

// DataExists 检查Redis中是否存在数据
func DataExists() (bool, error) {
	cfg := config.Get()

	// Redis的key
	key := cfg.RedisKeyPrefix + "data"

	// 检查key在不在
	exists, err := redisClient.Exists(ctx, key).Result()
	if err != nil {
		log.Printf("❌ 检查数据存在性失败: %v", err)
		return false, fmt.Errorf("检查数据存在性失败: %v", err)
	}

	return exists > 0, nil
}

// GetTTL 获取数据剩余存活时间
func GetTTL() (time.Duration, error) {
	cfg := config.Get()

	// Redis的key
	key := cfg.RedisKeyPrefix + "data"

	// 获取TTL
	ttl, err := redisClient.TTL(ctx, key).Result()
	if err != nil {
		log.Printf("❌ 获取TTL失败: %v", err)
		return 0, fmt.Errorf("获取TTL失败: %v", err)
	}

	return ttl, nil
}

// Ping 测试Redis连接是否存活
func Ping() error {
	if redisClient == nil {
		return fmt.Errorf("Redis客户端未初始化")
	}
	return redisClient.Ping(ctx).Err()
}

// IsUpdating 检查是否正在更新数据
func IsUpdating() bool {
	return isUpdating
}

// SetUpdating 设置更新标记，防止并发更新操作
func SetUpdating(updating bool) {
	isUpdating = updating
}
