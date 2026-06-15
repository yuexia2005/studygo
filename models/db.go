package models

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/sony/gobreaker"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB
var RDB *redis.Client
var RedisCB *gobreaker.CircuitBreaker

// 连接数据库
func InitDB() {
	user := os.Getenv("DB_USER")
	if user == "" {
		user = "root"
	}

	password := os.Getenv("DB_PASSWORD")
	if password == "" {
		log.Fatal("环境变量 DB_PASSWORD 未设置")
	}

	host := os.Getenv("DB_HOST")
	if host == "" {
		host = "10.151.248.223"
	}

	post := os.Getenv("DB_POST")
	if post == "" {
		post = "3306"
	}

	dbname := os.Getenv("DB_NAME")
	if dbname == "" {
		dbname = "video_feed"
	}

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local", user, password, host, post, dbname)

	log.Printf("正在连接数据库")
	//启用grom的sql日志(测试热点防击穿)
	newLogger := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags),
		logger.Config{
			SlowThreshold: time.Second,
			LogLevel:      logger.Info, // 打印所有 SQL
			Colorful:      true,
		},
	)

	var err error
	DB, err = gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: newLogger,
	})

	if err != nil {
		panic("连接数据库失败: " + err.Error())
	}
	log.Println("数据库连接成功")

	//自动迁移
	DB.AutoMigrate(&User{}, &Video{}, &Like{}, &Comment{})
}

// 连接redis
func InitRedis() {
	addr := os.Getenv("REDIS_ADDR")
	if addr == "" {
		addr = "redis:6379"
	}
	RDB = redis.NewClient(&redis.Options{
		Addr: addr,
	})
	//测试连接
	ctx := context.Background()
	var err error
	for i := 0; i < 10; i++ {
		if _, err = RDB.Ping(ctx).Result(); err == nil {
			log.Println("Redis 连接成功")
			break
		}
		log.Printf("Redis 连接失败 (尝试 %d/10): %v", i+1, err)
		time.Sleep(2 * time.Second)
	}
	if err != nil {
		log.Printf("警告: Redis 最终连接失败: %v,服务继续运行但缓存功能不可用", err)
		// 不 panic，让服务启动，但后续熔断器会处理
	}

	//新增熔断器配置
	settings := gobreaker.Settings{
		Name:        "redis",
		MaxRequests: 1,               // 允许在半开状态下发送1个请求
		Interval:    5 * time.Second, // 重置计数器的时间间隔
		Timeout:     3 * time.Second, // 熔断状态持续时间
		ReadyToTrip: func(counts gobreaker.Counts) bool {
			// 简化触发条件：2个请求中有1个失败就熔断
			return counts.Requests >= 2 && float64(counts.TotalFailures)/float64(counts.Requests) >= 0.5
		},
	}
	RedisCB = gobreaker.NewCircuitBreaker(settings)

}
