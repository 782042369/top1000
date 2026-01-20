package crawler

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"regexp"
	"strings"
	"sync"
	"time"
	"top1000/internal/config"
	"top1000/internal/model"
	"top1000/internal/storage"
)

const (
	logPrefix       = "🔍 爬虫"
	httpTimeout     = 10 * time.Second
	maxRetries      = 1
	retryInterval   = 1 * time.Second
	linesPerItem    = 3
	timeLineIndex   = 0
	dataStartLine   = 2
	timePrefix      = "create time "
	timeSuffix      = " by "
	fieldSeparator  = "："
	sitePattern     = `站名：(.*?) 【ID：(\d+)】`
)

var (
	siteRegex = regexp.MustCompile(sitePattern)
	taskMutex sync.Mutex
)

// FetchTop1000 从IYUU获取数据并返回（向后兼容，使用默认超时）
func FetchTop1000() (*model.ProcessedData, error) {
	ctx, cancel := context.WithTimeout(context.Background(), httpTimeout)
	defer cancel()
	return FetchTop1000WithContext(ctx)
}

// FetchTop1000WithContext 从IYUU获取数据并返回（支持外部传入context）
func FetchTop1000WithContext(ctx context.Context) (*model.ProcessedData, error) {
	if !taskMutex.TryLock() {
		return nil, fmt.Errorf("任务正在执行中")
	}
	defer taskMutex.Unlock()

	var lastErr error
	for attempt := 0; attempt < maxRetries; attempt++ {
		if attempt > 0 {
			log.Printf("[%s] 第 %d 次重试...", logPrefix, attempt)
			time.Sleep(retryInterval)
		}

		data, err := doFetchWithContext(ctx)
		if err == nil {
			return data, nil
		}
		lastErr = err
		log.Printf("[%s] 第 %d 次尝试失败: %v", logPrefix, attempt+1, err)
	}

	return nil, lastErr
}

// doFetchWithContext 执行HTTP请求获取数据（支持外部传入context）
func doFetchWithContext(ctx context.Context) (*model.ProcessedData, error) {
	log.Printf("[%s] 开始爬取IYUU数据...", logPrefix)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, config.DefaultAPIURL, nil)
	if err != nil {
		return nil, fmt.Errorf("创建HTTP请求失败: %w", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("获取数据失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API返回错误状态码: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应体失败: %w", err)
	}

	log.Printf("[%s] 数据获取成功（%d 字节）", logPrefix, len(body))

	processed := parseResponse(string(body))
	if err := processed.Validate(); err != nil {
		log.Printf("[%s] 数据验证失败: %v", logPrefix, err)
		return nil, err
	}

	return &processed, nil
}

// parseResponse 解析原始文本为结构化数据
func parseResponse(rawData string) model.ProcessedData {
	lines := strings.Split(normalizeLineEndings(rawData), "\n")

	timeLine := ""
	dataLines := []string{}
	if len(lines) > 0 {
		timeLine = lines[timeLineIndex]
	}
	if len(lines) > dataStartLine {
		dataLines = lines[dataStartLine:]
	}

	items, skippedCount := parseDataLines(dataLines)

	logParsingWarnings(dataLines, skippedCount)
	log.Printf("[%s] 数据解析完成（%d 条）", logPrefix, len(items))

	return model.ProcessedData{
		Time:  extractTime(timeLine),
		Items: items,
	}
}

// normalizeLineEndings 统一换行符为\n
func normalizeLineEndings(s string) string {
	return strings.ReplaceAll(s, "\r\n", "\n")
}

// parseDataLines 解析数据行
func parseDataLines(dataLines []string) ([]model.SiteItem, int) {
	var items []model.SiteItem
	skippedCount := 0

	for i := 0; i <= len(dataLines)-linesPerItem; i += linesPerItem {
		group := dataLines[i : i+linesPerItem]

		item, ok := parseItemGroup(group)
		if !ok {
			skippedCount++
			continue
		}

		item.ID = len(items) + 1
		items = append(items, item)
	}

	return items, skippedCount
}

// parseItemGroup 解析单组数据（3行）
func parseItemGroup(group []string) (model.SiteItem, bool) {
	match := siteRegex.FindStringSubmatch(group[0])
	if len(match) < 3 {
		return model.SiteItem{}, false
	}

	return model.SiteItem{
		SiteName:    match[1],
		SiteID:      match[2],
		Duplication: extractFieldValue(group[1]),
		Size:        extractFieldValue(group[2]),
	}, true
}

// extractFieldValue 从"字段名：值"格式中提取值
func extractFieldValue(line string) string {
	parts := strings.Split(line, fieldSeparator)
	if len(parts) > 1 {
		return strings.TrimSpace(parts[1])
	}
	return ""
}

// logParsingWarnings 记录解析警告
func logParsingWarnings(dataLines []string, skippedCount int) {
	remainingLines := len(dataLines) % linesPerItem
	if remainingLines != 0 {
		log.Printf("[%s] 警告：剩余 %d 行未处理", logPrefix, remainingLines)
	}
	if skippedCount > 0 {
		log.Printf("[%s] 警告：跳过 %d 条格式错误的数据", logPrefix, skippedCount)
	}
}

// extractTime 提取时间字符串，去除前缀和后缀
func extractTime(rawTime string) string {
	rawTime = strings.TrimPrefix(rawTime, timePrefix)
	if idx := strings.Index(rawTime, timeSuffix); idx != -1 {
		rawTime = rawTime[:idx]
	}
	return rawTime
}

// PreloadData 启动时预加载数据（如果Redis中没有数据或数据过期）
func PreloadData() {
	log.Println("[🔍 爬虫] 检查是否需要预加载数据...")

	// 创建带超时的context（启动时预加载不希望等待太久）
	ctx, cancel := context.WithTimeout(context.Background(), httpTimeout)
	defer cancel()

	// 检查数据状态（存在性+过期检查）
	needsLoad := checkDataLoadRequired(ctx)
	if !needsLoad {
		log.Println("[🔍 爬虫] ✅ Redis中已有新鲜数据，无需预加载")
		return
	}

	// 没有数据或数据过期，尝试获取新数据
	log.Println("[🔍 爬虫] 🚀 Redis中无数据或数据过期，开始预加载...")
	data, err := FetchTop1000WithContext(ctx)
	if err != nil {
		log.Printf("[🔍 爬虫] ❌ 预加载失败: %v", err)
		log.Printf("[🔍 爬虫] 💡 提示：首次访问时会自动重试获取数据")
		return
	}

	// 存入Redis（使用同一个context）
	if err := storage.SaveDataWithContext(ctx, *data); err != nil {
		log.Printf("[🔍 爬虫] ❌ 保存预加载数据失败: %v", err)
		return
	}

	log.Printf("[🔍 爬虫] ✅ 预加载成功，已存入Redis（共 %d 条记录）", len(data.Items))
}

// checkDataLoadRequired 检查是否需要加载数据（支持外部传入context）
func checkDataLoadRequired(ctx context.Context) bool {
	// 检查数据是否存在
	exists, err := storage.DataExistsWithContext(ctx)
	if err != nil {
		log.Printf("[🔍 爬虫] ⚠️ 检查数据存在性失败: %v", err)
		// 出错时保守处理,视为需要加载
		return true
	}
	if !exists {
		return true
	}

	// 检查数据是否过期
	isExpired, err := storage.IsDataExpiredWithContext(ctx)
	if err != nil {
		log.Printf("[🔍 爬虫] ⚠️ 检查数据过期失败: %v", err)
		// 出错时保守处理,视为需要加载
		return true
	}

	return isExpired
}
