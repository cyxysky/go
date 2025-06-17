package routes

import (
	"gin-web-api/config"
	"gin-web-api/handlers"
	"gin-web-api/middleware"

	"github.com/gin-gonic/gin"
)

func SetupRoutes(cfg *config.Config) *gin.Engine {
	// 设置Gin模式
	gin.SetMode(cfg.GinMode)

	r := gin.New()

	// 中间件
	r.Use(gin.Logger())
	r.Use(gin.Recovery())
	r.Use(middleware.CORSMiddleware())

	// 初始化处理器
	authHandler := handlers.NewAuthHandler(cfg)
	postHandler := handlers.NewPostHandler()

	// API版本分组
	v1 := r.Group("/api/v1")
	{
		// 健康检查
		v1.GET("/health", func(c *gin.Context) {
			c.JSON(200, gin.H{
				"status":  "ok",
				"message": "服务运行正常",
			})
		})

		// 认证相关路由
		auth := v1.Group("/auth")
		{
			auth.POST("/register", authHandler.Register)
			auth.POST("/login", authHandler.Login)
			auth.POST("/logout", middleware.JWTMiddleware(cfg), authHandler.Logout)
			auth.GET("/profile", middleware.JWTMiddleware(cfg), authHandler.GetProfile)
		}

		// 文章相关路由
		posts := v1.Group("/posts")
		{
			posts.GET("", postHandler.GetPosts)           // 获取文章列表
			posts.GET("/:id", postHandler.GetPost)        // 获取单个文章
			
			// 需要认证的路由
			posts.Use(middleware.JWTMiddleware(cfg))
			posts.POST("", postHandler.CreatePost)        // 创建文章
			posts.PUT("/:id", postHandler.UpdatePost)     // 更新文章
			posts.DELETE("/:id", postHandler.DeletePost)  // 删除文章
		}
	}

	return r
} 