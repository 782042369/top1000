package main

import (
	"log"
	"os"

	"github.com/joho/godotenv"
	"top1000/internal/server"
)

func main() {
	// 加载 .env 文件（如果有的话）
	if err := godotenv.Load(); err != nil {
		log.Printf("⚠️ 警告: 无法加载 .env 文件: %v", err)
		log.Println("🔧 将使用系统环境变量")
	}

	// 检查必需的环境变量（Redis配置必须要有）
	requiredEnvs := []string{"REDIS_ADDR", "REDIS_PASSWORD"}
	missingEnvs := []string{}
	for _, env := range requiredEnvs {
		if os.Getenv(env) == "" {
			missingEnvs = append(missingEnvs, env)
		}
	}

	// 缺少必需的环境变量则直接退出
	if len(missingEnvs) > 0 {
		log.Fatalf("❌ 缺少必需的环境变量: %v\n请检查 .env 文件或系统环境变量配置", missingEnvs)
	}

	// 启动服务器
	server.StartWatcher()
}
