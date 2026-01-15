package api

import (
	"context"
	"log"
	"time"
	"top1000/internal/crawler"
	"top1000/internal/storage"

	"github.com/gofiber/fiber/v2"
)

const (
	dataUpdateLogPrefix = "📊 Top1000"
	defaultAPITimeout   = 15 * time.Second // API默认超时时间
)

// GetTop1000Data 提供Top1000数据的API接口
func GetTop1000Data(c *fiber.Ctx) error {
	// 从Fiber的context提取标准的context.Context
	// 设置超时保护（如果客户端没设置超时）
	ctx, cancel := context.WithTimeout(c.Context(), defaultAPITimeout)
	defer cancel()

	// 检查数据是否需要更新
	if shouldUpdateData(ctx) {
		refreshData(ctx)
	}

	// 从Redis读取数据并返回（传递context）
	data, err := storage.LoadDataWithContext(ctx)
	if err != nil {
		log.Printf("[%s] ❌ 加载数据失败: %v", dataUpdateLogPrefix, err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "无法加载数据",
		})
	}

	return c.JSON(data)
}

// shouldUpdateData 检查数据是否需要更新
func shouldUpdateData(ctx context.Context) bool {
	// 数据不存在或出错时,需要更新
	exists, err := storage.DataExistsWithContext(ctx)
	if err != nil || !exists {
		return true
	}

	// 数据过期时,需要更新
	isExpired, err := storage.IsDataExpiredWithContext(ctx)
	return err != nil || isExpired
}

// refreshData 刷新数据（带容错机制）
func refreshData(ctx context.Context) {
	// 防止并发更新
	if storage.IsUpdating() {
		log.Printf("[%s] ⏸️ 正在更新中，跳过", dataUpdateLogPrefix)
		return
	}

	storage.SetUpdating(true)
	defer storage.SetUpdating(false)

	// 保存旧数据用于容错（传递context）
	oldData, _ := storage.LoadDataWithContext(ctx)

	log.Printf("[%s] 🔍 开始爬取新数据...", dataUpdateLogPrefix)
	newData, err := crawler.FetchTop1000WithContext(ctx)
	if err != nil {
		// 爬取失败，如果有旧数据则使用旧数据（容错）
		if oldData != nil {
			log.Printf("[%s] ✅ 爬取失败，使用旧数据: %v", dataUpdateLogPrefix, err)
			return
		}
		log.Printf("[%s] ❌ 爬取失败且无旧数据: %v", dataUpdateLogPrefix, err)
		return
	}

	if err := storage.SaveDataWithContext(ctx, *newData); err != nil {
		log.Printf("[%s] ❌ 保存数据失败: %v", dataUpdateLogPrefix, err)
		return
	}

	log.Printf("[%s] ✅ 数据更新成功（%d 条）", dataUpdateLogPrefix, len(newData.Items))
}
