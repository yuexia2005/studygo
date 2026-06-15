package controllers

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"time"
	"video_feed/models"

	"github.com/gin-gonic/gin"
	"github.com/rabbitmq/amqp091-go"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

// 辅助函数：当 Redis 缓存失效时，从数据库重建缓存
func rebuildLikeCache(ctx context.Context, vid uint) error {
	//每个视频有唯一的key
	likeKey := fmt.Sprintf("video:%d:likes", vid)

	// 从数据库查询所有点赞过该视频的用户 ID
	var userIDs []uint
	if err := models.DB.WithContext(ctx).
		Model(&models.Like{}).
		Where("video_id = ?", vid).
		Pluck("user_id", &userIDs).Error; err != nil {
		return fmt.Errorf("查询点赞用户失败: %w", err)
	}
	// 写入 Redis Set（所有操作都使用传入的 ctx）
	if len(userIDs) > 0 {
		// 将 uid 列表同步到 Redis Set
		members := make([]interface{}, len(userIDs))
		for i, id := range userIDs {
			members[i] = id
		}
		if err := models.RDB.SAdd(ctx, likeKey, members...).Err(); err != nil {
			return fmt.Errorf("Redis SAdd 失败: %w", err)
		}
	} else {
		// 如果没人点赞，也给 Redis 一个空集合标识，防止频繁回源数据库（缓存穿透）
		if err := models.RDB.SAdd(ctx, likeKey, -1).Err(); err != nil { // 放入一个不存在的占位符
			return fmt.Errorf("Redis SAdd 失败: %w", err)
		}
	}

	// 设置 24 小时过期，防止冷数据常驻内存
	if err := models.RDB.Expire(ctx, likeKey, 24*time.Hour).Err(); err != nil {
		return fmt.Errorf("Redis Expire 失败: %w", err)
	}
	return nil
}

// 降级逻辑：当 Redis 卡死或报错时，直接同步操作数据库
func handleLikeFallback(c *gin.Context, userID uint, vid uint) {
	log.Printf("[Fallback] Redis异常,触发数据库降级方案。用户: %d, 视频: %d", userID, vid)

	var like models.Like
	//直接查找MySQL
	result := models.DB.Where("user_id = ? AND video_id = ? ", userID, vid).First(&like)

	var finalLiked bool

	// 开启事务保证 Likes 表和 Videos 表同步
	err := models.DB.Transaction(func(tx *gorm.DB) error {
		if result.Error == nil {
			// 已点赞 -> 取消点赞
			if err := tx.Delete(&like).Error; err != nil {
				return err
			}
			finalLiked = false
			return tx.Model(&models.Video{}).Where("id = ? AND like_count > 0", vid).
				Update("like_count", gorm.Expr("like_count - 1")).Error
		} else {
			// 未点赞 -> 增加点赞
			if err := tx.Create(&models.Like{UserID: userID, VideoID: vid}).Error; err != nil {
				return err
			}
			finalLiked = true
			return tx.Model(&models.Video{}).Where("id = ?", vid).
				Update("like_count", gorm.Expr("like_count + 1")).Error
		}
	})
	if err != nil {
		c.JSON(500, gin.H{"error": "系统繁忙，请稍后再试"})
		return
	}

	//查最新的点赞总数返回前端
	var latestCount int64
	models.DB.Model(&models.Video{}).Where("id = ? ", vid).Pluck("like_count", &latestCount)

	c.JSON(200, gin.H{
		"liked":      finalLiked,
		"like_count": latestCount,
		"mode":       "fallback", // 告知前端或调试，当前处于降级模式
	})
}

// 定义一个控制器函数 ToggleLike,用于切换点赞状态,接收 Gin 上下文
func ToggleLike(c *gin.Context) {

	//从上下文获取userid
	userIDVal, exists := c.Get("user_id")
	if !exists {
		c.JSON(401, gin.H{"error": "未登录"})
		return
	}
	userID := userIDVal.(uint)
	//获取 URL 路径参数 id
	videoID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(400, gin.H{"error": "无效的视频ID"})
		return
	}
	vid := uint(videoID)
	ctx := context.Background()
	likeKey := fmt.Sprintf("video:%d:likes", vid)

	res, err := models.RedisCB.Execute(func() (interface{}, error) {
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		defer cancel()
		//检查缓存是否存在
		exists, err := models.RDB.Exists(ctx, likeKey).Result()
		if err != nil {
			return nil, err // Redis 直接报错，触发熔断降级
		}
		//如果不存在重建缓存
		if exists == 0 {
			if err := rebuildLikeCache(ctx, vid); err != nil {
				return nil, err
			}
		}

		//判断用户是否点赞
		isLiked, _ := models.RDB.SIsMember(ctx, likeKey, userID).Result()
		var action string
		var finalLiked bool

		if isLiked {
			//已点赞 需要取消点赞
			models.RDB.SRem(ctx, likeKey, userID)
			action = "unlike"
			finalLiked = false
		} else {
			//未点赞
			models.RDB.SAdd(ctx, likeKey, userID)
			action = "like"
			finalLiked = true
		}
		//获取点赞数（从redis获取）返回前端
		likeCount, _ := models.RDB.SCard(ctx, likeKey).Result()

		// 返回给熔断器执行器的数据包
		return map[string]interface{}{
			"action":     action,
			"finalLiked": finalLiked,
			"likeCount":  likeCount,
		}, nil
	})

	//核心降级：如果熔断开启或报错
	if err != nil {
		handleLikeFallback(c, userID, vid)
		return
	}

	// --- 如果 Redis 成功，解析结果并继续你的逻辑 ---
	data := res.(map[string]interface{})
	action := data["action"].(string)
	finalLiked := data["finalLiked"].(bool)
	likeCount := data["likeCount"].(int64)

	//发送异步信息给MQ进行Mysql落盘
	//构造任务消息体
	task := map[string]interface{}{
		"task_type": action,
		"user_id":   userID,
		"video_id":  vid,
	}
	body, _ := json.Marshal(task)
	models.MQChannel.PublishWithContext(ctx,
		"",
		"video_tasks",
		false, false,
		amqp091.Publishing{
			ContentType: "application/json",
			Body:        body,
		},
	)
	go func() {
		//更新热榜
		models.RDB.ZAdd(ctx, "hot:rank", redis.Z{
			Score:  float64(likeCount),
			Member: vid,
		})
	}()

	//返回结果
	c.JSON(200, gin.H{
		"liked":      finalLiked,
		"like_count": likeCount,
	})
}
