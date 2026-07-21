package controllers

import (
	"context"
	"log"
	"strconv"
	"sync/atomic"
	"time"
	"video_feed/models"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"github.com/sony/gobreaker"
	"golang.org/x/sync/singleflight"
)

// 防击穿
var hotGroup singleflight.Group

type HotVideo struct {
	models.Video
	Username string `json:"username"`
	IsLiked  bool   `json:"is_liked"`
}

// 标志位：0 表示没有人在重建，1 表示正在重建中
var isRebuilding int32

// AsyncRebuildHotRank 异步重建热榜
func AsyncRebuildHotRank() {
	//先检查熔断器状态：如果处于 Open（开启）状态，说明 Redis 还没好，直接不干了
	// 只有在 Closed（正常）或 HalfOpen（尝试恢复）时才去重建
	if models.RedisCB.State() == gobreaker.StateOpen {
		log.Println("[Hot] Redis 处于熔断状态，取消异步重建")
		return
	}
	// 抢占锁：尝试将 isRebuilding 从 0 改为 1
	// 如果失败，说明已经有别的请求在重建了，直接退出
	if !atomic.CompareAndSwapInt32(&isRebuilding, 0, 1) {
		return
	}
	// 启动协程后台干活
	go func() {
		// 函数结束时，把标志位置回 0
		defer atomic.StoreInt32(&isRebuilding, 0)

		log.Println("[Redis] 触发异步重建热榜...")
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		// 重建,从数据库查找存入redis
		var topVideos []models.Video
		models.DB.WithContext(ctx).Order("like_count DESC").Limit(200).Find(&topVideos)
		if len(topVideos) > 0 {
			pipe := models.RDB.Pipeline()
			// 先删除旧的热榜 Key，防止旧数据干扰
			pipe.Del(ctx, "hot:rank")
			for _, v := range topVideos {
				pipe.ZAdd(ctx, "hot:rank", redis.Z{
					Score:  float64(v.LikeCount),
					Member: v.ID,
				})
			}
			_, err := pipe.Exec(ctx)
			if err != nil {
				log.Printf("[Redis] 重建热榜失败: %v", err)
			} else {
				log.Println("[Redis] 热榜异步重建完成")
			}
		}
	}()
}

// 用于判断此视频是否被点过赞
func fillLikedForHot(videos []HotVideo, userID uint) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	//查询点赞过的视频id存入切片
	var likedIDs []uint
	models.DB.WithContext(ctx).Model(&models.Like{}).Where("user_id = ?", userID).Pluck("video_id", &likedIDs)
	//构建 map，键为视频 ID，值为 true，用于 O(1) 查找
	likeMap := make(map[uint]bool)
	for _, id := range likedIDs {
		likeMap[id] = true
	}
	for i := range videos {
		videos[i].IsLiked = likeMap[videos[i].ID]
	}
}

func GetHot(c *gin.Context) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	userIDRaw, exists := c.Get("user_id")
	if !exists {
		c.JSON(401, gin.H{"error": "未登录"})
		return
	}
	userID := userIDRaw.(uint)

	//设置limit默认值
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "5"))
	if limit > 50 {
		limit = 50
	}

	// 用熔断器保护 Redis 读操作
	// 从 Redis 获取热度最高的前 limit 个视频 ID（按分数降序）
	result, err := models.RedisCB.Execute(func() (interface{}, error) {
		return models.RDB.ZRevRange(ctx, "hot:rank", 0, int64(limit-1)).Result()
	})
	var ids []string

	//redis出错或者结果为空
	if err != nil || result == nil {
		// 触发异步重建
		AsyncRebuildHotRank()
		// 热榜为空时，从数据库按点赞数降序查询
		result, err, _ := hotGroup.Do("hot_fallback", func() (interface{}, error) {
			var videos []HotVideo
			models.DB.WithContext(ctx).Table("videos").
				Select("videos.*,users.username").
				Joins("LEFT JOIN users ON users.id = videos.user_id").
				Order("videos.like_count DESC").
				Limit(limit).
				Find(&videos)
			return videos, nil
		})
		if err != nil { // ← 加错误处理
			c.JSON(500, gin.H{"error": "查询失败"})
			return
		}
		//判断一下是否点过赞
		videos := result.([]HotVideo)
		fillLikedForHot(videos, userID)
		c.JSON(200, gin.H{"list": videos})
		return
	}

	// 热榜为空（比如 key 刚被删），也降级
	ids = result.([]string)
	if len(ids) == 0 {
		AsyncRebuildHotRank()
		result, err, _ := hotGroup.Do("hot_fallback", func() (interface{}, error) {
			var videos []HotVideo
			models.DB.WithContext(ctx).Table("videos").
				Select("videos.*,users.username").
				Joins("LEFT JOIN users ON users.id = videos.user_id").
				Order("videos.like_count DESC").
				Limit(limit).
				Find(&videos)
			return videos, nil
		})
		if err != nil {
			c.JSON(500, gin.H{"error": "查询失败"})
			return
		}
		//判断一下是否点过赞
		videos := result.([]HotVideo)
		fillLikedForHot(videos, userID)
		c.JSON(200, gin.H{"list": videos})
		return
	}

	//从缓存获取，将ids转化成uint类型存入videoIDs
	var videoIDs []uint
	for _, idStr := range ids {
		id, _ := strconv.ParseUint(idStr, 10, 64)
		videoIDs = append(videoIDs, uint(id))
	}

	//查询视频详情
	var results []HotVideo
	models.DB.WithContext(ctx).Table("videos").
		Select("videos.*,users.username").
		Joins("LEFT JOIN users ON users.id = videos.user_id").
		Where("videos.id IN ?", videoIDs).
		Order("videos.like_count DESC").
		Find(&results)

	//按照redis返回的顺序重组
	videoMap := make(map[uint]HotVideo)
	for _, v := range results {
		videoMap[v.ID] = v
	}
	var videos []HotVideo
	for _, id := range videoIDs {
		if v, ok := videoMap[id]; ok {
			videos = append(videos, v)
		}
	}

	fillLikedForHot(videos, userID)
	c.JSON(200, gin.H{"list": videos})
}
