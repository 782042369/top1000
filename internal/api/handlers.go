package api

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"net/url"
	"time"
	"top1000/internal/config"
	"top1000/internal/crawler"
	"top1000/internal/storage"

	"github.com/gofiber/fiber/v2"
)

const (
	dataUpdateLogPrefix      = "📊 Top1000"
	sitesUpdateLogPrefix     = "🔗 Sites"
	defaultAPITimeout        = 15 * time.Second // API默认超时时间
	defaultHTTPClientTimeout = 5 * time.Second  // HTTP客户端超时时间
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

// GetSitesData 提供IYUU站点数据的API接口
func GetSitesData(c *fiber.Ctx) error {
	cfg := config.Get()

	// 检查是否配置了IYUU_SIGN
	if cfg.IYYUSign == "" {
		return c.Status(fiber.StatusBadGateway).JSON(fiber.Map{
			"error": "未配置IYUU_SIGN环境变量",
		})
	}

	// 从Fiber的context提取标准的context.Context
	ctx, cancel := context.WithTimeout(c.Context(), defaultAPITimeout)
	defer cancel()

	// 检查数据是否存在，不存在或正在更新时触发更新
	if shouldUpdateSitesData(ctx) {
		refreshSitesData(ctx, cfg.IYYUSign)
	}

	// 从Redis读取数据并返回
	data, err := storage.LoadSitesDataWithContext(ctx)
	if err != nil {
		log.Printf("[%s] ❌ 加载站点数据失败: %v", sitesUpdateLogPrefix, err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "无法加载站点数据",
		})
	}

	// 设置响应头
	c.Set("Content-Type", "application/json; charset=utf-8")
	c.Set("Cache-Control", "public, max-age=3600") // 缓存1小时

	return c.JSON(data)
}

// shouldUpdateSitesData 检查站点数据是否需要更新
func shouldUpdateSitesData(ctx context.Context) bool {
	// 数据不存在时，需要更新
	exists, err := storage.SitesDataExistsWithContext(ctx)
	if err != nil || !exists {
		return true
	}
	return false
}

// refreshSitesData 刷新站点数据（带容错机制）
func refreshSitesData(ctx context.Context, sign string) {
	// 防止并发更新
	if storage.IsSitesUpdating() {
		log.Printf("[%s] ⏸️ 正在更新中，跳过", sitesUpdateLogPrefix)
		return
	}

	storage.SetSitesUpdating(true)
	defer storage.SetSitesUpdating(false)

	log.Printf("[%s] 🔍 开始获取站点数据...", sitesUpdateLogPrefix)

	// 构建API URL（使用net/url包，更安全规范）
	apiURL, err := url.Parse("https://api.iyuu.cn/index.php")
	if err != nil {
		log.Printf("[%s] ❌ 解析基础URL失败: %v", sitesUpdateLogPrefix, err)
		return
	}
	params := url.Values{}
	params.Add("service", "App.Api.Sites")
	params.Add("sign", sign)
	params.Add("version", "2.0.0")
	apiURL.RawQuery = params.Encode()

	// 创建HTTP客户端（使用配置的超时时间）
	client := &http.Client{
		Timeout: defaultHTTPClientTimeout,
	}

	// 发送GET请求
	resp, err := client.Get(apiURL.String())
	if err != nil {
		log.Printf("[%s] ❌ 请求失败: %v", sitesUpdateLogPrefix, err)
		return
	}
	defer resp.Body.Close()

	// 读取响应体
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("[%s] ❌ 读取响应失败: %v", sitesUpdateLogPrefix, err)
		return
	}

	// 解析JSON
	var result interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		log.Printf("[%s] ❌ 解析JSON失败: %v", sitesUpdateLogPrefix, err)
		return
	}

	// 保存到Redis（24小时TTL）
	if err := storage.SaveSitesDataWithContext(ctx, result); err != nil {
		log.Printf("[%s] ❌ 保存数据失败: %v", sitesUpdateLogPrefix, err)
		return
	}

	log.Printf("[%s] ✅ 站点数据更新成功", sitesUpdateLogPrefix)
}
