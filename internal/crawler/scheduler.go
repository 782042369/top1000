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
	httpTimeout          = 30 * time.Second
	maxRetries           = 1 // 小项目不需要太多重试
	retryInterval        = 2 * time.Second // 缩短重试间隔
	linesPerItem         = 3
	timeLineIndex        = 0
	dataStartLineIndex   = 2
	timePrefix           = "create time "
	timeSuffix           = " by "
	fieldSeparator       = "："
	sitePattern          = `站名：(.*?) 【ID：(\d+)】`
)

var (
	siteRegex = regexp.MustCompile(sitePattern)
	taskMutex sync.Mutex
)

// FetchData 从IYUU获取数据，带重试机制
func FetchData() error {
	if !taskMutex.TryLock() {
		log.Println("任务正在执行中，跳过本次调度")
		return fmt.Errorf("任务正在执行中")
	}
	defer taskMutex.Unlock()

	storage.SetUpdating(true)
	defer storage.SetUpdating(false)

	var lastErr error
	for attempt := 0; attempt < maxRetries; attempt++ {
		if attempt > 0 {
			log.Printf("第 %d 次重试中...", attempt)
			time.Sleep(retryInterval)
		}

		if err := doFetch(); err == nil {
			return nil
		} else {
			lastErr = err
			log.Printf("第 %d 次尝试失败: %v", attempt+1, err)
		}
	}

	log.Printf("重试 %d 次后仍失败，终止", maxRetries)
	return lastErr
}

// doFetch 执行HTTP请求获取数据
func doFetch() error {
	cfg := config.Get()
	log.Println("开始获取数据...")

	ctx, cancel := context.WithTimeout(context.Background(), httpTimeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, cfg.Top1000APIURL, nil)
	if err != nil {
		return fmt.Errorf("创建HTTP请求失败: %w", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("获取数据失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("API返回错误状态码: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("读取响应体失败: %w", err)
	}

	log.Printf("✅ 数据获取成功（大小: %d 字节）", len(body))

	processed := processData(string(body))
	if err := storage.SaveData(processed); err != nil {
		log.Printf("❌ 保存数据到Redis失败: %v", err)
		return err
	}

	log.Println("✅ 数据更新完成")
	return nil
}

// processData 解析原始文本为结构化数据
func processData(rawData string) model.ProcessedData {
	lines := strings.Split(normalizeLineEndings(rawData), "\n")

	timeLine := ""
	dataLines := []string{}
	if len(lines) > 0 {
		timeLine = lines[timeLineIndex]
	}
	if len(lines) > dataStartLineIndex {
		dataLines = lines[dataStartLineIndex:]
	}

	items, skippedCount := parseDataLines(dataLines)

	logWarnings(dataLines, skippedCount)
	log.Printf("数据处理完成：共 %d 条记录", len(items))

	result := model.ProcessedData{
		Time:  parseTime(timeLine),
		Items: items,
	}

	if err := result.Validate(); err != nil {
		log.Printf("⚠️ 数据验证失败: %v", err)
	}

	return result
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

// logWarnings 记录解析警告
func logWarnings(dataLines []string, skippedCount int) {
	remainingLines := len(dataLines) % linesPerItem
	if remainingLines != 0 {
		log.Printf("警告：数据行数不是%d的倍数，剩余 %d 行未处理", linesPerItem, remainingLines)
	}
	if skippedCount > 0 {
		log.Printf("警告：跳过了 %d 条格式不正确的数据", skippedCount)
	}
}

// parseTime 提取时间字符串，去除前缀和后缀
func parseTime(rawTime string) string {
	rawTime = strings.TrimPrefix(rawTime, timePrefix)
	if idx := strings.Index(rawTime, timeSuffix); idx != -1 {
		rawTime = rawTime[:idx]
	}
	return rawTime
}
