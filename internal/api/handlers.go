package api

import (
	"log"
	"sync"
	"time"
	"top1000/internal/crawler"
	"top1000/internal/model"
	"top1000/internal/storage"

	"github.com/gofiber/fiber/v2"
)

const (
	// maxUpdateWaitTime 最大等待更新时间
	maxUpdateWaitTime = 30 * time.Second
	// updateCheckInterval 更新检查间隔
	updateCheckInterval = 100 * time.Millisecond
)

var (
	// 内存缓存
	cacheData *model.ProcessedData
	// 读写锁，保护并发访问
	cacheMutex sync.RWMutex
	// 加载中的标记，防止重复加载
	loadingFlag bool
	// 加载完成通道，用于等待加载完成
	loadDone chan struct{}
)

// GetTop1000Data 提供Top1000数据的API接口
func GetTop1000Data(c *fiber.Ctx) error {
	// 1. 尝试从内存缓存读取
	if data, found := tryGetFromCache(); found {
		return c.JSON(data)
	}

	// 2. 检查数据状态（是否存在或过期）
	needsUpdate, err := checkDataStatus()
	if err != nil {
		log.Printf("⚠️ 检查数据状态失败: %v", err)
	}

	// 3. 如果需要更新，触发异步更新并等待
	if needsUpdate {
		if data, ok := waitForDataUpdate(c); ok {
			return c.JSON(data)
		}
	}

	// 4. 从Redis加载数据
	data, err := loadDataFromStorage()
	if err != nil {
		log.Printf("❌ 从存储加载数据失败: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "无法加载数据",
		})
	}

	// 5. 更新内存缓存
	updateMemoryCache(data)
	return c.JSON(data)
}

// tryGetFromCache 尝试从内存缓存读取数据
func tryGetFromCache() (*model.ProcessedData, bool) {
	cacheMutex.RLock()
	defer cacheMutex.RUnlock()

	if cacheData != nil {
		data := *cacheData
		return &data, true
	}

	// 如果正在加载，等待加载完成
	if loadingFlag && loadDone != nil {
		cacheMutex.RUnlock()
		<-loadDone
		cacheMutex.RLock()

		if cacheData != nil {
			data := *cacheData
			return &data, true
		}
	}

	return nil, false
}

// checkDataStatus 检查数据是否过期（需要更新则返回true）
func checkDataStatus() (bool, error) {
	// 检查数据是否存在
	exists, err := storage.DataExists()
	if err != nil {
		log.Printf("⚠️ 检查数据是否存在失败: %v", err)
		return true, err // 出错时默认需要更新
	}

	// 如果数据不存在，需要更新
	if !exists {
		return true, nil
	}

	// 检查数据是否过期
	isExpired, err := storage.IsDataExpired()
	if err != nil {
		log.Printf("⚠️ 检查数据是否过期失败: %v，将重新获取", err)
		return true, nil // 出错时默认需要更新
	}

	return isExpired, nil
}

// waitForDataUpdate 等待数据更新完成（最多等30秒，超时则返回旧数据）
func waitForDataUpdate(c *fiber.Ctx) (*model.ProcessedData, bool) {
	log.Println("⚠️ 数据不存在或已过期，触发实时更新...")

	// 异步触发更新
	go triggerDataUpdate()

	// 等待更新完成
	timeout := time.After(maxUpdateWaitTime)
	ticker := time.NewTicker(updateCheckInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			// 检查更新是否完成
			if !storage.IsUpdating() {
				if dataExists, _ := storage.DataExists(); dataExists {
					data, err := storage.LoadData()
					if err == nil && data != nil {
						updateMemoryCache(data)
						return data, true
					}
				}
			}
		case <-timeout:
			// 超时，尝试返回旧数据
			log.Println("⚠️ 等待数据更新超时，尝试返回旧数据")
			data, err := storage.LoadData()
			if err == nil && data != nil {
				updateMemoryCache(data)
				return data, true
			}
			// 实在没有数据，返回错误
			c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
				"error": "数据正在更新中，请稍后再试",
			})
			return nil, false
		}
	}
}

// triggerDataUpdate 触发数据更新（在goroutine中执行，避免阻塞主线程）
func triggerDataUpdate() {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("❌ 数据更新panic: %v", r)
		}
	}()

	if err := crawler.FetchData(); err != nil {
		log.Printf("❌ 实时更新失败: %v", err)
	} else {
		InvalidateCache()
		log.Println("✅ 实时更新成功，缓存已失效")
	}
}

// loadDataFromStorage 从Redis加载数据（有锁保护，确保并发安全）
func loadDataFromStorage() (*model.ProcessedData, error) {
	cacheMutex.Lock()
	defer cacheMutex.Unlock()

	// 双重检查（可能在等待锁期间已被其他请求加载）
	if cacheData != nil {
		data := *cacheData
		return &data, nil
	}

	// 设置加载标记
	loadingFlag = true
	loadDone = make(chan struct{})

	// 从 Redis 加载数据（需要临时释放锁，避免阻塞其他请求）
	cacheMutex.Unlock()
	data, err := storage.LoadData()
	cacheMutex.Lock()

	if err != nil {
		// 清除加载标记
		clearLoadingFlag()
		return nil, err
	}

	// 清除加载标记
	clearLoadingFlag()
	return data, nil
}

// updateMemoryCache 更新内存缓存（同时清除loading标记）
func updateMemoryCache(data *model.ProcessedData) {
	cacheMutex.Lock()
	defer cacheMutex.Unlock()

	cacheData = data
	if loadingFlag {
		loadingFlag = false
		if loadDone != nil {
			close(loadDone)
			loadDone = nil
		}
	}
}

// clearLoadingFlag 清除加载标记（错误时使用，调用前必须先加锁）
func clearLoadingFlag() {
	loadingFlag = false
	if loadDone != nil {
		close(loadDone)
		loadDone = nil
	}
}

// InvalidateCache 使缓存失效（数据更新后调用，强制下次从Redis重新加载）
func InvalidateCache() {
	cacheMutex.Lock()
	defer cacheMutex.Unlock()
	cacheData = nil
}
