package crawler

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"
	"top1000/internal/config"
	"top1000/internal/model"
)

var siteRegex = regexp.MustCompile(`站名：(.*?) 【ID：(\d+)】`)

const (
	// HTTP请求超时时间
	httpTimeout = 30 * time.Second
)

// InitializeData 创建public目录和初始数据文件（如果不存在）
func InitializeData() error {
	cfg := config.Get()

	// 如果public目录不存在则创建
	if err := os.MkdirAll(cfg.DataDir, 0755); err != nil {
		log.Printf("创建数据目录失败: %v", err)
		return err
	}

	// 检查数据文件是否存在
	if _, err := os.Stat(cfg.DataFilePath); err != nil {
		if os.IsNotExist(err) {
			return ScheduleJob()
		}
		// 如果是其他错误，记录但不中断程序
		log.Printf("检查数据文件时发生错误: %v", err)
		return ScheduleJob()
	}

	// 检查数据是否过期
	return checkExpired()
}

// ScheduleJob 从远程API获取并处理数据
func ScheduleJob() error {
	cfg := config.Get()

	// 创建带超时的context
	ctx, cancel := context.WithTimeout(context.Background(), httpTimeout)
	defer cancel()

	// 创建HTTP请求
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, cfg.Top1000APIURL, nil)
	if err != nil {
		log.Printf("创建HTTP请求失败: %v", err)
		return err
	}

	// 执行请求
	client := &http.Client{
		Timeout: httpTimeout,
	}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("获取数据失败: %v", err)
		return err
	}
	defer resp.Body.Close()

	// 检查响应状态码
	if resp.StatusCode != http.StatusOK {
		log.Printf("API返回错误状态码: %d", resp.StatusCode)
		return err
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("读取响应体失败: %v", err)
		return err
	}

	processed := processData(string(body))

	file, err := os.Create(cfg.DataFilePath)
	if err != nil {
		log.Printf("创建文件失败: %v", err)
		return err
	}
	defer file.Close()

	// 不使用缩进以减小文件大小
	encoder := json.NewEncoder(file)
	if err := encoder.Encode(processed); err != nil {
		log.Printf("写入JSON数据失败: %v", err)
		return err
	}

	log.Println("数据更新成功")
	return nil
}

// processData 将原始数据转换为结构化格式
func processData(rawData string) model.ProcessedData {
	lines := strings.Split(strings.ReplaceAll(rawData, "\r\n", "\n"), "\n")
	timeLine := ""
	dataLines := []string{}

	if len(lines) > 0 {
		timeLine = lines[0]
	}
	if len(lines) > 2 {
		dataLines = lines[2:]
	}

	var items []model.SiteItem

	// 以3行为一组处理数据
	for i := 0; i <= len(dataLines)-3; i += 3 {
		group := dataLines[i : i+3]
		siteLine := group[0]
		dupLine := group[1]
		sizeLine := group[2]

		match := siteRegex.FindStringSubmatch(siteLine)
		if len(match) < 3 {
			continue
		}

		siteName := match[1]
		siteID := match[2]

		duplication := ""
		size := ""

		dupParts := strings.Split(dupLine, "：")
		if len(dupParts) > 1 {
			duplication = strings.TrimSpace(dupParts[1])
		}

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

	return model.ProcessedData{
		Time:  parseTime(timeLine),
		Items: items,
	}
}

// parseTime 提取并格式化时间字符串
func parseTime(rawTime string) string {
	rawTime = strings.Replace(rawTime, "create time ", "", 1)
	rawTime = strings.Replace(rawTime, " by http://api.iyuu.cn/ptgen/", "", 1)
	// 修改: 支持新的时间格式 '2025-12-11 07:52:33 by https://api.iyuu.cn/'
	rawTime = strings.Split(rawTime, " by ")[0]
	return rawTime
}

// checkExpired 验证数据是否超过配置的过期时间，如果需要则更新
func checkExpired() error {
	cfg := config.Get()

	file, err := os.Open(cfg.DataFilePath)
	if err != nil {
		return ScheduleJob()
	}
	defer file.Close()

	var data model.ProcessedData
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&data); err != nil {
		return ScheduleJob()
	}

	// 使用正确的布局格式解析时间 "2025-12-11 07:52:33"
	dataTime, err := time.Parse("2006-01-02 15:04:05", data.Time)
	if err != nil {
		return ScheduleJob()
	}

	// 比较时间差是否超过配置的过期时间
	if time.Since(dataTime) > cfg.DataExpireDuration {
		return ScheduleJob()
	}

	return nil
}
