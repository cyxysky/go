package main

import (
	"log"

	"gin-web-api/config"
	"gin-web-api/database"
	"gin-web-api/models"
	redisClient "gin-web-api/redis"
	"gin-web-api/routes"
)

func main() {
	// 加载配置
	cfg := config.LoadConfig()

	// 初始化数据库
	database.InitDatabase(cfg)

	// 初始化Redis
	redisClient.InitRedis(cfg)

	// 自动迁移数据库表
	if err := database.AutoMigrate(&models.User{}, &models.Post{}); err != nil {
		log.Fatal("数据库迁移失败:", err)
	}

	// 设置路由
	r := routes.SetupRoutes(cfg)

	// 启动服务器
	log.Printf("服务器启动在端口 %s", cfg.Port)
	if err := r.Run(":" + cfg.Port); err != nil {
		log.Fatal("服务器启动失败:", err)
	}
}
