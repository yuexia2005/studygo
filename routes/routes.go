package routes

import (
	"time"
	"video_feed/controllers"
	"video_feed/middleware"

	"github.com/gin-gonic/gin"
)

func SetupRouter() *gin.Engine {
	r := gin.Default()

	// 全局限流
	r.Use(middleware.RateLimit())

	// CORS中间件
	r.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "POST, GET, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Authorization, Content-Type")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	})
	//--------------

	//配置静态服务
	r.Static("/uploads", "./uploads")
	//公开路由
	r.GET("/", func(c *gin.Context) {
		c.File("./videofeed.html")
	})
	r.POST("/register", controllers.Register)
	r.POST("/login", controllers.Login)

	//健康检查路由(供 Docker 探测）
	r.GET("/health", func(c *gin.Context) {
		c.String(200, "ok")
	})

	r.GET("/slow", func(c *gin.Context) {
		time.Sleep(10 * time.Second)
		c.String(200, "slow response")
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
	return r
}
