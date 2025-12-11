package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/compress"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/robfig/cron/v3"
)

// SiteItem 表示top1000列表中的站点条目
type SiteItem struct {
	SiteName     string `json:"siteName"`
	SiteID       string `json:"siteid"`
	Duplication  string `json:"duplication"`
	Size         string `json:"size"`
	ID           int    `json:"id"`
}

// ProcessedData 表示结构化数据
type ProcessedData struct {
	Time  string     `json:"time"`
	Items []SiteItem `json:"items"`
}

const (
	jsonFilePath = "./public/top1000.json"
)

var siteRegex = regexp.MustCompile(`站名：(.*?) 【ID：(\d+)】`)

func main() {
	startWatcher()
}

// getTop1000Data 提供top1000数据的API接口
func getTop1000Data(c *fiber.Ctx) error {
	file, err := os.Open(jsonFilePath)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "无法读取数据文件",
		})
	}
	defer file.Close()
	log.Println("正在读取数据文件...", file)
	var data ProcessedData
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&data); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "数据解析失败",
		})
	}

	return c.JSON(data)
}

// initializeData 创建public目录和初始数据文件（如果不存在）
func initializeData() error {
	// 如果public目录不存在则创建
	if err := os.MkdirAll("./public", 0755); err != nil {
		return fmt.Errorf("创建public目录失败: %w", err)
	}

	// 检查数据文件是否存在
	if _, err := os.Stat(jsonFilePath); os.IsNotExist(err) {
		return scheduleJob()
	}

	// 检查数据是否过期
	return checkExpired()
}

// scheduleJob 从远程API获取并处理数据
func scheduleJob() error {
	log.Println("正在从远程API获取数据...")
	resp, err := http.Get("https://api.iyuu.cn/top1000.php")
	if err != nil {
		return fmt.Errorf("获取数据失败: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("读取响应体失败: %w", err)
	}

	processed := processData(string(body))

	file, err := os.Create(jsonFilePath)
	if err != nil {
		return fmt.Errorf("创建文件失败: %w", err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(processed); err != nil {
		return fmt.Errorf("写入JSON数据失败: %w", err)
	}

	log.Println("数据更新成功")
	return nil
}

// processData 将原始数据转换为结构化格式
func processData(rawData string) ProcessedData {
	lines := strings.Split(strings.ReplaceAll(rawData, "\r\n", "\n"), "\n")
	timeLine := ""
	dataLines := []string{}

	if len(lines) > 0 {
		timeLine = lines[0]
	}
	if len(lines) > 2 {
		dataLines = lines[2:]
	}

	var items []SiteItem

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

		items = append(items, SiteItem{
			SiteName:     siteName,
			SiteID:       siteID,
			Duplication:  duplication,
			Size:         size,
			ID:           len(items) + 1,
		})
	}

	return ProcessedData{
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
		log.Printf("打开数据文件失败，触发更新: %v", err)
		return scheduleJob()
	}
	defer file.Close()

	var data ProcessedData
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&data); err != nil {
		log.Printf("解码数据失败，触发更新: %v", err)
		return scheduleJob()
	}

	// 修改: 使用正确的布局格式解析时间 "2025-12-11 07:52:33"
	dataTime, err := time.Parse("2006-01-02 15:04:05", data.Time)
	if err != nil {
		log.Printf("时间解析失败，触发更新: %v", err)
		return scheduleJob()
	}

	// 修改: 正确比较时间差是否超过24小时
	if time.Since(dataTime).Hours() > 24 {
		log.Println("数据已过期，正在更新...")
		return scheduleJob()
	}

	return nil
}

func startWatcher() {
	app := fiber.New(fiber.Config{
		AppName: "Top1000 Service",
	})

	// 中间件
	app.Use(recover.New())
	app.Use(logger.New())
	app.Use(cors.New())
	app.Use(compress.New(compress.Config{
		Level: compress.LevelBestSpeed,
	}))

	// API路由应在静态文件之前定义，避免被静态文件中间件拦截
	app.Get("/top1000.json", getTop1000Data)

	// 静态文件服务 (限制只服务于/web-dist路径下的文件)
	// 为非HTML文件添加一年缓存，HTML文件不缓存
	app.Static("/", "./web-dist", fiber.Static{
		CacheDuration: 24 * time.Hour,
		Browse:        true,
		MaxAge:        0, // 默认不缓存
		ModifyResponse: func(c *fiber.Ctx) error {
			// 检查文件扩展名，只为非HTML文件设置缓存
			path := c.Path()
			if !strings.HasSuffix(path, ".html") && !strings.HasSuffix(path, "/") && c.Response().StatusCode() == fiber.StatusOK {
				// 非HTML文件且不是目录，设置长期缓存
				c.Response().Header.Set("Cache-Control", "public, max-age=31536000") // 一年缓存
			} else {
				// HTML文件或者目录索引不缓存
				c.Response().Header.Set("Cache-Control", "no-cache, no-store, must-revalidate")
				c.Response().Header.Set("Pragma", "no-cache")
				c.Response().Header.Set("Expires", "0")
			}
			return nil
		},
	})

	// 初始化数据
	if err := initializeData(); err != nil {
		log.Printf("初始化数据失败: %v", err)
	}

	// 安排定时任务定期更新数据
	c := cron.New()
	c.AddFunc("@daily", func() {
		if err := scheduleJob(); err != nil {
			log.Printf("定时任务执行失败: %v", err)
		}
	})
	c.Start()

	// 启动服务器
	log.Fatal(app.Listen(":7066"))
}
