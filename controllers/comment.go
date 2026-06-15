package controllers

import (
	"strconv"
	"video_feed/models"

	"github.com/gin-gonic/gin"
)

// 创建评论
func CreateComment(c *gin.Context) {
	//从上下文获取用户id
	userIDRaw, exists := c.Get("user_id")
	if !exists {
		c.JSON(401, gin.H{"error": "未登录"})
		return
	}
	userID := userIDRaw.(uint)

	//从URL路径参数获取视频id
	videoID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(400, gin.H{"error": "无效的视频ID"})
		return
	}

	var input struct {
		Content string `json:"content" binding:"required"`
	}
	//尝试将输入的 JSON 绑定到 input 变量
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(400, gin.H{"error": "评论内容不能为空"})
		return
	}

	//创建 models.Comment 结构体实例，填充 Content、UserID（当前用户）、VideoID
	comment := models.Comment{
		Content: input.Content,
		UserID:  userID,
		VideoID: uint(videoID),
	}
	//上传数据库
	if err := models.DB.Create(&comment).Error; err != nil {
		c.JSON(500, gin.H{"error": "发表评论失败"})
		return
	}

	//获取用户信息
	var user models.User
	models.DB.First(&user, userID)

	c.JSON(201, gin.H{
		"id":         comment.ID,
		"content":    comment.Content,
		"created_at": comment.CreatedAt,
		"user": gin.H{
			"id":       user.ID,
			"username": user.Username,
		},
	})
}

// 获取评论列表
func GetComments(c *gin.Context) {
	videoID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(400, gin.H{"error": "无效的视频ID"})
		return
	}

	//设定limit和lastId初始值
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "5"))
	if limit < 1 || limit > 50 {
		limit = 5
	}
	lastId, _ := strconv.Atoi(c.DefaultQuery("last_id", "0"))

	//用于存放评论的结构体类型
	type CommentWithUser struct {
		models.Comment
		Username string `json:"username"`
	}

	var comments []CommentWithUser
	//构建grom查询
	query := models.DB.Table("comments").
		Select("comments.*,users.username").
		Joins("LEFT JOIN users ON users.id = comments.user_id").
		Where("comments.video_id=?", videoID).
		Order("comments.created_at DESC").
		Limit(limit)

	//分页
	if lastId > 0 {
		query = query.Where("comments.id < ? ", lastId)
	}

	//执行查询，将结果填充到 comments 切片
	if err := query.Find(&comments).Error; err != nil {
		c.JSON(500, gin.H{"error": "获取评论失败"})
		return
	}

	var hasMore bool
	var nextLastId int = 0

	if len(comments) == limit {
		//取本页最后一个评论id
		nextLastId = int(comments[len(comments)-1].ID)
		var count int64
		models.DB.Model(&models.Comment{}).
			Where("video_id=? AND id<?", videoID, nextLastId).
			Count(&count)
		hasMore = count > 0
	}
	c.JSON(200, gin.H{
		"comments": comments,
		"has_more": hasMore,
		"last_id":  nextLastId,
	})
}

func DeleteComment(c *gin.Context) {
	//从上下文获取用户id
	userIDRaw, exists := c.Get("user_id")
	if !exists {
		c.JSON(401, gin.H{"error": "未登录"})
		return
	}
	userID := userIDRaw.(uint)

	commentID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(400, gin.H{"error": "无效的评论id"})
		return
	}

	//用于存放被查找的评论
	var comment models.Comment
	if err := models.DB.First(&comment, commentID).Error; err != nil {
		c.JSON(404, gin.H{"error": "评论不存在"})
		return
	}

	// 检查权限：评论作者或视频作者可以删除
	var video models.Video
	models.DB.First(&video, comment.VideoID)
	if comment.UserID != userID && video.UserID != userID {
		c.JSON(403, gin.H{"error": "无权删除此评论"})
		return
	}

	if err := models.DB.Delete(&comment).Error; err != nil {
		c.JSON(500, gin.H{"error": "删除失败"})
		return
	}

	c.JSON(200, gin.H{"message": "删除成功"})
}
