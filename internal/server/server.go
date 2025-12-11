package server

import (
	"fmt"
	"log"
	"os"
	"os/user"
	"strings"
	"time"

	"top1000/internal/api"
	"top1000/internal/crawler"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/compress"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/robfig/cron/v3"
)

// StartWatcher 启动服务监控
func StartWatcher() {
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
	app.Get("/top1000.json", api.GetTop1000Data)

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
	if err := crawler.InitializeData(); err != nil {
		log.Printf("初始化数据失败: %v", err)
		log.Printf("当前工作目录: %s", os.Getenv("PWD"))
		log.Printf("当前用户ID: %d", os.Getuid())
		log.Printf("public目录权限详情: %v", getFilePermissionInfo("./public"))
	}

	// 安排定时任务定期更新数据
	c := cron.New()
	c.AddFunc("@daily", func() {
		if err := crawler.ScheduleJob(); err != nil {
			log.Printf("定时任务执行失败: %v", err)
		}
	})
	c.Start()

	// 启动服务器
	log.Fatal(app.Listen(":7066"))
}

// getFilePermissionInfo 获取文件或目录的权限信息
func getFilePermissionInfo(path string) string {
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return "目录不存在"
		}
		return fmt.Sprintf("获取信息失败: %v", err)
	}

	// 获取当前用户
	currentUser, userErr := user.Current()
	userInfo := "未知"
	if userErr == nil {
		userInfo = fmt.Sprintf("用户名:%s, UID:%s", currentUser.Username, currentUser.Uid)
	}

	// 获取所有者信息（在Unix系统上）
	return fmt.Sprintf("权限: %s, 类型: %s, 当前用户: %s", info.Mode(), info.Mode().Type(), userInfo)
}
