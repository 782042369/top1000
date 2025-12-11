package crawler

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"
	"top1000/internal/model"
)

const (
	jsonFilePath = "./public/top1000.json"
)

var siteRegex = regexp.MustCompile(`站名：(.*?) 【ID：(\d+)】`)

// InitializeData 创建public目录和初始数据文件（如果不存在）
func InitializeData() error {
	// 如果public目录不存在则创建
	if err := os.MkdirAll("./public", 0755); err != nil {
		log.Printf("创建public目录失败详细信息: 当前工作目录=%s, 用户ID=%d, 错误=%v",
			os.Getenv("PWD"), os.Getuid(), err)
		log.Printf("创建public目录失败: %v", err)
		return err
	}

	// 检查数据文件是否存在
	if _, err := os.Stat(jsonFilePath); err != nil {
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
	resp, err := http.Get("https://api.iyuu.cn/top1000.php")
	if err != nil {
		log.Printf("获取数据失败: %v", err)
		return err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("读取响应体失败: %v", err)
		return err
	}

	processed := processData(string(body))

	file, err := os.Create(jsonFilePath)
	if err != nil {
		log.Printf("创建文件失败: %v", err)
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
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

// checkExpired 验证数据是否超过一天，如果需要则更新
func checkExpired() error {
	file, err := os.Open(jsonFilePath)
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

	// 正确比较时间差是否超过24小时
	if time.Since(dataTime).Hours() > 24 {
		return ScheduleJob()
	}

	return nil
}
