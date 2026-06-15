package controllers

import (
	"encoding/json"
	"log"
	"path/filepath"
	"video_feed/models"

	"github.com/gin-gonic/gin"
	"github.com/rabbitmq/amqp091-go"
	"github.com/redis/go-redis/v9"
)

func UploadVideo(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(401, gin.H{"error": "未登录"})
		return
	}
	//获取表单数据
	title := c.PostForm("title")
	description := c.PostForm(("description"))
	file, err := c.FormFile("video")
	if err != nil {
		c.JSON(400, gin.H{"error": "视频文件不能为空"})
		return
	}
	//保存文件到 uploads/ 目录
	//从上传文件的原始文件名中提取基础文件名（去除路径部分）
	filename := filepath.Base(file.Filename)
	//拼接目标保存路径
	dst := "./uploads/" + filename
	if err := c.SaveUploadedFile(file, dst); err != nil {
		c.JSON(500, gin.H{"error": "保存视频失败"})
		return
	}
	//定义vidio所拥有的初始属性
	video := models.Video{
		Title:       title,
		Description: description,
		FilePath:    dst,
		UserID:      userID.(uint),
	}
	//上传数据库
	if err := models.DB.Create(&video).Error; err != nil {
		c.JSON(500, gin.H{"error": "记录视频信息失败"})
		return
	}

	//上传新视频时，redis缓存更新
	ctx := c.Request.Context()

	//测试redis连接
	pipe := models.RDB.Pipeline()
	pipe.ZAdd(ctx, "feed:zset", redis.Z{Score: float64(video.ID), Member: video.ID})
	pipe.ZAdd(ctx, "hot:rank", redis.Z{Score: 0, Member: video.ID})
	if _, err := pipe.Exec(ctx); err != nil {
		log.Printf("Redis 缓存更新失败: %v", err)
	}

	// 发送 MQ 消息
	//构造并序列化任务体：
	taskBody, _ := json.Marshal(map[string]interface{}{
		"video_id":  video.ID,
		"file_path": video.FilePath,
		"user_id":   video.UserID,
	})

	//发布消息到 RabbitMQ
	err = models.MQChannel.Publish(
		"",
		"video_tasks",
		false,
		false,
		amqp091.Publishing{
			ContentType:  "application/json",
			Body:         taskBody,
			DeliveryMode: amqp091.Persistent,
		},
	)
	if err != nil {
		log.Printf("发送 MQ 消息失败: %v", err)
	}

	c.JSON(200, gin.H{"video_id": video.ID})

}
