package main

import (
	"log"

	"github.com/joho/godotenv"
	"top1000/internal/server"
)

func main() {
	// 加载 .env 文件（非必需，失败时使用系统环境变量）
	_ = godotenv.Load()

	log.Println("✅ 环境变量已加载")

	server.Start()
}
