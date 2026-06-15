package controllers

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"time"
	"video_feed/models"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"golang.org/x/sync/singleflight"
)

// 定义 singleflight.Group 用于防击穿
var feedGroup singleflight.Group

// 定义返回结构
type FeedVideo struct {
	models.Video
	Username string `json:"username"`
	IsLiked  bool   `json:"is_liked"`
}

func GetFeed(c *gin.Context) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	//获取userID
	userIDRaw, exist := c.Get("user_id")
	if !exist {
		c.JSON(401, gin.H{"error": "未登录"})
		return
	}
	userID := userIDRaw.(uint)

	//定义填充 IsLiked 是否点赞的辅助函数
	fillLiked := func(videos []FeedVideo) []FeedVideo {
		// 获取当前用户点赞过的视频ID集合
		var likedVideoIDs []uint
		models.DB.Model(&models.Like{}).Where("user_id = ?", userID).Pluck("video_id", &likedVideoIDs)
		//用于快速判断某个视频是否被当前用户点赞
		likeMap := make(map[uint]bool)
		for _, id := range likedVideoIDs {
			likeMap[id] = true
		}
		//遍历video取id,设置成true或者false
		for i := range videos {
			videos[i].IsLiked = likeMap[videos[i].ID]
		}
		return videos
	}

	//解析last_id
	lastIdStr := c.DefaultQuery("last_id", "0")
	last_Id, err := strconv.Atoi(lastIdStr)
	if err != nil {
		c.JSON(400, gin.H{"error": "参数 last_id 必须是整数"})
		return
	}
	//解析limit
	limit, err := strconv.Atoi(c.DefaultQuery("limit", "10"))
	if err != nil {
		c.JSON(400, gin.H{"error": "参数 limit 必须是整数"})
		return
	}
	//上下限检查
	if limit > 50 {
		limit = 50
	}
	if limit < 1 {
		limit = 10
	}
	//--------------------------------------------------------------------------------------------------------
	//定义了一个Redis 键名，用于存储视频 ID 的有序集合（ZSet）
	cachekey := "feed:zset"

	//声明变量 ids（字符串切片，存放从 Redis 获取的视频 ID）
	var videoIds []string

	// 从 Redis ZSet 获取视频 ID 列表(使用熔断器)
	redisFunc := func() (interface{}, error) {
		// 为 Redis 操作创建带超时的 context
		redisCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()

		//分页
		if last_Id == 0 {
			cmd := models.RDB.ZRevRange(redisCtx, cachekey, 0, int64(limit-1))
			return cmd.Result()
		} else {
			// 使用 ZRevRangeByScore，Max 设置为 lastId-1
			opt := &redis.ZRangeBy{
				Min: "-inf",
				Max: strconv.Itoa(last_Id - 1),
			}
			cmd := models.RDB.ZRevRangeByScore(redisCtx, cachekey, opt)
			//将结果存储到ids
			ids, err := cmd.Result()
			if err != nil {
				return nil, err
			}
			//判断limit
			if len(ids) > limit {
				ids = ids[:limit]
			}
			return ids, nil
		}
	}

	res, err := models.RedisCB.Execute(redisFunc)
	if err != nil {
		log.Printf("Redis 读取失败(熔断/错误):%v,降级到数据库", err)
		videoIds = nil //强制走数据库
	} else {
		videoIds = res.([]string)
	}

	//如果缓存命中且有数据
	if err == nil && len(videoIds) > 0 {
		var result []FeedVideo
		//查询数据库并存入result
		models.DB.WithContext(ctx).Table("videos").
			Select("videos.*,users.username").
			Joins("LEFT JOIN users ON users.id = videos.user_id").
			Where("videos.id IN ?", videoIds).
			Find(&result)

			//按照缓存videoIds顺序进行排序（为以后功能准备）
			// 如果有效视频数量小于 limit，说明缓存中有失效或不足的 ID
		if len(result) < limit || len(result) < len(videoIds) {
			//删除缓存也用熔断器保护
			go func() {
				redisCtx := context.Background()
				_, _ = models.RedisCB.Execute(func() (interface{}, error) {
					return nil, models.RDB.Del(redisCtx, cachekey).Err()
				})
			}()
		} else {
			//创建一个 map，键为视频 ID（uint 类型），值为 FeedVideo 结构体，用于快速查找。
			videoMap := make(map[uint]FeedVideo)
			//遍历 results，将每个视频存入 map，键为视频 ID
			for _, v := range result {
				videoMap[v.ID] = v
			}

			//声明一个新的 videos 切片，用于存放按正确顺序排列的结果。
			var videos []FeedVideo
			for _, idStr := range videoIds {
				id, _ := strconv.ParseUint(idStr, 10, 64)
				if v, ok := videoMap[uint(id)]; ok {
					videos = append(videos, v)
				}
			}
			videos = fillLiked(videos)
			c.Header("X-Cache", "HIT")
			c.JSON(200, gin.H{"list": videos})
			return
		}
	}

	//----------------------------------------------------------------------------------------------------------
	//如果缓存没命中，从数据库查询
	//热点防击穿
	//定义 key：用于区分不同的请求（不同的分页参数）
	key := fmt.Sprintf("feed:%d:%d", last_Id, limit)

	//gooup.DO方法
	result, err, _ := feedGroup.Do(key, func() (interface{}, error) {
		log.Println("执行数据库查询")
		//声明一个Video的结构体空切片，用于存储查询结果
		var videos []FeedVideo
		//一个构建好的查询构建器
		query := models.DB.WithContext(ctx).Table("videos").
			Select("videos.*,users.username").
			Joins("LEFT JOIN users ON users.id = videos.user_id").
			Order("videos.id DESC").
			Limit(limit)
		//分页
		if last_Id > 0 {
			query = query.Where("videos.id < ? ", last_Id)
		}
		//查询，存放到videos切片
		if err := query.Find(&videos).Error; err != nil {
			return nil, err
		}
		return videos, nil
	})
	if err != nil {
		c.JSON(500, gin.H{"error": "查询失败"})
		return
	}
	//将videos还原成切片
	videos, ok := result.([]FeedVideo)
	if !ok {
		c.JSON(500, gin.H{"error": "数据格式错误"})
		return
	}
	//从纯数据库查询到的 视频id 写入 Redis ZSet(使用熔断器)
	if last_Id == 0 && limit >= 5 && len(videos) > 0 {
		go func() {
			redisCtx := context.Background()
			var allIDs []uint
			models.DB.WithContext(context.Background()).Model(&models.Video{}).Order("id DESC").Limit(100).Pluck("id", &allIDs)
			_, _ = models.RedisCB.Execute(func() (interface{}, error) {
				//准备批量命令的容器
				pipe := models.RDB.Pipeline()
				for _, id := range allIDs {
					pipe.ZAdd(redisCtx, cachekey, redis.Z{Score: float64(id), Member: id})
				}
				pipe.Expire(redisCtx, cachekey, 5*time.Minute)
				if _, err := pipe.Exec(redisCtx); err != nil {
					log.Printf("Redis 缓存写入失败: %v", err)
				}
				return nil, nil
			})
		}()

	}
	videos = fillLiked(videos)
	c.Header("X-Cache", "MISS")
	c.JSON(200, gin.H{"list": videos})
}
