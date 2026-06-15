package middleware

import (
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

// 基于用户ID的限流器映射
var userLimiters sync.Map

// 基于IP的限流器映射（用于未登录用户）
var ipLimiters sync.Map

// getUserLimiter 获取或创建用户级别的限流器
func getUserLimiter(userID uint) *rate.Limiter {
	limiter, _ := userLimiters.LoadOrStore(userID, rate.NewLimiter(5, 20))
	return limiter.(*rate.Limiter)
}

// getIPLimiter 获取或创建IP级别的限流器
func getIPLimiter(ip string) *rate.Limiter {
	limiter, _ := ipLimiters.LoadOrStore(ip, rate.NewLimiter(10, 50))
	return limiter.(*rate.Limiter)
}

// RateLimit 限流中间件
// - 已登录用户：基于 userID 限流，每用户 5 QPS，突发 20
// - 未登录用户：基于 IP 限流，每 IP 10 QPS，突发 50

func RateLimit() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 对 OPTIONS 请求直接放行，不进行限流
		if c.Request.Method == "OPTIONS" {
			c.Next()
			return
		}

		// 判断用户是否已登录，使用不同的限流策略
		userIDRaw, exists := c.Get("user_id")
		if exists {
			// 已登录用户：基于用户ID限流
			userID, ok := userIDRaw.(uint)
			if !ok {
				c.Next()
				return
			}
			limiter := getUserLimiter(userID)
			if !limiter.Allow() {
				c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
					"error": "请求过于频繁,请稍后再试",
				})
				return
			}
		} else {
			// 未登录用户：基于 IP 限流
			clientIP := c.ClientIP()
			limiter := getIPLimiter(clientIP)
			if !limiter.Allow() {
				c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
					"error": "请求过于频繁,请稍后再试",
				})
				return
			}
		}

		c.Next()
	}
}
