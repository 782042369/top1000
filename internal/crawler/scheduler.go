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

var (
	siteRegex = regexp.MustCompile(`站名：(.*?) 【ID：(\d+)】`)
	// 任务互斥锁，防止并发更新
	taskMutex sync.Mutex
)

const (
	// HTTP请求超时，30秒足够
	httpTimeout = 30 * time.Second
	// 最多重试3次，避免无限重试
	maxRetries = 3
	// 重试间隔5秒，给API恢复时间
	retryInterval = 5 * time.Second
	// 每3行一条数据（站点名、重复度、大小）
	linesPerItem = 3
	// 时间行在第0行
	timeLineIndex = 0
	// 数据从第2行开始（跳过时间行和空行）
	dataStartLineIndex = 2
)

// FetchData 从IYUU获取数据，带重试机制
func FetchData() error {
	// 加锁，防止并发更新
	if !taskMutex.TryLock() {
		log.Println("任务正在执行中，跳过本次调度")
		return fmt.Errorf("任务正在执行中")
	}
	defer taskMutex.Unlock()

	// 标记正在更新
	storage.SetUpdating(true)
	defer storage.SetUpdating(false)

	var lastErr error

	// 重试循环，最多3次
	for attempt := 0; attempt < maxRetries; attempt++ {
		if attempt > 0 {
			log.Printf("第 %d 次重试中...", attempt)
			time.Sleep(retryInterval)
		}

		err := doFetch()
		if err == nil {
			// 成功，完成
			return nil
		}

		// 记录错误，继续重试
		lastErr = err
		log.Printf("第 %d 次尝试失败: %v", attempt+1, err)
	}

	// 3次都失败，终止重试
	log.Printf("重试 %d 次后仍失败，终止", maxRetries)
	return lastErr
}

// doFetch 执行HTTP请求获取数据
func doFetch() error {
	cfg := config.Get()

	log.Println("开始获取数据...")

	// 创建带超时的context，30秒超时
	ctx, cancel := context.WithTimeout(context.Background(), httpTimeout)
	defer cancel()

	// 创建HTTP请求
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, cfg.Top1000APIURL, nil)
	if err != nil {
		log.Printf("创建HTTP请求失败: %v", err)
		return err
	}

	// 发送请求
	client := &http.Client{
		Timeout: httpTimeout,
	}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("获取数据失败: %v", err)
		return err
	}
	defer resp.Body.Close()

	// 检查状态码
	if resp.StatusCode != http.StatusOK {
		log.Printf("API返回错误状态码: %d", resp.StatusCode)
		return fmt.Errorf("API返回错误状态码: %d", resp.StatusCode)
	}

	// 读取响应体
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("读取响应体失败: %v", err)
		return err
	}

	log.Printf("✅ 数据获取成功（大小: %d 字节）", len(body))

	processed := processData(string(body))

	// 存储到Redis
	if err := storage.SaveData(processed); err != nil {
		log.Printf("❌ 保存数据到Redis失败: %v", err)
		return err
	}

	log.Println("✅ 数据更新完成")
	return nil
}

// processData 解析原始文本为结构化数据
func processData(rawData string) model.ProcessedData {
	lines := strings.Split(strings.ReplaceAll(rawData, "\r\n", "\n"), "\n")
	timeLine := ""
	dataLines := []string{}

	if len(lines) > 0 {
		timeLine = lines[timeLineIndex]
	}
	if len(lines) > dataStartLineIndex {
		dataLines = lines[dataStartLineIndex:]
	}

	var items []model.SiteItem
	skippedCount := 0

	// 每3行解析一条数据（站点名、重复度、大小）
	for i := 0; i <= len(dataLines)-linesPerItem; i += linesPerItem {
		group := dataLines[i : i+linesPerItem]
		siteLine := group[0]
		dupLine := group[1]
		sizeLine := group[2]

		// 使用正则提取站点名和ID
		match := siteRegex.FindStringSubmatch(siteLine)
		if len(match) < 3 {
			skippedCount++
			continue
		}

		siteName := match[1]
		siteID := match[2]

		// 提取重复度
		duplication := ""
		dupParts := strings.Split(dupLine, "：")
		if len(dupParts) > 1 {
			duplication = strings.TrimSpace(dupParts[1])
		}

		// 提取文件大小
		size := ""
		sizeParts := strings.Split(sizeLine, "：")
		if len(sizeParts) > 1 {
			size = strings.TrimSpace(sizeParts[1])
		}

		items = append(items, model.SiteItem{
			SiteName:    siteName,
			SiteID:      siteID,
			Duplication: duplication,
			Size:        size,
			ID:          len(items) + 1,
		})
	}

	// 检查是否有不完整的数据
	remainingLines := len(dataLines) % linesPerItem
	if remainingLines != 0 {
		log.Printf("警告：数据行数不是%d的倍数，剩余 %d 行未处理", linesPerItem, remainingLines)
	}

	if skippedCount > 0 {
		log.Printf("警告：跳过了 %d 条格式不正确的数据", skippedCount)
	}

	log.Printf("数据处理完成：共 %d 条记录", len(items))

	result := model.ProcessedData{
		Time:  parseTime(timeLine),
		Items: items,
	}

	// 验证数据，失败时仍返回（容错机制）
	if err := result.Validate(); err != nil {
		log.Printf("⚠️ 数据验证失败: %v", err)
		// 即使验证失败也返回数据，但记录警告
	}

	return result
}

// parseTime 提取时间字符串，去除前缀和后缀
func parseTime(rawTime string) string {
	// 去掉"create time "前缀
	rawTime = strings.TrimPrefix(rawTime, "create time ")

	// 去掉" by https://api.iyuu.cn/"后缀（如果有）
	if idx := strings.Index(rawTime, " by "); idx != -1 {
		rawTime = rawTime[:idx]
	}

	return rawTime
}
