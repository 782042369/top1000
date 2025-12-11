package api

import (
	"encoding/json"
	"log"
	"os"
	"top1000/internal/model"

	"github.com/gofiber/fiber/v2"
)

const (
	jsonFilePath = "./public/top1000.json"
)

// GetTop1000Data 提供top1000数据的API接口
func GetTop1000Data(c *fiber.Ctx) error {
	file, err := os.Open(jsonFilePath)
	if err != nil {
		log.Printf("打开数据文件失败: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "无法读取数据文件",
		})
	}
	defer file.Close()

	var data model.ProcessedData
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&data); err != nil {
		log.Printf("解析数据文件失败: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "数据解析失败",
		})
	}

	return c.JSON(data)
}
