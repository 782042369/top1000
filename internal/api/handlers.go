package api

import (
	"fmt"
	"log"
	"sync"
	"time"
	"top1000/internal/crawler"
	"top1000/internal/model"
	"top1000/internal/storage"

	"github.com/gofiber/fiber/v2"
)

const (
	maxUpdateWaitTime   = 10 * time.Second // 小项目不需要等太久
	updateCheckInterval = 200 * time.Millisecond // 降低检查频率
)

var (
	cacheData  *model.ProcessedData
	cacheMutex sync.RWMutex
	loadingFlag bool
	loadDone    chan struct{}
)

// GetTop1000Data 提供Top1000数据的API接口
func GetTop1000Data(c *fiber.Ctx) error {
	if data, found := tryGetFromCache(); found {
		return c.JSON(data)
	}

	needsUpdate, err := checkDataStatus()
	if err != nil {
		log.Printf("⚠️ 检查数据状态失败: %v", err)
	}

	if needsUpdate {
		if data, ok := waitForDataUpdate(c); ok {
			return c.JSON(data)
		}
	}

	data, err := loadDataFromStorage()
	if err != nil {
		log.Printf("❌ 从存储加载数据失败: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "无法加载数据",
		})
	}

	updateMemoryCache(data)
	return c.JSON(data)
}

// tryGetFromCache 尝试从内存缓存读取数据
func tryGetFromCache() (*model.ProcessedData, bool) {
	cacheMutex.RLock()
	defer cacheMutex.RUnlock()

	if cacheData != nil {
		return cacheData, true
	}

	if loadingFlag && loadDone != nil {
		cacheMutex.RUnlock()
		<-loadDone
		cacheMutex.RLock()

		if cacheData != nil {
			return cacheData, true
		}
	}

	return nil, false
}

// checkDataStatus 检查数据是否过期
func checkDataStatus() (bool, error) {
	exists, err := storage.DataExists()
	if err != nil {
		return true, fmt.Errorf("检查数据存在性失败: %w", err)
	}

	if !exists {
		return true, nil
	}

	isExpired, err := storage.IsDataExpired()
	if err != nil {
		return true, fmt.Errorf("检查数据过期失败: %w", err)
	}

	return isExpired, nil
}

// waitForDataUpdate 等待数据更新完成
func waitForDataUpdate(c *fiber.Ctx) (*model.ProcessedData, bool) {
	log.Println("⚠️ 数据不存在或已过期，触发实时更新...")

	go triggerDataUpdate()

	timeout := time.After(maxUpdateWaitTime)
	ticker := time.NewTicker(updateCheckInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if data := tryLoadAndUpdate(); data != nil {
				return data, true
			}
		case <-timeout:
			log.Println("⚠️ 等待数据更新超时，尝试返回旧数据")
			if data := tryLoadAndUpdate(); data != nil {
				log.Println("✅ 返回旧数据成功")
				return data, true
			}
			log.Println("❌ 无法加载旧数据，返回错误")
			c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
				"error": "数据加载失败，请稍后再试",
			})
			return nil, false
		}
	}
}

// tryLoadAndUpdate 尝试加载数据并更新缓存
func tryLoadAndUpdate() *model.ProcessedData {
	// 检查是否还在更新
	if storage.IsUpdating() {
		return nil
	}

	// 检查数据是否存在
	dataExists, err := storage.DataExists()
	if err != nil {
		log.Printf("⚠️ 检查数据存在性失败: %v", err)
		return nil
	}
	if !dataExists {
		return nil
	}

	// 加载数据
	data, err := storage.LoadData()
	if err != nil || data == nil {
		return nil
	}

	// 更新缓存并返回
	updateMemoryCache(data)
	return data
}

// triggerDataUpdate 触发数据更新
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

// loadDataFromStorage 从存储加载数据
func loadDataFromStorage() (*model.ProcessedData, error) {
	cacheMutex.Lock()
	defer cacheMutex.Unlock()

	if cacheData != nil {
		return cacheData, nil
	}

	loadingFlag = true
	loadDone = make(chan struct{})

	cacheMutex.Unlock()
	data, err := storage.LoadData()
	cacheMutex.Lock()

	if err != nil {
		clearLoadingFlag()
		return nil, err
	}

	clearLoadingFlag()
	return data, nil
}

// updateMemoryCache 更新内存缓存
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

// clearLoadingFlag 清除加载标记
func clearLoadingFlag() {
	loadingFlag = false
	if loadDone != nil {
		close(loadDone)
		loadDone = nil
	}
}

// InvalidateCache 使缓存失效
func InvalidateCache() {
	cacheMutex.Lock()
	defer cacheMutex.Unlock()
	cacheData = nil
}
