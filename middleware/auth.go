package middleware

import (
	"strings"
	"video_feed/utils"

	"github.com/gin-gonic/gin"
)

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		//如果token为空的情况
		if authHeader == "" {
			c.AbortWithStatusJSON(401, gin.H{"error": "未提供token"})
			return
		}
		//不为空则按空格分成两部分,检测格式是否正确
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.AbortWithStatusJSON(401, gin.H{"error": "token格式错误"})
			return
		}
		//运用utils里的PaserToken方法，检测是否过期
		userID, err := utils.ParseToken(parts[1])
		if err != nil {
			c.AbortWithStatusJSON(401, gin.H{"error": "token 无效"})
			return
		}
		//从token解析用户id并赋值
		c.Set("user_id", userID)
		c.Next()
	}
}
