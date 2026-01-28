package api

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"

	"github.com/gofiber/fiber/v2"
	"top1000/internal/config"
	"top1000/internal/crawler"
	"top1000/internal/model"
	"top1000/internal/storage"
)

// Handler API 处理器（依赖注入模式）
// 组合多个接口，遵循"组合优于继承"原则
type Handler struct {
	store      storage.DataStore  // 数据存储接口
	sitesStore storage.SitesStore // 站点存储接口
	lock       storage.UpdateLock // 更新锁接口
	crawler    Crawler            // 爬虫接口
	httpClient *http.Client       // HTTP 客户端
}

// Crawler 爬虫接口（小而专注）
// 定义爬虫的核心能力，方便测试和替换实现
type Crawler interface {
	// FetchTop1000WithContext 带 context 的数据爬取
	FetchTop1000WithContext(ctx context.Context) (*model.ProcessedData, error)
}

// NewHandler 创建 Handler 实例（依赖注入）
// 接收接口类型，遵循"依赖倒置原则"
func NewHandler(store storage.DataStore, sitesStore storage.SitesStore, lock storage.UpdateLock) *Handler {
	return &Handler{
		store:      store,
		sitesStore: sitesStore,
		lock:       lock,
		crawler:    &defaultCrawler{}, // 使用默认爬虫实现
		httpClient: &http.Client{Timeout: defaultHTTPClientTimeout},
	}
}

// defaultCrawler 默认爬虫实现（实现 Crawler 接口）
type defaultCrawler struct{}

// FetchTop1000WithContext 调用底层爬虫
func (d *defaultCrawler) FetchTop1000WithContext(ctx context.Context) (*model.ProcessedData, error) {
	return crawler.FetchTop1000WithContext(ctx)
}

// RegisterRoutes 注册路由
func (h *Handler) RegisterRoutes(app *fiber.App) {
	app.Get("/top1000.json", h.GetTop1000Data)
	app.Get("/sites.json", h.GetSitesData)
}

// ===== 以下改为 Handler 的方法 =====

// GetTop1000Data 提供Top1000数据的API接口
// @Summary 获取Top1000站点数据
// @Description 获取Top1000站点列表数据，数据会自动更新（24小时过期）
// @Tags Top1000
// @Accept json
// @Produce json
// @Success 200 {object} model.ProcessedData
// @Failure 500 {object} map[string]string "error": "无法加载数据"
// @Router /top1000.json [get]
func (h *Handler) GetTop1000Data(c *fiber.Ctx) error {
	// 从Fiber的context提取标准的context.Context
	// 设置超时保护（如果客户端没设置超时）
	ctx, cancel := context.WithTimeout(c.Context(), defaultAPITimeout)
	defer cancel()

	// 检查数据是否需要更新
	if h.shouldUpdateData(ctx) {
		if err := h.refreshData(ctx); err != nil {
			log.Printf("[%s] ⚠️ 刷新数据失败: %v", dataUpdateLogPrefix, err)
			// 容错：继续尝试读取旧数据
		}
	}

	// 从存储读取数据并返回（传递context）
	data, err := h.store.LoadData(ctx)
	if err != nil {
		log.Printf("[%s] ❌ 加载数据失败: %v", dataUpdateLogPrefix, err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "无法加载数据",
		})
	}

	return c.JSON(data)
}

// shouldUpdateData 检查数据是否需要更新
func (h *Handler) shouldUpdateData(ctx context.Context) bool {
	// 数据不存在或出错时,需要更新
	exists, err := h.store.DataExists(ctx)
	if err != nil || !exists {
		return true
	}

	// 数据过期时,需要更新
	isExpired, err := h.store.IsDataExpired(ctx)
	return err != nil || isExpired
}

// refreshData 刷新数据（带容错机制）
// 返回 error 让调用者知道刷新是否成功
func (h *Handler) refreshData(ctx context.Context) error {
	// 防止并发更新
	if h.lock.IsUpdating() {
		log.Printf("[%s] ⏸️ 正在更新中，跳过", dataUpdateLogPrefix)
		return nil
	}

	h.lock.SetUpdating(true)
	defer h.lock.SetUpdating(false)

	// 保存旧数据用于容错（传递context）
	oldData, err := h.store.LoadData(ctx)
	if err != nil {
		log.Printf("[%s] ⚠️ 加载旧数据失败: %v", dataUpdateLogPrefix, err)
		// 容错：旧数据不存在时继续爬取新数据
	}

	log.Printf("[%s] 🔍 开始爬取新数据...", dataUpdateLogPrefix)
	newData, err := h.crawler.FetchTop1000WithContext(ctx)
	if err != nil {
		// 爬取失败，如果有旧数据则使用旧数据（容错）
		if oldData != nil {
			log.Printf("[%s] ✅ 爬取失败，使用旧数据: %v", dataUpdateLogPrefix, err)
			return err
		}
		log.Printf("[%s] ❌ 爬取失败且无旧数据: %v", dataUpdateLogPrefix, err)
		return err
	}

	if err := h.store.SaveData(ctx, *newData); err != nil {
		log.Printf("[%s] ❌ 保存数据失败: %v", dataUpdateLogPrefix, err)
		return err
	}

	log.Printf("[%s] ✅ 数据更新成功（%d 条）", dataUpdateLogPrefix, len(newData.Items))
	return nil
}

// GetSitesData 提供IYUU站点数据的API接口
// @Summary 获取IYUU站点列表
// @Description 获取IYUU站点列表数据（需要配置IYUU_SIGN环境变量）
// @Tags Sites
// @Accept json
// @Produce json
// @Success 200 {object} interface{} "站点列表数据"
// @Failure 502 {object} map[string]string "error": "未配置IYUU_SIGN环境变量"
// @Failure 500 {object} map[string]string "error": "无法加载站点数据"
// @Router /sites.json [get]
func (h *Handler) GetSitesData(c *fiber.Ctx) error {
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
	if h.shouldUpdateSitesData(ctx) {
		if err := h.refreshSitesData(ctx, cfg.IYYUSign); err != nil {
			log.Printf("[%s] ⚠️ 刷新站点数据失败: %v", sitesUpdateLogPrefix, err)
			// 容错：继续尝试读取旧数据
		}
	}

	// 从存储读取数据并返回
	data, err := h.sitesStore.LoadSitesData(ctx)
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
func (h *Handler) shouldUpdateSitesData(ctx context.Context) bool {
	// 数据不存在时，需要更新
	exists, err := h.sitesStore.SitesDataExists(ctx)
	if err != nil || !exists {
		return true
	}
	return false
}

// refreshSitesData 刷新站点数据（带容错机制）
// 返回 error 让调用者知道刷新是否成功
func (h *Handler) refreshSitesData(ctx context.Context, sign string) error {
	// 防止并发更新
	if h.lock.IsSitesUpdating() {
		log.Printf("[%s] ⏸️ 正在更新中，跳过", sitesUpdateLogPrefix)
		return nil
	}

	h.lock.SetSitesUpdating(true)
	defer h.lock.SetSitesUpdating(false)

	log.Printf("[%s] 🔍 开始获取站点数据...", sitesUpdateLogPrefix)

	// 构建API URL（使用net/url包，更安全规范）
	apiURL, err := url.Parse("https://api.iyuu.cn/index.php")
	if err != nil {
		log.Printf("[%s] ❌ 解析基础URL失败: %v", sitesUpdateLogPrefix, err)
		return fmt.Errorf("解析基础URL失败: %w", err)
	}
	params := url.Values{}
	params.Add("service", "App.Api.Sites")
	params.Add("sign", sign)
	params.Add("version", "2.0.0")
	apiURL.RawQuery = params.Encode()

	// 创建HTTP客户端（从 context 提取超时时间）
	client := getHTTPClient(ctx)

	// 发送GET请求
	resp, err := client.Get(apiURL.String())
	if err != nil {
		log.Printf("[%s] ❌ 请求失败: %v", sitesUpdateLogPrefix, err)
		return fmt.Errorf("请求失败: %w", err)
	}
	defer resp.Body.Close()

	// 读取响应体
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("[%s] ❌ 读取响应失败: %v", sitesUpdateLogPrefix, err)
		return fmt.Errorf("读取响应失败: %w", err)
	}

	// 解析JSON
	var result interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		log.Printf("[%s] ❌ 解析JSON失败: %v", sitesUpdateLogPrefix, err)
		return fmt.Errorf("解析JSON失败: %w", err)
	}

	// 保存到存储（24小时TTL）
	if err := h.sitesStore.SaveSitesData(ctx, result); err != nil {
		log.Printf("[%s] ❌ 保存数据失败: %v", sitesUpdateLogPrefix, err)
		return fmt.Errorf("保存数据失败: %w", err)
	}

	log.Printf("[%s] ✅ 站点数据更新成功", sitesUpdateLogPrefix)
	return nil
}

