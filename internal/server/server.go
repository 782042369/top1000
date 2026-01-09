package server

import (
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"top1000/internal/api"
	"top1000/internal/config"
	"top1000/internal/storage"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/compress"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/limiter"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
)

// StartWatcher 启动Web服务器（配置中间件、路由、Redis）
func StartWatcher() {
	cfg := config.Get()

	// 验证配置
	if err := config.Validate(); err != nil {
		log.Fatalf("❌ 配置验证失败: %v", err)
	}

	log.Println("========================================")
	log.Println("   Top1000 服务正在启动...")
	log.Println("========================================")

	app := fiber.New(fiber.Config{
		AppName:      "Top1000 Service",
		StrictRouting: true, // 启用严格路由
		BodyLimit:    4 * 1024 * 1024, // 限制请求体大小为 4MB
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	})

	// 中间件配置
	app.Use(recover.New()) // 错误恢复

	// 日志中间件（配置详细日志）
	app.Use(logger.New(logger.Config{
		Format:     "[${time}] ${status} - ${method} ${path} - ${latency}\n",
		TimeFormat: "2006-01-02 15:04:05",
		TimeZone:   "Asia/Shanghai",
	}))

	// CORS 中间件（生产环境限制来源）
	corsOrigins := os.Getenv("CORS_ORIGINS")
	if corsOrigins == "" {
		corsOrigins = "*"
	}
	// 当使用通配符时，不能启用 AllowCredentials（安全限制）
	// 只有在指定具体域名时才允许携带凭证
	allowCredentials := corsOrigins != "*"

	app.Use(cors.New(cors.Config{
		AllowOrigins:     corsOrigins,
		AllowMethods:     "GET,OPTIONS",
		AllowHeaders:     "Origin,Content-Type,Accept,Authorization",
		ExposeHeaders:    "Content-Length,ETag,Cache-Control",
		MaxAge:           86400, // 预检请求缓存 24 小时
		AllowCredentials: allowCredentials,
	}))

	// 安全响应头（手动配置，不使用Helmet）
	// Helmet的COEP配置无法禁用，因此手动配置
	app.Use(func(c *fiber.Ctx) error {
		// XSS保护
		c.Set("X-XSS-Protection", "1; mode=block")
		// 禁止MIME类型嗅探
		c.Set("X-Content-Type-Options", "nosniff")
		// 防止点击劫持
		c.Set("X-Frame-Options", "DENY")
		// CSP：允许外部监控脚本、图片、数据上报
		c.Set("Content-Security-Policy", "default-src 'self'; script-src 'self' 'unsafe-inline' 'unsafe-eval' https://log.939593.xyz; img-src 'self' data: https: https://lsky.939593.xyz:11111; style-src 'self' 'unsafe-inline'; connect-src 'self' https://log.939593.xyz;")
		// 不设置COEP和COOP，允许跨域资源加载
		return c.Next()
	})

	// 速率限制（防止 DDoS）
	app.Use(limiter.New(limiter.Config{
		Max:        100, // 每分钟最多 100 次请求
		Expiration: 1 * time.Minute,
		KeyGenerator: func(c *fiber.Ctx) string {
			return c.IP() // 基于 IP 限流
		},
		LimitReached: func(c *fiber.Ctx) error {
			return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
				"error": "请求过于频繁，请稍后再试",
			})
		},
		SkipFailedRequests:     false,
		SkipSuccessfulRequests: false,
	}))

	// 响应压缩
	app.Use(compress.New(compress.Config{
		Level: compress.LevelBestSpeed,
	}))

	// API路由应在静态文件之前定义，避免被静态文件中间件拦截
	app.Get("/top1000.json", api.GetTop1000Data)

	// 健康检查端点
	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status":    "ok",
			"timestamp": time.Now().Unix(),
		})
	})

	// 静态文件服务（限制只服务于/web-dist路径下的文件）
	// 为非HTML文件添加一年缓存，HTML文件不缓存
	app.Static("/", cfg.WebDistDir, fiber.Static{
		CacheDuration: cfg.CacheDuration,
		Browse:        true,
		MaxAge:        0, // 默认不缓存
		ModifyResponse: func(c *fiber.Ctx) error {
			// 检查文件扩展名，只为非HTML文件设置缓存
			path := c.Path()
			if !strings.HasSuffix(path, ".html") && !strings.HasSuffix(path, "/") && c.Response().StatusCode() == fiber.StatusOK {
				// 非HTML文件且不是目录，设置长期缓存
				c.Response().Header.Set("Cache-Control", "public, max-age=31536000") // 一年缓存
			} else {
				// HTML文件或目录索引不缓存
				c.Response().Header.Set("Cache-Control", "no-cache, no-store, must-revalidate")
				c.Response().Header.Set("Pragma", "no-cache")
				c.Response().Header.Set("Expires", "0")
			}
			return nil
		},
	})

	// 初始化 Redis
	log.Println("正在初始化 Redis 连接...")
	if err := storage.InitRedis(); err != nil {
		log.Fatalf("❌ Redis 初始化失败: %v", err)
	}
	log.Println("✅ Redis 初始化成功")

	// 在后台监听系统信号，实现优雅关闭
	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
		<-sigChan

		log.Println("接收到关闭信号，正在优雅关闭服务器...")

		// 关闭 Redis 连接
		if err := storage.CloseRedis(); err != nil {
			log.Printf("关闭 Redis 连接失败: %v", err)
		} else {
			log.Println("Redis 连接已关闭")
		}

		// 关闭 HTTP 服务器
		if err := app.ShutdownWithTimeout(10 * time.Second); err != nil {
			log.Printf("服务器关闭失败: %v", err)
		} else {
			log.Println("服务器已优雅关闭")
		}
	}()

	// 启动服务器
	log.Println("========================================")
	log.Printf("✅ 服务已启动，监听端口: %s", cfg.Port)
	log.Printf("📦 存储方式: Redis (%s)", cfg.RedisAddr)
	log.Println("🔄 数据更新策略: 实时更新（过期时自动获取）")
	log.Println("🔒 安全措施: 速率限制、安全响应头、CORS 保护")
	log.Println("========================================")
	log.Fatal(app.Listen(":" + cfg.Port))
}
