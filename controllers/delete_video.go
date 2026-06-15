package controllers

import (
	"log"
	"os"
	"strconv"
	"video_feed/models"

	"github.com/gin-gonic/gin"
)

func DeleteVideo(c *gin.Context) {
	//获取当前用户ID
	userIDRaw, exists := c.Get("user_id")
	if !exists {
		c.JSON(401, gin.H{"error": "未登录"})
		return
	}
	userID := userIDRaw.(uint)

	//获取视频ID
	videoIDParam := c.Param("id")
	videoID, err := strconv.ParseUint(videoIDParam, 10, 64)
	if err != nil {
		c.JSON(400, gin.H{"error": "无效的视频ID"})
		return
	}

	//查询视频
	var video models.Video
	if err := models.DB.First(&video, videoID).Error; err != nil {
		c.JSON(404, gin.H{"error": "视频不存在"})
		return
	}

	//权限检查:只能删除自己上传的视频
	if video.UserID != userID {
		c.JSON(403, gin.H{"error": "无权删除此视频"})
		return
	}

	//删除数据库记录
	if err := models.DB.Delete(&video).Error; err != nil {
		c.JSON(500, gin.H{"error": "删除失败"})
		return
	}

	//从Redis Zset 中移除视频ID
	ctx := c.Request.Context()
	cacheKey := "feed:zset"
	if err := models.RDB.Del(ctx, cacheKey).Err(); err != nil {
		log.Printf("删除缓存失败:%v", err)
	}
	c.JSON(200, gin.H{"message": "删除成功"})

	//删除服务器本地文件
	if err := os.Remove(video.FilePath); err != nil && !os.IsNotExist(err) {
		log.Printf("删除文件失败:%v", err)
	}

}
