package controllers

import (
	"video_feed/models"
	"video_feed/utils"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

// 定义了一个名RegisterInput 的结构体，用于接收和验证用户注册请求的 JSON 数据
type RegisterInput struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required,min=6"`
}

// 将请求体的json数据绑定到input上
// 注册用户
func Register(c *gin.Context) {
	var input RegisterInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	// 检查用户名是否存在
	//用于存放查询到的用户（如果存在）
	var existUser models.User
	if models.DB.Where("username = ?", input.Username).First(&existUser).Error == nil {
		c.JSON(400, gin.H{"error": "用户名已存在"})
		return
	}
	//加密代码
	hashed, _ := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	user := models.User{
		Username: input.Username,
		Password: string(hashed),
	}
	if err := models.DB.Create(&user).Error; err != nil {
		c.JSON(500, gin.H{"error": "注册失败"})
		return
	}
	c.JSON(200, gin.H{"message": "注册成功"})
}

// 登陆
func Login(c *gin.Context) {
	// 将请求体的json数据绑定到input上
	var input RegisterInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	//判断用户名是否存在于数据库
	var user models.User
	if err := models.DB.Where("username = ?", input.Username).First(&user).Error; err != nil {
		c.JSON(401, gin.H{"error": "用户名或密码错误"})
		return
	}
	//判断密码是否正确
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(input.Password)); err != nil {
		c.JSON(401, gin.H{"error": "用户名或密码错误"})
		return
	}
	token, _ := utils.GenerateToken(user.ID)
	c.JSON(200, gin.H{
		"token":   token,
		"user_id": user.ID,
	})
}
