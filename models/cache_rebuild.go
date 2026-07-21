package models

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
)

var lastRebuildTime time.Time
var rebuildMu sync.Mutex

// RebuildAllCaches Redis 恢复后重建 Feed 和热榜缓存
func RebuildAllCaches() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	//重建冷却时间
	rebuildMu.Lock()
	if time.Since(lastRebuildTime) < 2*time.Minute {
		rebuildMu.Unlock()
		return // 2分钟内重建过，跳过
	}
	lastRebuildTime = time.Now()
	rebuildMu.Unlock()

	// 1. Feed 缓存
	var feedIDs []uint
	DB.WithContext(ctx).Model(&Video{}).Order("id DESC").Limit(100).Pluck("id", &feedIDs)
	if len(feedIDs) > 0 {
		//先清空旧的 feed:zset
		pipe := RDB.Pipeline()
		pipe.Del(ctx, "feed:zset")
		for _, id := range feedIDs {
			pipe.ZAdd(ctx, "feed:zset", redis.Z{Score: float64(id), Member: id})
		}
		pipe.Expire(ctx, "feed:zset", 5*time.Minute)
		if _, err := pipe.Exec(ctx); err != nil {
			log.Printf("[恢复] Feed 缓存重建失败: %v", err)
		} else {
			log.Println("[恢复] Feed 缓存重建完成")
		}
	}

	// 2. 热榜缓存
	var topVideos []Video
	DB.WithContext(ctx).Order("like_count DESC").Limit(200).Find(&topVideos)
	if len(topVideos) > 0 {
		//先清空旧的 feed:zset
		pipe := RDB.Pipeline()
		pipe.Del(ctx, "hot:rank")
		for _, v := range topVideos {
			pipe.ZAdd(ctx, "hot:rank", redis.Z{Score: float64(v.LikeCount), Member: v.ID})
		}
		if _, err := pipe.Exec(ctx); err != nil {
			log.Printf("[恢复] 热榜缓存重建失败: %v", err)
		} else {
			log.Println("[恢复] 热榜缓存重建完成")
		}
	}
}
