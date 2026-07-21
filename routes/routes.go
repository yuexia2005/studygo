package routes

import (
	"context"
	"net/http"
	"os"
	"strings"
	"time"
	"video_feed/controllers"
	"video_feed/middleware"
	"video_feed/models"

	"github.com/gin-gonic/gin"
)

func SetupRouter() *gin.Engine {
	r := gin.Default()

	// 全局限流
	r.Use(middleware.RateLimit())

	// CORS中间件：通过 CORS_ORIGIN 环境变量控制允许的域名
	// 未设置时不添加 Allow-Origin，仅允许同源请求（生产环境推荐）
	// 本地开发可设置 CORS_ORIGIN=* 或 CORS_ORIGIN=http://localhost:3000
	r.Use(func(c *gin.Context) {
		if origin := os.Getenv("CORS_ORIGIN"); origin != "" {
			c.Header("Access-Control-Allow-Origin", origin)
			c.Header("Access-Control-Allow-Methods", "POST, GET, PUT, DELETE, OPTIONS")
			c.Header("Access-Control-Allow-Headers", "Authorization, Content-Type")
		}
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	})
	//--------------

	//配置静态服务
	r.Static("/uploads", "./uploads")
	// Vue3 前端编译产物的静态资源 (JS/CSS/图片等)
	r.Static("/assets", "./dist/assets")
	// 公开路由 - 首页用 Vue3 前端
	r.GET("/", func(c *gin.Context) {
		c.File("./dist/index.html")
	})
	r.POST("/register", controllers.Register)
	r.POST("/login", controllers.Login)

	//健康检查路由(供 Docker 探测，同时验证 DB 和 Redis 连接）
	r.GET("/health", func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()

		// 检查数据库
		sqlDB, err := models.DB.DB()
		if err != nil || sqlDB.PingContext(ctx) != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"status": "unhealthy",
				"db":     "disconnected",
			})
			return
		}

		// 检查 Redis
		if _, err := models.RDB.Ping(ctx).Result(); err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"status": "unhealthy",
				"db":     "connected",
				"redis":  "disconnected",
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"status": "healthy",
			"db":     "connected",
			"redis":  "connected",
		})
	})

	//需要认证的路由
	auth := r.Group("/api")
	//添加中间件
	auth.Use(middleware.AuthMiddleware())
	{
		auth.POST("/video/upload", controllers.UploadVideo)
		auth.DELETE("/video/:id", controllers.DeleteVideo)

		auth.GET("/feed", controllers.GetFeed)

		auth.POST("/video/:id/like", controllers.ToggleLike)
		auth.GET("/hot", controllers.GetHot)

		auth.POST("/video/:id/comment", controllers.CreateComment)
		auth.GET("/video/:id/comments", controllers.GetComments)
		auth.DELETE("/comment/:id", controllers.DeleteComment)
	}

	// SPA 前端路由兜底：未匹配的 GET 请求返回 index.html
	// 解决 Vue Router 直接访问 /hot、/feed 等页面刷新时 404 的问题
	r.NoRoute(func(c *gin.Context) {
		if strings.HasPrefix(c.Request.URL.Path, "/api/") {
			c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
			return
		}
		c.File("./dist/index.html")
	})

	return r
}
